package main

import (
	"aws-multitool/acloud"
	"aws-multitool/cli"
	"aws-multitool/core"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog"
	// "runtime"
	// "log"
)

type AWSMaster struct {
	Profile    string
	AccessKey  string
	SecretKey  string
	Region     string
	AccountID  string
	OtherProps map[string]string
}

func main() {
	cli.Welcome()
	ZeroLog()
	sP := Profile()
	AwsLogin(sP)
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

func NewBrowser(url string) (core.Connection, error) {
	browser := rod.New().MustConnect()
	Connection := core.Connect(browser, url)
	cli.Success("Connection after: ", Connection)

	return Connection, nil
}

func Profile() (m *AWSMaster) {
	credentials, err := ReadAWSMasterFile()
	if err != nil {
		fmt.Println("Error reading AWS credentials:", err)
		return
	}

	var profileNames []string
	for _, creds := range credentials {
		profileNames = append(profileNames, creds.Profile)
	}

	prompt := promptui.Select{
		Label: "Select a profile : ",
		Items: profileNames,
	}

	_, selected, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return
	}

	if selected == "profile sandbox" {
		// run Sandbox method
		newCreds, err := Sandbox()
		if err != nil {
			fmt.Println("Error logging into ACloudGuru:", err)
			return
		}

		err = UpdateAWSCredentials("sandbox", newCreds.SandboxCredential.KeyID, newCreds.SandboxCredential.AccessKey)

		if err != nil {
			fmt.Println("Error updating AWS credentials:", err)
			return
		}
		fmt.Println("Sandbox credentials updated successfully!")
	}

	return &AWSMaster{
		Profile:   selected,
		AccessKey: credentials[0].AccessKey,
		SecretKey: credentials[0].SecretKey,
		Region:    credentials[0].Region,
	}
}

func Sandbox() (acloud.ACloudProvider, error) {
	var p acloud.ACloudProvider
	login, err := acloud.ACloudLogin(p)
	// p.Connection = connect
	p.ACloudEnv.Url = login.Url
	p.ACloudEnv.Username = login.Username
	p.ACloudEnv.Password = login.Password
	if err != nil {
		fmt.Println("Error logging into ACloudGuru:", err)
		return p, err
	}

	p, err = acloud.ConnectBrowser(p)
	cli.PrintIfErr(err)
	cli.Success("environment : ", p.ACloudEnv)

	// log browser
	cli.Success("Browser : ", p.Connection.Browser)

	// //login to acloud
	p.Connection, err = core.Login(core.WebsiteLogin{p.ACloudEnv.Url, p.ACloudEnv.Username, p.ACloudEnv.Password}, p.Connection.Browser)
	cli.PrintIfErr(err)
	cli.Success("A Cloud Provider : ", p)

	time.Sleep(1 * time.Second)

	//scrape credentials
	elems, err := acloud.Sandbox(p.Connection, p.ACloudEnv.Download_key)
	cli.PrintIfErr(err)
	cli.Success("rod html elements : ", elems)

	//copy credentials to clipboard
	creds, err := acloud.SimpleCopy(elems)
	cli.PrintIfErr(err)
	acloud.DisplayCreds(creds)

	//save provider
	p.SandboxCredential = creds
	return p, err
}

func ReadAWSMasterFile() ([]AWSMaster, error) {
	var credentials []AWSMaster

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
	var currentCreds AWSMaster

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// New profile
			if currentProfile != "" {
				credentials = append(credentials, currentCreds)
			}
			currentProfile = line[1 : len(line)-1]
			currentCreds = AWSMaster{Profile: currentProfile}
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

func UpdateAWSCredentials(profile, keyID, accessKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	credentialsFile := filepath.Join(homeDir, ".aws", "credentials")
	tempCredentialsFile := credentialsFile + ".tmp"

	credentials, err := os.Open(credentialsFile)
	if err != nil {
		return err
	}
	defer credentials.Close()

	tempFile, err := os.Create(tempCredentialsFile)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	writer := bufio.NewWriter(tempFile)

	scanner := bufio.NewScanner(credentials)
	insideTargetProfile := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if insideTargetProfile {
				insideTargetProfile = false
			}

			profileName := strings.TrimSpace(line[1 : len(line)-1])
			if profileName == profile {
				insideTargetProfile = true
				_, _ = fmt.Fprintf(writer, "[%s]\n", profile)
				_, _ = fmt.Fprintf(writer, "aws_access_key_id = %s\n", keyID)
				_, _ = fmt.Fprintf(writer, "aws_secret_access_key = %s\n", accessKey)
				_, _ = fmt.Fprintln(writer)
			} else {
				_, _ = fmt.Fprintln(writer, line)
			}
		} else if !insideTargetProfile {
			_, _ = fmt.Fprintln(writer, line)
		}
	}

	if !insideTargetProfile {
		// Profile not found, create a new entry
		_, _ = fmt.Fprintf(writer, "[%s]\n", profile)
		_, _ = fmt.Fprintf(writer, "aws_access_key_id = %s\n", keyID)
		_, _ = fmt.Fprintf(writer, "aws_secret_access_key = %s\n", accessKey)
		_, _ = fmt.Fprintln(writer)
	}

	writer.Flush()

	err = os.Rename(tempCredentialsFile, credentialsFile)
	if err != nil {
		return err
	}

	return nil
}

func OpenAWSConsole(selectedProfile *AWSMaster) {
	var openConsole string
	fmt.Print("Do you want to open the AWS Management Console? (y/n): ")
	fmt.Scan(&openConsole)

	if openConsole == "n" || openConsole == "no" || openConsole == "N" || openConsole == "No" || openConsole == "NO" {
		fmt.Println("AWS Management Console will not be opened.")
		return
	}

	if selectedProfile == nil || selectedProfile.Region == "" {
		fmt.Println("Please select a valid AWS profile with a specified region.")
		return
	}

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

	fmt.Println("Navigating to AWS Management Console page...")

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(consoleURL)
	defer page.MustClose()

	// Optionally, you can take further actions using rod to interact with the page if needed.
	// For example, you might want to log in with credentials or other interactions.

	fmt.Println("AWS Management Console page navigated to.")
	select {} // This line will prevent the program from exiting and keep the browser open.
}

func AwsLogin(selectedProfile *AWSMaster) {

	// Set the AWS environment variables for the selected profile
	os.Setenv("AWS_ACCESS_KEY_ID", selectedProfile.AccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", selectedProfile.SecretKey)
	os.Setenv("AWS_REGION", selectedProfile.Region)

	fmt.Println("AWS profile set to:", selectedProfile.Profile)

	// Open AWS Management Console with the selected profile
	OpenAWSConsole(selectedProfile)

}
