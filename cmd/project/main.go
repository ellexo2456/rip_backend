package main

import (
	"RIpPeakBack/internal/app"
	"log"
)

// @title RIpPeakBack
// @version 1.0
// @description rip course project about alpinists and their expeditions

// @contact.name Alex Chinaev
// @contact.url https://vk.com/l.chinaev
// @contact.email ax.chinaev@yandex.ru

// @license.name AS IS (NO WARRANTY)

// @host 127.0.0.1
// @schemes http
// @BasePath /
func main() {
	log.Println("App Start")

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	application.StartServer()

	log.Println("App term")
}
