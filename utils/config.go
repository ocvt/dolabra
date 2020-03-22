package utils

import (
	"os"
)

type Config struct {
	DBName    string
	ApiUrl    string
	ClientUrl string
}

func GetConfig() *Config {
	return &Config{
		DBName:    "dolabra-sqlite",
		ApiUrl:    os.Getenv("API_URL"),
		ClientUrl: os.Getenv("CLIENT_URL"),
	}
}
