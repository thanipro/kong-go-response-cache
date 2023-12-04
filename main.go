package main

import (
	"github.com/Kong/go-pdk/server"
	"log"
)

const (
	Version  = "0.1.0"
	Priority = 1000
)

func New() interface{} {
	return &Config{}
}

func main() {
	err := server.StartServer(New, Version, Priority)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
