package main

import (
	"copy/internal/app"
	"flag"
	"log"
)

const defaultPort = "8080"

func main() {
	port := flag.String("port", defaultPort, "Port to listen on and connect to")
	flag.Parse()

	app, err := app.NewApp(*port)
	if err != nil {
		log.Fatalf("failed to create new app, %s", err)
	}
	app.Run()
}
