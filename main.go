package main

import (
	"aws-multitool/acloud"
	"aws-multitool/cli"
	"aws-multitool/core"
	"bufio"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	// "database/sql"
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

	for {

		// Ask the user if they want to switch AWS profile or open the console
		pSwitch := promptSwitch()

		if pSwitch == "Switch Profile" {
			// If the user chooses to switch profile, call the Profile function
			profile()
		} else if pSwitch == "Open AWS Console" {
			// If the user chooses to open the AWS console, call the awsConsole function
			awsConsole()
		} else if pSwitch == "Set Credentials" {
			// If the user chooses to set credentials, call the retrieveCredentials function
			setCreds("", "", "", "")
		} else if pSwitch == "Exit" {
			break
		}
	}

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

func promptSwitch() string {
	prompt := promptui.Select{
		Label: "Choose an option",
		Items: []string{"Switch Profile", "Open AWS Console", "Set Credentials", "Exit"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func profile() (m *AWSMaster) {
	credentials, err := readAWSMasterFile()
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
		newCreds, err := sandbox()
		if err != nil {
			fmt.Println("Error logging into ACloudGuru:", err)
			return
		}

		err = replaceProfileCredentials("sandbox", newCreds.SandboxCredential.KeyID, newCreds.SandboxCredential.AccessKey)

		if err != nil {
			fmt.Println("Error updating AWS credentials:", err)
			return
		}
		fmt.Println("Sandbox credentials updated successfully!")
	}

	//set environment for $AWS_PROFILE
	os.Setenv("AWS_PROFILE", selected)

	return &AWSMaster{
		Profile:   selected,
		AccessKey: credentials[0].AccessKey,
		SecretKey: credentials[0].SecretKey,
		Region:    credentials[0].Region,
	}
}

func sandbox() (acloud.ACloudProvider, error) {
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

func readAWSMasterFile() ([]AWSMaster, error) {
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

func replaceProfileCredentials(profileName, awsAccessKeyID, awsSecretAccessKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	credentialsFile := filepath.Join(homeDir, ".aws", "credentials")

	// Open the credentials file in read mode
	file, err := os.Open(credentialsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a temporary file to write the modified contents
	tmpFile, err := os.CreateTemp("", "aws_credentials")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	scanner := bufio.NewScanner(file)
	inProfile := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "["+profileName+"]") {
			inProfile = true
		} else if inProfile && strings.HasPrefix(line, "[") {
			inProfile = false
		}

		if inProfile {
			if strings.HasPrefix(line, "aws_access_key_id") {
				line = fmt.Sprintf("aws_access_key_id = %s", awsAccessKeyID)
			} else if strings.HasPrefix(line, "aws_secret_access_key") {
				line = fmt.Sprintf("aws_secret_access_key = %s", awsSecretAccessKey)
			}
		}

		_, err = fmt.Fprintln(tmpFile, line)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Replace the original credentials file with the temporary file
	err = os.Rename(tmpFile.Name(), credentialsFile)
	if err != nil {
		return err
	}

	return nil
}

func getAwsConsoleUrl() (consoleURL string) {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting AWS account ID:", err)
		return
	}

	accountIdSlice := strings.TrimSpace(string(output))
	if accountIdSlice == "" {
		fmt.Println("Failed to get the AWS account ID.")
		return
	}
	// print account id
	fmt.Println("Retrieved Account ID : ", accountIdSlice)

	accountID := string(accountIdSlice)
	consoleURL = fmt.Sprintf("https://%s.signin.aws.amazon.com/console", accountID)
	return consoleURL
}

func awsConsole() {

	// print the current env var for AWS_PROFILE
	fmt.Println("AWS_PROFILE : ", os.Getenv("AWS_PROFILE"))

	consoleURL := getAwsConsoleUrl()

	fmt.Println("Navigating to AWS Management Console page..." + consoleURL)

	//set credentials
	// setCreds(os.Getenv("AWS_PROFILE"), consoleURL, "", "")

	u := launcher.New().Bin("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome").Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	// //login to acloud
	connection, err := core.Login(core.WebsiteLogin{consoleURL, "", ""}, browser)
	cli.PrintIfErr(err)
	cli.Success("connection : ", connection)

}
