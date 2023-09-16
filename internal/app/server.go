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

	router.GET("/", a.loadMainPage)
	router.GET("/alpinist/:id", a.getAlpinistPage)
	router.GET("/alpinist/filter", a.filterAlpinistsByCountry)
	router.GET("/alpinist/delete", a.deleteAlpinist)
	router.POST("/expedition", a.addService)
	router.PUT("/expedition", a.changeExpeditionInfoFields)
	router.PUT("/expedition/status/user", a.changeExpeditionUserStatus)
	router.PUT("/expedition/status/moderator", a.changeExpeditionModeratorStatus)
	router.GET("/expedition/filter", a.filterExpeditionsByStatus)

	router.Static("/image", "./static/images")

	err := router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

// loadMainPage godoc
// @Summary      returns the main page
// @Description  load alpinists from db and returns the main page with them as a context
// @Tags         alpinists
// @Produce      html
// @Success      200  {array} ds.Alpinist
// @Failure      500  {string} string "can`t get the alpinists list"
// @Router       / [get]
func (a *Application) loadMainPage(c *gin.Context) {
	alpinists, err := a.repository.GetActiveAlpinists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t get the alpinists list",
		})
	}

	c.HTML(http.StatusOK, "base.tmpl", gin.H{
		"alpinists": *alpinists,
	})
	c.HTML(http.StatusOK, "card_item.tmpl", gin.H{
		"alpinists": *alpinists,
	})
}

// getAlpinistPage godoc
// @Summary      returns the page of the alpinist
// @Description  returns the page of the alpinist by the provided id
// @Tags         alpinists
// @Produce      html
// @Param        id path uint true "id os alpinist"
// @Success      200  {object} ds.Alpinist
// @Failure      500  {string} string "can`t get the alpinists list"
// @Failure      400  {string} string "invalid parameter id"
// @Failure      400  {string} string "negative parameter id"
// @Failure      404  {string} string "id is out of rage"
// @Router       /alpinist/{id} [get]
func (a *Application) getAlpinistPage(c *gin.Context) {
	alpinists, err := a.repository.GetActiveAlpinists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t get the alpinists list",
		})
	}

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
}

// filterAlpinistsByCountry godoc
// @Summary      returns the page with a filtered alpinists
// @Description  returns the page with an alpinists that had been filtered by a country
// @Tags         alpinists
// @Produce      html
// @Param        name query string true "country name"
// @Success      200  {array} ds.Alpinist
// @Failure      500  {string} string "can`t get the alpinists list"
// @Router       /alpinist/filter/{name} [get]
func (a *Application) filterAlpinistsByCountry(c *gin.Context) {
	alpinists, err := a.repository.GetActiveAlpinists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t get the alpinists list",
		})
	}

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
}

// deleteAlpinist godoc
// @Summary      deletes an alpinist
// @Description  deletes an alpinist by a given id and returns the page without it
// @Tags         alpinists
// @Produce      html
// @Param        id query uint true "alpinists id"
// @Success      200  {array} ds.Alpinist
// @Failure      500  {string} string "can`t get the alpinists list"
// @Failure      400  {string} string "invalid parameter id"
// @Failure      400  {string} string "negative parameter id"
// @Failure      500  {string} string "can`t open db connection"
// @Failure      500  {string} string "can`t delete alpinist in db"
// @Failure      500  {string} string "can`t delete alpinist in db"
// @Router       /alpinist/delete/{id} [get]
func (a *Application) deleteAlpinist(c *gin.Context) {
	alpinists, err := a.repository.GetActiveAlpinists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t get the alpinists list",
		})
	}

	id, err := strconv.Atoi(c.DefaultQuery("id", ""))
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

	var activeAlpinists []ds.Alpinist
	var alpinistToDelete ds.Alpinist
	var flag bool
	for _, alpinist := range *alpinists {
		if alpinist.ID != uint(id) {
			activeAlpinists = append(activeAlpinists, alpinist)
		} else {
			alpinistToDelete = alpinist
			flag = true
		}
	}

	if !flag {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "id is out of range",
		})
		return
	}

	var db *sql.DB
	_ = godotenv.Load()
	db, err = sql.Open("postgres", dsn.FromEnv())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t open db connection",
		})
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE alpinists SET status = $1 WHERE id = $2", "удалён", alpinistToDelete.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t delete alpinist in db",
		})
		return
	}

	c.HTML(http.StatusOK, "base.tmpl", gin.H{
		"alpinists": activeAlpinists,
	})
	c.HTML(http.StatusOK, "card_item.tmpl", gin.H{
		"alpinists": activeAlpinists,
	})
}

// addService godoc
// @Summary      adds an alpinist to expedition
// @Description  creates expedition and adds an alpinist to
// @Tags         alpinists, expeditions
// @Accept       json
// @Produce      json
// @Success      200  {object} ds.Expedition
// @Failure      400  {string} string "invalid request body"
// @Failure      400  {string} string "invalid query parameters (ids)"
// @Failure      400  {string} string "negative id"
// @Failure      404  {string} string "id is out of range"
// @Failure      500  {string} string "can`t post expedition into db"
// @Router       /expedition [post]
func (a *Application) addService(c *gin.Context) {
	expedition, errMessage, code := getExpedition(c, a)
	if expedition == nil {
		c.JSON(code, gin.H{
			"status":  "fail",
			"message": errMessage,
		})
		return
	}

	expedition.CreatedAt = time.Now()
	var err error
	if expedition.ID, err = a.repository.AddExpedition(*expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t post expedition into db",
		})
		return
	}

	c.JSON(http.StatusOK, expedition)
	return
}

// changeExpeditionInfoFields godoc
// @Summary      changes an expedition
// @Description  changes an expedition information fields that can be changed by a user
// @Tags         expeditions
// @Accept       json
// @Produce      json
// @Failure      400  {string} string "invalid request body"
// @Failure      400  {string} string "invalid query parameters (ids)"
// @Failure      400  {string} string "negative id"
// @Failure      404  {string} string "id is out of range"
// @Failure      500  {string} string "can`t post expedition into db"
// @Success      200  {object} ds.Expedition
// @Router       /expedition [put]
func (a *Application) changeExpeditionInfoFields(c *gin.Context) {
	expedition, errMessage, code := getExpedition(c, a)
	if expedition == nil {
		c.JSON(code, gin.H{
			"status":  "fail",
			"message": errMessage,
		})
		return
	}
	dbExpedition, err := a.repository.GetExpeditionById(expedition.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "can`t found expedition with such id",
		})
		return
	}

	expedition.Status = dbExpedition.Status
	expedition.CreatedAt = dbExpedition.CreatedAt
	expedition.FormedAt = dbExpedition.FormedAt
	expedition.ClosedAt = dbExpedition.ClosedAt

	if err = a.repository.UpdateExpedition(*expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t update expedition in db",
		})
		return
	}

	c.JSON(http.StatusOK, expedition)
	return
}

// changeExpeditionUserStatus godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status with that one witch can be changed by a user
// @Tags         expeditions
// @Param        id query uint true "expedition id"
// @Accept       json
// @Success      204
// @Failure      400  {string} string "invalid status"
// @Failure      400  {string} string "invalid query parameters (ids)"
// @Failure      400  {string} string "negative id"
// @Failure      404  {string} string "id is out of range"
// @Failure      500  {string} string "can`t update status in db"
// @Router       /expedition/status/user/{id} [put]
func (a *Application) changeExpeditionUserStatus(c *gin.Context) {
	changeStatus(c, a, checkUserStatus)
}

// changeExpeditionModeratorStatus godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status with that one witch can be changed by a moderator
// @Tags         expeditions
// @Param        id query uint true "expedition id"
// @Accept       json
// @Success      204
// @Failure      400  {string} string "invalid status"
// @Failure      400  {string} string "invalid query parameters (ids)"
// @Failure      400  {string} string "negative id"
// @Failure      404  {string} string "id is out of range"
// @Failure      500  {string} string "can`t update status in db"
// @Router       /expedition/status/moderator/{id} [put]
func (a *Application) changeExpeditionModeratorStatus(c *gin.Context) {
	changeStatus(c, a, checkModeratorStatus)
}

// filterExpeditionsByStatus godoc
// @Summary      returns the page with a filtered expeditions
// @Description  returns the page with an expeditions that had been filtered by a status
// @Param        name query string true "new status of the expedition"
// @Tags         expeditions
// @Produce      html
// @Success      200  {array} ds.Expedition
// @Failure      500  {string} string "error with db"
// @Router       /expedition/filter/{name} [get]
func (a *Application) filterExpeditionsByStatus(c *gin.Context) {
	searchQuery := c.DefaultQuery("name", "")

	foundExpeditions, err := a.repository.FilterByStatus(searchQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "error with db",
		})
		return
	}

	// заменить на шаблоны с заявками
	c.JSON(http.StatusOK, foundExpeditions)
}

func changeStatus(c *gin.Context, a *Application, checkStatus func(string) bool) {
	var expedition ds.Expedition

	if err := c.ShouldBindJSON(&expedition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status",
		})
		return
	}

	if !checkStatus(expedition.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status",
		})
		return
	}

	if expedition.Status == "завершён" {
		expedition.FormedAt = time.Now()
	}
	if expedition.Status == "удалён" {
		expedition.ClosedAt = time.Now()
	}

	id, err := strconv.Atoi(c.DefaultQuery("id", ""))
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

	expedition.ID = uint(id)
	if _, err := a.repository.GetExpeditionById(expedition.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "id is out of range",
		})
		return
	}

	if err := a.repository.UpdateStatus(expedition.Status, expedition.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t update status in db",
		})
		return
	}

	c.Status(http.StatusNoContent)
	return
}

func checkUserStatus(status string) bool {
	if status != "введён" && status != "в работе" {
		return false
	}
	return true
}

func checkModeratorStatus(status string) bool {
	if status != "завершён" && status != "удалён" && status != "отменён" {
		return false
	}
	return true
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
