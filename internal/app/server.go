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
	"time"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	alpinists, err := a.repository.GetActiveAlpinists()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	}

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "base.tmpl", gin.H{
			"alpinists": *alpinists,
		})
		context.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"alpinists": *alpinists,
		})
	})

	router.GET("/alpinist/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "fail",
				"message": "invalid parameter id",
			})
			return
		}

		if id < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "fail",
				"message": "negative parameter id",
			})
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
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "fail",
				"message": "id is out of rage",
			})
			return
		}

		c.HTML(http.StatusOK, "card.tmpl", gin.H{
			"alpinist": alpinist,
		})
	})

	router.GET("/alpinist/filter", func(c *gin.Context) {
		searchQuery := c.DefaultQuery("name", "")

		var foundAlpinists []ds.Alpinist
		for _, alpinist := range *alpinists {
			if strings.HasPrefix(strings.ToLower(alpinist.Country), strings.ToLower(searchQuery)) {
				foundAlpinists = append(foundAlpinists, alpinist)
			}
		}

		c.HTML(http.StatusOK, "base.tmpl", gin.H{
			"alpinists": foundAlpinists,
		})
		c.HTML(http.StatusOK, "card_item.tmpl", gin.H{
			"alpinists": foundAlpinists,
		})
	})

	router.GET("/alpinist/delete", func(context *gin.Context) {
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

	router.POST("/expedition", func(c *gin.Context) {
		expedition, errMessage, code := getExpedition(c, a)
		if expedition == nil {
			c.JSON(code, gin.H{
				"status":  "fail",
				"message": errMessage,
			})
			return
		}

		expedition.CreatedAt = time.Now()

		if expedition.ID, err = a.repository.AddExpedition(*expedition); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": "can`t post expedition into db",
			})
			return
		}

		c.JSON(http.StatusOK, expedition)
		return
	})

	router.PUT("/expedition/", func(c *gin.Context) {
		expedition, errMessage, code := getExpedition(c, a)
		if expedition == nil {
			c.JSON(code, gin.H{
				"status":  "fail",
				"message": errMessage,
			})
			return
		}

		if err = a.repository.UpdateExpedition(*expedition); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": "can`t update expedition in db",
			})
			return
		}

		c.JSON(http.StatusOK, expedition)
		return
	})

	router.Static("/image", "./static/images")

	err = router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

func getExpedition(c *gin.Context, a *Application) (*ds.Expedition, string, int) {
	var expedition = ds.Expedition{}

	var err error
	if err = c.ShouldBindJSON(&expedition); err != nil {
		return nil, "invalid request body", http.StatusBadRequest
	}

	query := c.DefaultQuery("ids", "")
	var ids []int
	if ids, err = toIntArray(query); err != nil {
		return nil, "invalid query parameters (ids)", http.StatusBadRequest
	}

	for _, alpinistId := range ids {
		if alpinistId < 0 {
			return nil, "negative id", http.StatusBadRequest
		}

		if alp, err := a.repository.GetAlpinistByID(alpinistId); err != nil {
			return nil, "id is out of range", http.StatusNotFound

		} else {
			expedition.Alpinists = append(expedition.Alpinists, *alp)
		}
	}

	return &expedition, "", 0
}

func toIntArray(str string) ([]int, error) {
	chunks := strings.Split(str, ",")

	var res []int
	for _, c := range chunks {
		if i, err := strconv.Atoi(c); err != nil {
			return nil, err
		} else {
			res = append(res, i)
		}
	}

	return res, nil
}
