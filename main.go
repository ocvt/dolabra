package main

import (
	"github.com/ocvt/dolabra/app"
)

func main() {
	app.Initialize()
	app.Run(":3000")
}
