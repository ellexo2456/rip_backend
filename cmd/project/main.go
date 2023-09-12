package main

import (
	"RIpPeakBack/internal/app"
	"log"
)

func main() {
	log.Println("App Start")
	app.StartServer()
	log.Println("App term")
}
