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

	services := [][]string{{"0", "k2", "hard", "image/evr.jpg"}, {"1", "everest", "toze hard but less hard", "image/evr.jpg"}}

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

	router.Static("/image", "./resources")

	err := router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}
