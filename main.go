package main

import (
	"os"

	"xenotification/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "7000"
	}
	app.Start(port)
}
