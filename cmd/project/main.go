package main

import (
	"RIpPeakBack/internal/api"
	"log"
)

func main() {
	log.Println("App Start")
	api.StartServer()
	log.Println("App term")
}
