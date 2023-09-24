package app

import "C"
import (
	"RIpPeakBack/internal/app/ds"
	"RIpPeakBack/internal/app/dsn"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(context *gin.Context) {
		alpinists, err := a.repository.GetActiveAlpinists()
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}

		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"alpinists": *alpinists,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"alpinists": *alpinists,
		})
	})

	router.GET("/service/:id", func(context *gin.Context) {
		alpinists, err := a.repository.GetActiveAlpinists()
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}

		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.AbortWithStatus(404)
			return
		}

		if id < 0 {
			context.AbortWithStatus(404)
			return
		}

		if len(*alpinists) == 0 {
			context.AbortWithStatus(404)
			return
		}

		var alpinist ds.Alpinist
		flag := true
		for _, alp := range *alpinists {
			if alp.ID == uint(id) {
				alpinist = alp
				flag = false
				break
			}
		}

		if flag {
			context.AbortWithStatus(404)
		}

		context.HTML(http.StatusOK, "card.tmpl", gin.H{
			"alpinist": alpinist,
		})
	})

	router.GET("/filter", func(context *gin.Context) {
		alpinists, err := a.repository.GetActiveAlpinists()
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}

		searchQuery := context.DefaultQuery("name", "")

		var foundAlpinists []ds.Alpinist
		for _, alpinist := range *alpinists {
			if strings.HasPrefix(strings.ToLower(alpinist.Country), strings.ToLower(searchQuery)) {
				foundAlpinists = append(foundAlpinists, alpinist)
			}
		}

		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"alpinists": foundAlpinists,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"alpinists": foundAlpinists,
		})
	})

	router.GET("/service/delete", func(context *gin.Context) {
		alpinists, err := a.repository.GetActiveAlpinists()
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}

		id, err := strconv.Atoi(context.DefaultQuery("name", ""))
		if err != nil {
			context.AbortWithStatus(404)
			return
		}

		if id < 0 {
			context.AbortWithStatus(404)
			return
		}

		if len(*alpinists) == 0 {
			context.AbortWithStatus(404)
			return
		}

		var activeAlpinists []ds.Alpinist
		var alpinistToDelete ds.Alpinist

		for _, alpinist := range *alpinists {
			if alpinist.ID != uint(id) {
				activeAlpinists = append(activeAlpinists, alpinist)
			} else {
				alpinistToDelete = alpinist
			}
		}

		var db *sql.DB
		_ = godotenv.Load()
		db, err = sql.Open("postgres", dsn.FromEnv())
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		_, err = db.Exec("UPDATE alpinists SET status = $1 WHERE id = $2", "удалён", alpinistToDelete.ID)
		if err != nil {
			context.AbortWithStatus(500)
			return
		}

		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"alpinists": activeAlpinists,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"alpinists": activeAlpinists,
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
