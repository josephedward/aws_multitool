package acloud

import (
	"aws-multitool/cli"
	"aws-multitool/core"
	"os"
	"fmt"
	"github.com/go-rod/rod"
)

func ACloudLogin(p ACloudProvider) (core.WebsiteLogin, error) {
	login := ReadWebsiteLoginFromEnv(p)
	fmt.Println("login : ", login)

	return login, nil
}

func ReadWebsiteLoginFromEnv(p ACloudProvider) core.WebsiteLogin {
	ACloudEnv, err := cli.LoadEnvPath("./.env.acloud")
	cli.PrintIfErr(err)
	p.ACloudEnv = ACloudEnv
	return core.WebsiteLogin{
		Url:      getEnv("URL", "https://learn.acloud.guru/cloud-playground/cloud-sandboxes"),
		Username: getEnv("USERNAME", ""),
		Password: getEnv("PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func ConnectBrowser(p ACloudProvider) (ACloudProvider, error) {
	p.Connection.Browser = rod.New().MustConnect()
	ACloudEnv, err := cli.LoadEnvPath("./.env.acloud")
	cli.PrintIfErr(err)
	p.ACloudEnv = ACloudEnv
	Connection := core.Connect(p.Connection.Browser, p.ACloudEnv.Url)
	cli.Success("Connection after: ", Connection)
	p.Connection = Connection
	return p, nil
}