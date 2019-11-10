package main

import (
    "gitlab.com/ocvt/api/app"
)

func main() {
  app.Initialize()
  app.Run(":3000")
}
