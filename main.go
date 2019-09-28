package main

import (
    "gitlab.com/ocvt/api/app"
    "gitlab.com/ocvt/api/config"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	app.Run(":3000")
}
