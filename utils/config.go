package utils

import (
	"os"
)

type Config struct {
	DBName string
	ApiUrl string
}

func GetConfig() *Config {
	return &Config{
		DBName: "dolabra-sqlite",
		ApiUrl: os.Getenv("API_URL"),
	}
}
