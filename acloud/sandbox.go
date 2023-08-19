package acloud

import (
	"aws-multitool/cli"
	"aws-multitool/core"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"time"
)

type ACloudProvider struct {
	cli.ACloudEnv
	core.Connection
	SandboxCredential
	// *SQLiteRepository
}

type SandboxCredential struct {
	ID        int64
	User      string
	Password  string
	URL       string
	KeyID     string
	AccessKey string
}

func Sandbox(connect core.Connection, downloadKey string) (rod.Elements, error) {

	elems := make(rod.Elements, 0)
	// It will keep polling until one selector has found a match
	connect.Page.Race().ElementR("button", "Start AWS Sandbox").MustHandle(func(e *rod.Element) {
		e.MustClick()
		// time.Sleep(1 * time.Second)
		elems = Scrape(connect)
	}).Element("div[role='tabpanel']").MustHandle(func(e *rod.Element) {
		// time.Sleep(1 * time.Second)
		elems = Scrape(connect)
	}).MustDo()

	time.Sleep(1 * time.Second)

	//log the page
	connect.Page.MustScreenshot("sandbox.png")

	if len(elems) == 0 {
		return nil, errors.New("no elements found")
	}
	return elems, nil
}

func SimpleCopy(elems rod.Elements) (SandboxCredential, error) {
	return SandboxCredential{
		User:      elems[0].MustText(),
		Password:  elems[1].MustText(),
		URL:       elems[2].MustText(),
		KeyID:     elems[3].MustText(),
		AccessKey: elems[4].MustText(),
	}, nil
}

func Scrape(connect core.Connection) rod.Elements {

	elems := connect.Page.MustWaitLoad().MustElements("input[aria-label='Copy to clipboard']")
	return elems
}

func KeyVals(creds SandboxCredential) ([]string, []string) {
	keys := []string{"username", "password", "url", "keyid", "accesskey"}
	vals := []string{string(creds.User),
		string(creds.Password),
		string(creds.URL),
		string(creds.KeyID),
		string(creds.AccessKey)}

	return keys, vals
}

func DisplayCreds(creds SandboxCredential) {
	//if creds are empty, throw message and return
	if creds.User == "" {
		cli.Error("Warning: No Credentials Found")
		return
	}

	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Println("Sandbox Credentials: ")
	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Println("          " + cli.Cyan + "Username: " + cli.Yellow + creds.User + cli.Reset)
	fmt.Println("          " + cli.Cyan + "Password: " + cli.Yellow + creds.Password + cli.Reset)
	fmt.Println("          " + cli.Cyan + "URL: " + cli.Yellow + creds.URL + cli.Reset)
	fmt.Println("          " + cli.Cyan + "KeyID: " + cli.Yellow + creds.KeyID + cli.Reset)
	fmt.Println("          " + cli.Cyan + "AccessKey: " + cli.Yellow + creds.AccessKey + cli.Reset)
	fmt.Println("-----------------------------------------------------------------------------------")
}

