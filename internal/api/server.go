package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Server start up")

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	services := [][]string{
		{"0", "Райнхольд Андреас Месснер", "Италия", "17 сентября 1944 - нн (78 лет)", "image/rm.jpg"},
		{"1", "Юзеф Е́жи «Юрек» Куку́чка", " Польша", "24 марта 1948 - 24 октября 1989 (41 год)", "image/evr.jpg"},
	}

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"services": services,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"services": services,
		})
	})

	router.GET("/service/:id", func(context *gin.Context) {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.AbortWithStatus(404)
			return
		}

		if id >= len(services) || id < 0 {
			context.AbortWithStatus(404)
			return
		}

		context.HTML(http.StatusOK, "card.tmpl", gin.H{
			"services": services,
			"id":       id,
		})
	})

	router.Static("/image", "./static/images")

	err := router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}
