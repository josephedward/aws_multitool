package acloud

import (
	"aws-multitool/core"
	"os"
	"fmt"
)

func ACloudLogin(p ACloudProvider) (core.WebsiteLogin, error) {
	login := ReadWebsiteLoginFromEnv()
	fmt.Println("login : ", login)

	return login, nil
}

func ReadWebsiteLoginFromEnv() core.WebsiteLogin {
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

