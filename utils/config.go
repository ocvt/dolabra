package utils

import (
	"os"
)

type Config struct {
	DBName       string
	ApiUrl       string
	CookieDomain string
	FrontendUrl  string
}

func GetConfig() *Config {
	return &Config{
		DBName:       "dolabra-sqlite",
		ApiUrl:       os.Getenv("API_URL"),
		CookieDomain: os.Getenv("COOKIE_DOMAIN"),
		FrontendUrl:  os.Getenv("FRONTEND_URL"),
	}
}
