package app

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
	"time"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	router := gin.Default()

	router.GET("/", a.filterAlpinistsByCountry)
	router.GET("/alpinist/:id", a.getAlpinist)
	router.POST("/alpinist/delete", a.deleteAlpinist)
	router.POST("/alpinist", a.addAlpinist)
	router.POST("/alpinist/expedition", a.addAlpinistToLastExpedition)
	router.PUT("/alpinist", a.modifyAlpinist)

	router.PUT("/expedition", a.changeExpeditionInfoFields)
	router.PUT("/expedition/status/user", a.changeExpeditionUserStatus)
	router.PUT("/expedition/status/moderator", a.changeExpeditionModeratorStatus)
	router.GET("/expedition/filter", a.filterExpeditionsByStatusAndFormedTime)
	router.GET("/expedition/:id", a.getExpedition)
	router.DELETE("/expedition", a.deleteExpedition)

	err := router.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

// filterAlpinistsByCountry godoc
// @Summary      returns the page with a filtered alpinists
// @Description  returns the page with an alpinists that had been filtered by a country
// @Tags         alpinists
// @Produce      json
// @Param        country query string true "country name"
// @Success      200  {json}
// @Failure      500  {json}
// @Router       /{name} [get]
func (a *Application) filterAlpinistsByCountry(c *gin.Context) {
	country := c.DefaultQuery("country", "")

	var foundAlpinists *[]ds.Alpinist
	var err error
	if country == "" {
		foundAlpinists, err = a.repository.GetActiveAlpinists()
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}
	} else {
		foundAlpinists, err = a.repository.FilterByCountry(country)
		if err != nil {
			log.Println("Error with running\nServer down")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"country":   country,
		"alpinists": *foundAlpinists,
	})
}

// getAlpinist godoc
// @Summary      returns the page of the alpinist
// @Description  returns the page of the alpinist by the provided id
// @Tags         alpinists
// @Produce      json
// @Param        id path uint true "id of alpinist"
// @Success      200  {json}
// @Failure      500  {json}
// @Failure      400  {json}
// @Router       /alpinist/{id} [get]
func (a *Application) getAlpinist(c *gin.Context) {
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

	alpinist, err := a.repository.GetAlpinistByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alpinist": alpinist,
	})
}

// deleteAlpinist godoc
// @Summary      deletes an alpinist
// @Description  deletes an alpinist by a given id and returns the page without it
// @Tags         alpinists
// @Produce      json
// @Param        id query uint true "alpinists id"
// @Success      204
// @Failure      400  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Router       /alpinist/delete/{id} [post]
func (a *Application) deleteAlpinist(c *gin.Context) {
	id, err := strconv.Atoi(c.DefaultQuery("id", ""))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid id param",
		})
		return
	}
	if id < 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "negative id param",
		})
		return
	}

	var db *sql.DB
	_ = godotenv.Load()
	db, err = sql.Open("postgres", dsn.FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE alpinists SET status = $1 WHERE id = $2", "удалён", id)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.Status(http.StatusNoContent)
}

// addAlpinistToLastExpedition godoc
// @Summary      adds an alpinist to expedition
// @Description  creates expedition and adds an alpinist to
// @Tags         alpinists, expeditions
// @Accept       json
// @Produce      json
// @Success      200  {json}
// @Failure      400  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Router       /alpinist/expedition [post]
func (a *Application) addAlpinistToLastExpedition(c *gin.Context) {
	var expedition ds.Expedition
	if err := c.ShouldBindJSON(&expedition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
		})
		return
	}

	expedition.UserID = ds.UserID
	expedition.ModeratorID = ds.UserID
	setTime(&expedition)

	if expedition.Status != ds.StatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status",
		})
		return
	}

	expedition.CreatedAt = time.Now()
	var err error
	if expedition.ID, err = a.repository.AddExpedition(expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t post expedition into db",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": expedition.ID,
	})
}

// modifyAlpinist godoc
// @Summary      modify an alpinist
// @Description  modify an alpinist data
// @Tags         alpinists
// @Accept       json
// @Produce      json
// @Failure      400  {json}
// @Failure      500  {json}
// @Success      200  {json}
// @Router       /alpinist [put]
func (a *Application) modifyAlpinist(c *gin.Context) {
	var alpinist ds.Alpinist
	if err := c.ShouldBindJSON(&alpinist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
		})
		return
	}

	if err := a.repository.UpdateAlpinist(alpinist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t update expedition in db",
		})
		return
	}

	c.JSON(http.StatusOK, alpinist)
}

// changeExpeditionInfoFields godoc
// @Summary      changes an expedition
// @Description  changes an expedition information fields that can be changed by a user
// @Tags         expeditions
// @Accept       json
// @Produce      json
// @Failure      400  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Success      200  {json}
// @Router       /expedition [put]
func (a *Application) changeExpeditionInfoFields(c *gin.Context) {
	var expedition ds.Expedition
	if err := c.ShouldBindJSON(&expedition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
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

	if err = a.repository.UpdateExpedition(expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t update expedition in db",
		})
		return
	}

	c.JSON(http.StatusOK, expedition)
}

// changeExpeditionUserStatus godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status with that one witch can be changed by a user
// @Tags         expeditions
// @Accept       json
// @Success      204
// @Failure      400  {json}
// @Failure      403  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Router       /expedition/status/user/{id} [put]
func (a *Application) changeExpeditionUserStatus(c *gin.Context) {
	changeStatus(c, a, checkUserStatus)
}

// changeExpeditionModeratorStatus godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status with that one witch can be changed by a moderator
// @Tags         expeditions
// @Accept       json
// @Success      204
// @Failure      400  {json}
// @Failure      403  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Router       /expedition/status/moderator/{id} [put]
func (a *Application) changeExpeditionModeratorStatus(c *gin.Context) {
	changeStatus(c, a, checkModeratorStatus)
}

// filterExpeditionsByStatusAndFormedTime godoc
// @Summary      returns the page with a filtered expeditions
// @Description  returns the page with an expeditions that had been filtered by a status or/and formed time
// @Param        status query string false "new status of the expedition"
// @Param        startTime query string false "start time of interval for filter to formed time"
// @Param        endTime query string false "start time of interval for filter to formed time"
// @Tags         expeditions
// @Produce      json
// @Success      200  {json}
// @Failure      400  {json}
// @Failure      500  {json}
// @Router       /expedition/filter/{status} [get]
func (a *Application) filterExpeditionsByStatusAndFormedTime(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	startTime := c.DefaultQuery("startTime", "")
	endTime := c.DefaultQuery("endTime", "")

	if startTime != "" && endTime == "" || startTime == "" && endTime != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "missing times parameter",
		})
		return
	}

	var foundExpeditions *[]ds.Expedition
	var err error
	if status == "" && startTime == "" && endTime == "" {
		foundExpeditions, err = a.repository.GetExpeditions()
	}
	if status != "" && startTime == "" && endTime == "" {
		foundExpeditions, err = a.repository.FilterByStatus(status)
	}
	if status == "" && startTime != "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTime(startTime, endTime)
	}
	if status != "" && startTime != "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTimeAndStatus(startTime, endTime, status)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	draft, err := a.repository.GetDraft(ds.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"expedition": *foundExpeditions,
		"draft":      draft.ID,
	})
}

// addAlpinist godoc
// @Summary      adds the alpinist
// @Description  creates the alpinist and puts it to db
// @Tags         alpinists, expeditions
// @Accept       json
// @Produce      json
// @Success      200  {json}
// @Failure      400  {json}
// @Failure      500  {json}
// @Router       /alpinist [post]
func (a *Application) addAlpinist(c *gin.Context) {
	var alpinist ds.Alpinist
	if err := c.ShouldBindJSON(&alpinist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
		})
		return
	}

	var err error
	if alpinist.ID, err = a.repository.AddAlpinist(alpinist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": alpinist.ID,
	})
	return
}

// getExpedition godoc
// @Summary      returns the expedition by its id
// @Description  returns the expedition by its id
// @Tags         expeditions
// @Produce      json
// @Param        id path uint true "id of expedition"
// @Success      200  {json}
// @Failure      500  {json}
// @Failure      400  {json}
// @Failure      404  {json}
// @Router       /expedition/{id} [get]
func (a *Application) getExpedition(c *gin.Context) {
	expeditionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid parameter id",
		})
		return
	}

	if expeditionID < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "negative parameter id",
		})
		return
	}

	expedition, err := a.repository.GetExpeditionByID(expeditionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expedition": expedition,
	})
}

// deleteExpedition godoc
// @Summary      deletes an expedition
// @Description  deletes an expedition from db
// @Tags         expeditions
// @Produce      json
// @Param        id query uint true "expedition id"
// @Success      204
// @Failure      500  {json}
// @Failure      400  {json}
// @Failure      404  {json}
// @Failure      500  {json}
// @Router       /expedition/{id} [delete]
func (a *Application) deleteExpedition(c *gin.Context) {
	id, err := strconv.Atoi(c.DefaultQuery("id", ""))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid id param",
		})
		return
	}
	if id < 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "negative id param",
		})
		return
	}

	expedition, err := a.repository.GetExpeditionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	err = a.repository.DeleteExpedition(*expedition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func setTime(expedition *ds.Expedition) {
	if expedition.Status == ds.StatusFormed {
		expedition.FormedAt = time.Now()
	}
	if expedition.Status == ds.StatusCanceled {
		expedition.ClosedAt = time.Now()
	}
}

func changeStatus(c *gin.Context, a *Application, checkStatus func(ds.Expedition, int) bool) {
	var expedition ds.Expedition

	if err := c.ShouldBindJSON(&expedition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status",
		})
		return
	}

	expeditionWithStatus, err := a.repository.GetExpeditionById(expedition.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "id is out of range",
		})
		return
	}
	expedition.UserID = expeditionWithStatus.UserID

	if !checkStatus(expedition, ds.UserID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status or user",
		})
		return
	}

	expedition.UserID = ds.UserID
	expedition.ModeratorID = ds.UserID

	setTime(&expedition)

	//id, err := strconv.Atoi(c.DefaultQuery("id", ""))
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status":  "fail",
	//		"message": "invalid parameter id",
	//	})
	//	return
	//}
	//
	//if id < 0 {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status":  "fail",
	//		"message": "negative parameter id",
	//	})
	//	return
	//}

	if expedition.Status == ds.StatusFormed && expeditionWithStatus.Status != ds.StatusDraft {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "fail",
			"message": "can`t form order that isn`t a draft",
		})
		return
	}

	if expedition.Status == ds.StatusCanceled || expedition.Status == ds.StatusDenied && expeditionWithStatus.Status != ds.StatusFormed {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "fail",
			"message": "can`t close order that isn`t open",
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

func checkUserStatus(expedition ds.Expedition, id int) bool {
	if expedition.Status != ds.StatusDraft && expedition.Status != ds.StatusFormed && expedition.Status != ds.StatusDeleted {
		return false
	}
	if expedition.UserID != uint(id) {
		return false
	}
	return true
}

func checkModeratorStatus(expedition ds.Expedition, id int) bool {
	if expedition.Status != ds.StatusCanceled && expedition.Status != ds.StatusDenied {
		return false
	}
	return true
}

//func toIntArray(str string) ([]int, error) {
//	chunks := strings.Split(str, ",")
//
//	var res []int
//	for _, c := range chunks {
//		if i, err := strconv.Atoi(c); err != nil {
//			return nil, err
//		} else {
//			res = append(res, i)
//		}
//	}
//
//	return res, nil
//}
