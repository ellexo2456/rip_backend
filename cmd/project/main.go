package main

import (
	"RIpPeakBack/internal/app"
	"log"
)

func main() {
	log.Println("App Start")

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	application.StartServer()

	log.Println("App term")
}
