package utils

import (
  "os"
)

type Config struct {
  DBName string
  ClientUrl string
}

func GetConfig() *Config {
  var ClientUrlEnv string
  if os.Getenv("OCVT_CLIENT_URL") == "" {
    ClientUrlEnv = "https://ocvt.club"
  } else {
    ClientUrlEnv = os.Getenv("OCVT_CLIENT_URL")
  }

  return &Config{
    DBName: "ocvt-sqlite",
    ClientUrl: ClientUrlEnv,
  }
}
