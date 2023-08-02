package main

import (
	"aws-multitool/acloud"
	"aws-multitool/cli"
	"aws-multitool/core"
	"bufio"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	// "runtime"
)

type AWSCredentials struct {
	Profile    string
	AccessKey  string
	SecretKey  string
	Region     string
	AccountID  string
	OtherProps map[string]string
}

func ReadAWSCredentialsFile() ([]AWSCredentials, error) {
	var credentials []AWSCredentials

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(homeDir, ".aws", "config")
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentProfile := ""
	var currentCreds AWSCredentials

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// New profile
			if currentProfile != "" {
				credentials = append(credentials, currentCreds)
			}
			currentProfile = line[1 : len(line)-1]
			currentCreds = AWSCredentials{Profile: currentProfile}
		} else if strings.Contains(line, "=") {
			// Key-value pair
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "aws_access_key_id":
				currentCreds.AccessKey = value
			case "aws_secret_access_key":
				currentCreds.SecretKey = value
			case "region":
				currentCreds.Region = value
			default:
				if currentCreds.OtherProps == nil {
					currentCreds.OtherProps = make(map[string]string)
				}
				currentCreds.OtherProps[key] = value
			}
		}
	}

	// Append the last profile
	if currentProfile != "" {
		credentials = append(credentials, currentCreds)
	}

	return credentials, scanner.Err()
}

func SelectAWSProfile(credentials []AWSCredentials) *AWSCredentials {
	fmt.Println("Available AWS Profiles:")
	for _, cred := range credentials {
		fmt.Println("-", cred.Profile)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name of the profile you want to use: ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)

	for _, cred := range credentials {
		if cred.Profile == profileName {
			return &cred
		}
	}

	return nil
}

func OpenAWSConsole(selectedProfile *AWSCredentials) {
	if selectedProfile == nil || selectedProfile.Region == "" {
		fmt.Println("Please select a valid AWS profile with a specified region.")
		return
	}

	// Get the account number via AWS CLI
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting AWS account ID:", err)
		return
	}

	accountID := strings.TrimSpace(string(output))
	if accountID == "" {
		fmt.Println("Failed to get the AWS account ID.")
		return
	}

	selectedProfile.AccountID = accountID

	consoleURL := fmt.Sprintf("https://%s.signin.aws.amazon.com/console", selectedProfile.AccountID)

	cmd = exec.Command("open", consoleURL)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error opening AWS Management Console:", err)
		return
	}

	fmt.Println("AWS Management Console opened in your default web browser.")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func ReadWebsiteLoginFromEnv() core.WebsiteLogin {
	return core.WebsiteLogin{
		Url:      getEnv("URL", "https://learn.acloud.guru/cloud-playground/cloud-sandboxes"),
		Username: getEnv("USERNAME", ""),
		Password: getEnv("PASSWORD", ""),
	}
}

func ACloudLogin(p acloud.ACloudProvider) (core.WebsiteLogin, error) {
	login := ReadWebsiteLoginFromEnv()
	fmt.Println("login : ", login)

	return login, nil
}

func AwsLogin() {
	credentials, err := ReadAWSCredentialsFile()
	if err != nil {
		fmt.Println("Error reading AWS credentials file:", err)
		return
	}

	selectedProfile := SelectAWSProfile(credentials)
	if selectedProfile == nil {
		fmt.Println("Profile not found.")
		return
	}

	// Set the AWS environment variables for the selected profile
	os.Setenv("AWS_ACCESS_KEY_ID", selectedProfile.AccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", selectedProfile.SecretKey)
	os.Setenv("AWS_REGION", selectedProfile.Region)

	fmt.Println("AWS profile set to:", selectedProfile.Profile)

	// Open AWS Management Console with the selected profile
	OpenAWSConsole(selectedProfile)
}

func OpenDefaultBrowserAndNavigate(pageURL string) (*rod.Page, error) {
	browserInstance := rod.New().MustConnect()
	// Create a new Rod page
	page := browserInstance.MustPage(pageURL)

	// Navigate the page to the specified URL
	err := page.Navigate(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to the page: %w", err)
	}

	page.WaitLoad()

	return page, nil
}

func ZeroLog() {
	fmt.Println("os.Args : ", os.Args)
	// default
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	//if string prod is in args, set global level to info
	for _, arg := range os.Args {
		if arg == "prod" {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	}
	fmt.Println("global logger level : ")
	fmt.Println(zerolog.GlobalLevel())
}

func ConnectBrowser(p acloud.ACloudProvider) (acloud.ACloudProvider, error) {
	p.Connection.Browser = rod.New().MustConnect()
	ACloudEnv, err := cli.LoadEnv()
	cli.PrintIfErr(err)
	p.ACloudEnv = ACloudEnv
	Connection := core.Connect(p.Connection.Browser, p.ACloudEnv.Url)
	cli.Success("Connection after: ", Connection)
	p.Connection = Connection
	return p, nil
}

func Sandbox(p *acloud.ACloudProvider) (acloud.ACloudProvider, error) {

	cli.Success("p.Connection.Browser : ", p.Connection)
	cli.Success("p.ACloudEnv.Download_key : ", p.ACloudEnv.Download_key)

	//scrape credentials
	elems, err := acloud.Sandbox(p.Connection, p.ACloudEnv.Download_key)
	cli.PrintIfErr(err)
	cli.Success("rod html elements : ", elems)

	// copy credentials to clipboard
	p.SandboxCredential, err = acloud.CopySvg(elems)
	cli.PrintIfErr(err)
	cli.Success("credentials : ", p.SandboxCredential)

	acloud.DisplayCreds(p.SandboxCredential)

	return *p, err
}

func main() {
	cli.Welcome()
	ZeroLog()

	var p acloud.ACloudProvider
	login, err := ACloudLogin(p)
	// p.Connection = connect
	p.ACloudEnv.Url = login.Url
	p.ACloudEnv.Username = login.Username
	p.ACloudEnv.Password = login.Password
	if err != nil {
		fmt.Println("Error logging into ACloudGuru:", err)
		return
	}

	p, err = ConnectBrowser(p)
	// log browser
	cli.Success("Browser : ", p.Connection.Browser)
	cli.PrintIfErr(err)
	cli.Success("environment : ", p.ACloudEnv)

	//login to acloud
	p.Connection, err = core.Login(core.WebsiteLogin{p.ACloudEnv.Url, p.ACloudEnv.Username, p.ACloudEnv.Password})
	cli.PrintIfErr(err)
	// cli.Success("A Cloud Provider : ", p)

	//get sandbox credentials
	p, err = Sandbox(&p)
	cli.PrintIfErr(err)
	// cli.Success("p : ", p)

}
