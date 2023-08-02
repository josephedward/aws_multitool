package iq

import (
	"fmt"
	"aws-multitool/cli"
	"aws-multitool/core"

	"github.com/go-rod/rod"
	// "strconv"
)

// type IqProxy interface {
// 	CreateProvider(browserString string) IqProvider
// 	Login(connect core.Connection, login core.WebsiteLogin)
// 	GetRequests(connect core.Connection) []IqRequest //defaults to 20 requests
// 	GetBatchRequests(connect core.Connection, number int) []IqRequest
// }

type IqProvider struct {
	// LoggedIn bool
	core.AwsEnv
	core.Connection
	Requests []IqRequest
}

type IqRequest struct {
	Element *rod.Element
	Title   string
	Date    string
	Author  string
	Content string
	Budget  string
	Label   string
}

func GetRequests(connect core.Connection) (reqs []IqRequest) {
	//scrape website
	elems := GetTitles(connect)
	fmt.Println("elems : ", elems)

	//get the content of each request
	for _, elem := range elems {
		content := GetContent(connect, elem)
		author := GetAuthor(connect, elem)

		reqs = append(reqs, IqRequest{
			Element: elem,
			Title:   elem.MustText(),
			Content: content,
			Author:  author,
		})
	}
	return reqs
}

func GetBatchRequests(connect core.Connection, number int) (reqs []IqRequest) {
	reqs = []IqRequest{}
	for {
		// do something
		if len(reqs) < number {
			reqs = append(reqs, GetElements(connect)...)
		} else if len(reqs) == number {
			break
		} else {
			reqs = reqs[:number]
			break
		}
	}
	return reqs
}

func GetElements(connect core.Connection) (reqs []IqRequest) {

	// log the number of elements
	fmt.Println("Number of elements : ", len(connect.Page.MustWaitLoad().MustElements("div[class^='content ProjectRailItem']")))
	//log number of titles, dates, labels, budgets
	fmt.Println("Number of titles : ", len(connect.Page.MustWaitLoad().MustElements("div[class^='ProjectRailItem__title']")))
	fmt.Println("Number of dates : ", len(connect.Page.MustWaitLoad().MustElements("small[class^='ProjectRailItem__date']")))
	fmt.Println("Number of labels : ", len(connect.Page.MustWaitLoad().MustElements("div[class^='ProjectRailItem__label']")))
	fmt.Println("Number of budgets : ", len(connect.Page.MustWaitLoad().MustElements("div[class^='ProjectRailItem__budget']")))

	//get whole element
	elems := connect.Page.MustWaitLoad().MustElements("div[class^='content ProjectRailItem']")
	titles := connect.Page.MustWaitLoad().MustElements("div[class^='ProjectRailItem__title']")
	dates := connect.Page.MustWaitLoad().MustElements("small[class^='ProjectRailItem__date']")

	//loop over length of elems
	for i := 0; i < len(titles); i++ {
		var tempElem *rod.Element

		//declare temporary vars to hold values
		var tempTitle string
		var tempContent string
		var tempAuthor string
		var tempBudget string
		var tempLabel string
		var tempDate string

		tempBudget = "0"
		tempLabel = "None"

		//look at the element for a budget and label
		if len(elems[i].MustElements("div[class^='ProjectRailItem__budget']")) > 0 {
			tempBudget = elems[i].MustElement("div[class^='ProjectRailItem__budget']").MustText()
		}
		if len(elems[i].MustElements("div[class^='ProjectRailItem__label']")) > 0 {
			tempLabel = elems[i].MustElement("div[class^='ProjectRailItem__label']").MustText()
		}

		//check each element for nil
		if error := elems[i]; error != nil {
			tempElem = elems[i]
		}
		if error := titles[i]; error != nil {
			tempTitle = titles[i].MustText()
		}
		if error := dates[i]; error != nil {
			tempDate = dates[i].MustText()
		}

		//get content and author, use error handling to check for nil
		if error, _ := GetRequestDetails(connect, titles[i]); error != "" {
			tempContent, tempAuthor = GetRequestDetails(connect, titles[i])
		} else {
			continue
		}

		reqs = append(reqs, IqRequest{
			Element: tempElem,
			Title:   tempTitle,
			Date:    tempDate,
			Author:  tempAuthor,
			Content: tempContent,
			Budget:  tempBudget,
			Label:   tempLabel,
		})
	}
	return reqs
}

func GetRequestDetails(connect core.Connection, elem *rod.Element) (content string, author string) {
	//get the title, content, author, budget, label, and header
	elem.MustClick()
	content = connect.Page.MustWaitLoad().MustElement("div[class^='TextMessage__text']").MustText()
	author = connect.Page.MustWaitLoad().MustElement("span[class^='author']").MustText()

	return content, author
}

func GetTitles(connect core.Connection) rod.Elements {
	fmt.Println("connect : ", connect)
	elems := connect.Page.MustWaitLoad().MustElements("div[class^='ProjectRailItem__title']")
	fmt.Println("elems : ", elems)
	return elems
}

func GetContent(connect core.Connection, elem *rod.Element) string {
	elem.MustClick()
	temp := connect.Page.MustWaitLoad().MustElement("div[class^='TextMessage__text']").MustText()
	return temp
}

func GetAuthor(connect core.Connection, elem *rod.Element) string {
	elem.MustClick()
	temp := connect.Page.MustWaitLoad().MustElement("span[class^='author']").MustText()
	return temp
}

func InsertMessage(connect core.Connection, elem *rod.Element, message string) {
	elem.MustClick()
	textarea := connect.Page.MustWaitLoad().MustElement("textarea[id^='initialResponse']")
	textarea.MustSelectAllText().MustInput("")
	textarea.MustInput(message)
	cli.Success("Message inserted")
}

func DraftIntro() {}

func AnswerAllRequests(connect core.Connection, reqs []IqRequest) {
	for _, req := range reqs {
		InsertMessage(connect, req.Element, "Hello, I am a bot. I am responding to your request.")
	}
}

func DisplayRequests(reqs []IqRequest) {
	for _, req := range reqs {
		fmt.Println("Title : ")
		cli.Success(req.Title)
		fmt.Println("Date : ")
		cli.Success(req.Date)
		fmt.Println("Label : ")
		cli.Success(req.Label)
		fmt.Println("Budget : ")
		cli.Success(req.Budget)
		fmt.Println("Content : ")
		cli.Success(req.Content)
		fmt.Println("Author : ")
		cli.Success(req.Author)
		fmt.Println()
	}
}
