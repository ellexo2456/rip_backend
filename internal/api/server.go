package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.tmpl", gin.H{
			"title":    "some service",
			"services": [][]string{{"k2", "hard", "image/evr.jpg"}, {"everest", "toze hard but less hard", "image/evr.jpg"}},
		})
		c.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"title":    "some service",
			"services": [][]string{{"k2", "hard", "image/evr.jpg"}, {"everest", "toze hard but less hard", "image/evr.jpg"}},
		})
	})

	r.Static("/image", "./resources")

	err := r.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}
