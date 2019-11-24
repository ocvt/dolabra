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
  if os.Getenv("DOLABRA_CLIENT_URL") == "" {
    ClientUrlEnv = "http://localhost:1313"
  } else {
    ClientUrlEnv = os.Getenv("DOLABRA_CLIENT_URL")
  }

  return &Config{
    DBName: "dolabra-sqlite",
    ClientUrl: ClientUrlEnv,
  }
}
