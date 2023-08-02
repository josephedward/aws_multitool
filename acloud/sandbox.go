package acloud

import (
	"aws-multitool/cli"
	"aws-multitool/core"
	"errors"
	"fmt"

	// "os"
	"time"

	"github.com/go-rod/rod"
	"golang.design/x/clipboard"
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
		time.Sleep(1 * time.Second)
		// elems = Scrape(connect)
	}).Element("div[role='tabpanel']").MustHandle(func(e *rod.Element) {
		time.Sleep(1 * time.Second)
		// elems = Scrape(connect)
	}).MustDo()

	time.Sleep(1 * time.Second)
	
	//log the page
	connect.Page.MustScreenshot("sandbox.png")

	// if len(elems) == 0 {
	// 	return nil, errors.New("no elements found")
	// }
	return elems, nil
}

func Scrape(connect core.Connection) rod.Elements {

	elems := connect.Page.MustWaitLoad().MustElements("svg[aria-label='copy icon']")
	return elems
}

func Copy(elems rod.Elements) (SandboxCredential, error) {
	//initialize cliboard package
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	//have to copy to clipboard to get whole string
	elems[0].MustElement("svg[aria-label='copy icon']").MustClick()
	// write/read text format data of the clipboard, and
	// the byte buffer regarding the text are UTF8 encoded.
	un := clipboard.Read(clipboard.FmtText)
	//zero out the clipboard just in case
	clipboard.Write(clipboard.FmtText, nil)

	elems[1].MustElement("svg[aria-label='copy icon']").MustClick()

	pw := clipboard.Read(clipboard.FmtText)
	fmt.Println("pw : ", pw)

	clipboard.Write(clipboard.FmtText, nil)

	elems[2].MustElement("svg[aria-label='copy icon']").MustClick()
	url := clipboard.Read(clipboard.FmtText)
	fmt.Println("url : ", url)
	clipboard.Write(clipboard.FmtText, nil)

	elems[3].MustElement("svg[aria-label='copy icon']").MustClick()
	keyid := clipboard.Read(clipboard.FmtText)
	fmt.Println("keyid : ", keyid)
	clipboard.Write(clipboard.FmtText, nil)

	elems[4].MustElement("svg[aria-label='copy icon']").MustClick()
	accesskey := clipboard.Read(clipboard.FmtText)
	fmt.Println("accesskey : ", accesskey)
	clipboard.Write(clipboard.FmtText, nil)

	return SandboxCredential{
		User:      string(un),
		Password:  string(pw),
		URL:       string(url),
		KeyID:     string(keyid),
		AccessKey: string(accesskey),
	}, nil
}

func CopySvg(elems rod.Elements) (SandboxCredential, error) {
	// initialize cliboard package
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	un := ClipBoard(elems[0])
	fmt.Println("un : ", un)
	pw := ClipBoard(elems[1])
	fmt.Println("pw : ", pw)
	url := ClipBoard(elems[2])
	fmt.Println("url : ", url)
	keyid := ClipBoard(elems[3])
	fmt.Println("keyid : ", keyid)
	accesskey := ClipBoard(elems[4])
	fmt.Println("accesskey : ", accesskey)

	cli.Success("Credentials Copied")

	return SandboxCredential{
		User:      un,
		Password:  pw,
		URL:       url,
		KeyID:     keyid,
		AccessKey: accesskey,
	}, nil
}

// have to copy to clipboard to get whole string
func ClipBoard(elem *rod.Element) string {
	fmt.Println("elem : ", elem)
	time.Sleep(1 * time.Second)
	elem.MustClick()
	res := clipboard.Read(clipboard.FmtText)
	cli.Success("clipboard val : ", string(res))

	//zero out the clipboard
	clipboard.Write(clipboard.FmtText, nil)
	return string(res)
}

// func CopyHtml(elems rod.Elements) (SandboxCredential, error) {

// 	elems[0].MustElement("svg[aria-label='copy icon']").MustClick()
// 	un := elems[0].MustElement("input").MustProperty("value").String()
// 	fmt.Println("un : ", un)

// 	elems[1].MustElement("svg[aria-label='copy icon']").MustClick()
// 	pw := elems[1].MustElement("input").MustProperty("value").String()
// 	fmt.Println("pw : ", pw)
// 	elems[2].MustElement("svg[aria-label='copy icon']").MustClick()
// 	url := elems[2].MustElement("input").MustProperty("value").String()
// 	fmt.Println("url : ", url)
// 	elems[3].MustElement("svg[aria-label='copy icon']").MustClick()
// 	keyid := elems[3].MustElement("input").MustProperty("value").String()
// 	fmt.Println("keyid : ", keyid)
// 	elems[4].MustElement("svg[aria-label='copy icon']").MustClick()
// 	accesskey := elems[4].MustElement("input").MustProperty("value").String()
// 	fmt.Println("accesskey : ", accesskey)

// 	return SandboxCredential{
// 		User:      string(un),
// 		Password:  string(pw),
// 		URL:       string(url),
// 		KeyID:     string(keyid),
// 		AccessKey: string(accesskey),
// 	}, nil

// }

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
