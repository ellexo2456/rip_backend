package app

import (
	"RIpPeakBack/docs"
	"RIpPeakBack/internal/app/middleware"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
	"log"
	"mime/multipart"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"RIpPeakBack/internal/app/ds"
	"RIpPeakBack/internal/app/dsn"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	docs.SwaggerInfo.Title = "RIpPeakBack"
	docs.SwaggerInfo.Description = "rip course project about alpinists and their expeditions"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = ""

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/", a.filterAlpinistsByCountry)
	r.GET("/alpinist/:id", a.getAlpinist)

	r.POST("/auth/login", a.Login)
	r.POST("/auth/logout", a.Logout)
	r.POST("/auth/register", a.Register)
	//r.HandleFunc("/api/v1/auth/check", handler.CheckAuth).Methods(http.MethodPost, http.MethodOptions)

	mw := middleware.New(a.rr)
	authorized := r.Group("/")
	authorized.Use(mw.IsAuth())
	{
		authorized.DELETE("/alpinist/:id", a.deleteAlpinist)
		authorized.POST("/alpinist", a.addAlpinist)
		authorized.POST("/alpinist/expedition/:id", a.addAlpinistToLastExpedition)
		authorized.PUT("/alpinist", a.modifyAlpinist)
		r.MaxMultipartMemory = 8 << 20 // 8 MiB
		authorized.POST("/alpinist/image", a.uploadImage)
		authorized.DELETE("/alpinist/expedition/:id", a.deleteAlpinistFromLastExpedition)

		authorized.PUT("/expedition", a.changeExpeditionInfoFields)
		authorized.PUT("/expedition/status/form/:id", a.formExpedition)
		authorized.PUT("/expedition/:id/status", a.changeExpeditionModeratorStatus)
		authorized.GET("/expedition/filter", a.filterExpeditionsByStatusAndFormedTime)
		authorized.GET("/expedition/:id", a.getExpedition)
		authorized.DELETE("/expedition/:id", a.deleteExpedition)
	}

	err := r.Run()
	if err != nil {
		log.Println("Error with running\nServer down")
		return
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

// Login godoc
//
//	@Summary		login user
//	@Description	create user session and put it into cookie
//	@Tags			auth
//	@Accept			json
//	@Param			body	body		ds.Credentials	true	"user credentials"
//	@Success		200		{object}	object{body=object{id=int}}
//	@Failure		400		{object} object{status=string, message=string}
//	@Failure		401		{object} object{status=string, message=string}
//	@Failure		404		{object} object{status=string, message=string}
//	@Failure		409		{object} object{status=string, message=string}
//	@Failure		500		{object} object{status=string, message=string}
//	@Router			/auth/login [post]
func (a *Application) Login(ctx *gin.Context) {
	auth, err := a.auth(ctx)
	if auth == true {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": "you must be unauthorised\"",
		})
		return
	}

	//if err != nil {
	//	ds.WriteError(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//defer a.CloseAndAlert(r.Body)

	var c ds.Credentials
	if err := ctx.ShouldBindJSON(&c); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
		})
		return
	}

	if err = checkCredentials(c); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}
	c.Email = strings.TrimSpace(c.Email)

	expectedUser, err := a.repository.GetByEmail(c.Email)
	if err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	if !checkPasswords(expectedUser.Password, c.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "username or password is invalid",
		})
		return
	}

	session := ds.Session{
		Token:     uuid.NewString(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    int(expectedUser.ID),
		Role:      expectedUser.Role,
	}
	if err = a.rr.Add(session); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	//ctx.SetCookie("session_token", session.Token, int(session.ExpiresAt.Unix()), "/", "localhost", false, true)
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"id": expectedUser.ID,
	})
}

// Logout godoc
//
//	@Summary		logout user
//	@Description	delete current session and nullify cookie
//	@Tags			auth
//	@Success		204
//	@Failure		400	{object}	object{err=string}
//	@Failure		401	{object}	object{err=string}
//	@Failure		404	{object}	object{err=string}
//	@Failure		409	{object}	object{err=string}
//	@Failure		500	{object}	object{err=string}
//	@Router			/auth/logout [post]
func (a *Application) Logout(ctx *gin.Context) {
	auth, err := a.auth(ctx)
	if auth != true {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": "you must be unauthorised",
		})
		return
	}

	t, err := ctx.Cookie("session_token")

	if err = a.rr.DeleteByToken(t); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return

	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
	})

	ctx.Status(http.StatusNoContent)
}

// Register godoc
//
//	@Summary		register user
//	@Description	add new user to db and return it id
//	@Tags			auth
//	@Produce		json
//	@Accept			json
//	@Param			body	body		ds.Credentials	true	"user credentials"
//	@Success		200		{object}	object{body=object{id=int}}
//	@Failure		400		{object} object{status=string, message=string}
//	@Failure		401		{object} object{status=string, message=string}
//	@Failure		409		{object} object{status=string, message=string}
//	@Failure		500		{object} object{status=string, message=string}
//	@Router			/auth/register [post]
func (a *Application) Register(ctx *gin.Context) {
	auth, err := a.auth(ctx)
	if auth == true {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": "you must be unauthorised",
		})
		return
	}

	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid request body",
		})
		return
	}
	user.Email = strings.TrimSpace(user.Email)
	if err = checkCredentials(ds.Credentials{Email: user.Email, Password: user.Password}); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}
	user.Role = ds.Usr

	salt := make([]byte, 8)
	rand.Read(salt)
	user.Password = HashPassword(salt, user.Password)
	var id int
	if id, err = a.repository.AddUser(user); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	session := ds.Session{
		Token:     uuid.NewString(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    id,
		Role:      ds.Usr,
	}
	if err = a.rr.Add(session); err != nil {
		ctx.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	//ctx.SetCookie("session_token", session.Token, int(session.ExpiresAt.Unix()), "/", "localhost", false, true)
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (a *Application) auth(ctx *gin.Context) (bool, error) {
	c, err := ctx.Request.Cookie("session_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return false, ds.ErrUnauthorized
		}

		return false, ds.ErrBadRequest
	}
	if c.Expires.After(time.Now()) {
		return false, ds.ErrUnauthorized
	}
	sessionToken := c.Value
	sc, err := a.rr.SessionExists(sessionToken)
	if err != nil {
		return false, err
	}
	if sc.UserID == 0 {
		return false, ds.ErrUnauthorized
	}

	ctx.Set("sessionContext", sc)

	return true, ds.ErrAlreadyExists
}

func HashPassword(salt []byte, password []byte) []byte {
	hashedPass := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func checkPasswords(passHash []byte, plainPassword []byte) bool {
	salt := passHash[0:8]
	userPassHash := HashPassword(salt, plainPassword)
	return bytes.Equal(userPassHash, passHash)
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func checkCredentials(cred ds.Credentials) error {
	if cred.Email == "" || len(cred.Password) == 0 {
		return ds.ErrWrongCredentials
	}

	if !valid(cred.Email) {
		return ds.ErrWrongCredentials
	}

	return nil
}

// filterAlpinistsByCountry godoc
// @Summary      returns the page with a filtered alpinists
// @Description  returns the page with an alpinists that had been filtered by a country
// @Tags         alpinists
// @Produce      json
// @Param        country query string true "country name"
// @Success      200  {object} object{country=string, alpinists=[]ds.Alpinist, draft=ds.Expedition}
// @Failure      500  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
// @Router       / [get]
func (a *Application) filterAlpinistsByCountry(c *gin.Context) {
	country := c.DefaultQuery("country", "")

	var foundAlpinists *[]ds.Alpinist
	var err error
	if country == "" {
		foundAlpinists, err = a.repository.GetActiveAlpinists()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": err.Error(),
			})
			log.Println("Error with running\nServer down")
			return
		}
	} else {
		foundAlpinists, err = a.repository.FilterByCountry(country)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": err.Error(),
			})
			log.Println("Error with running\nServer down")
			return
		}
	}

	_, _ = a.auth(c)

	value, exists := c.Get("sessionContext")
	var draft ds.Expedition
	var u ds.User

	if exists {
		sc := value.(ds.SessionContext)

		draft, err = a.repository.GetDraft(sc.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				draft = ds.Expedition{}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "fail",
					"message": err.Error(),
				})
				return
			}
		}

		u, err = a.repository.GetUserByID(sc.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				u = ds.User{}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "fail",
					"message": err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":      u,
		"country":   country,
		"draft":     draft,
		"alpinists": *foundAlpinists,
	})
}

// getAlpinist godoc
// @Summary      returns the page of the alpinist
// @Description  returns the page of the alpinist by the provided id
// @Tags         alpinists
// @Produce      json
// @Param        id path uint true "id of alpinist"
// @Success      200  {object} object{alpinist=ds.Alpinist}
// @Failure      500  {object} object{status=string, message=string}
// @Failure      400  {object} object{status=string, message=string}
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

// deleteAlpinistFromLastExpedition godoc
// @Summary      deletes an alpinist
// @Description  deletes an alpinist from expedition
// @Tags         alpinists
// @Produce      json
// @Param        id path uint true "id" alpinist id
// @Success      204
// @Failure      400   {object} object{status=string, message=string}
// @Failure      404   {object} object{status=string, message=string}
// @Failure      500   {object} object{status=string, message=string}
// @Router       /alpinist/expedition/{id} [delete]
func (a *Application) deleteAlpinistFromLastExpedition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid parameter id",
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

	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	draft, err := a.repository.GetDraft(sc.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "no draft",
				"message": err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": err.Error(),
			})
			return
		}
	}

	draft.Alpinists = []ds.Alpinist{*alpinist}
	if err = a.repository.DeleteAlpinistFromExpedition(draft); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t post expedition into db",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// deleteAlpinist godoc
// @Summary      deletes an alpinist
// @Description  deletes an alpinist by a given id and returns the page without it
// @Tags         alpinists
// @Produce      json
// @Param        id path uint true "id"
// @Success      204
// @Failure      400   {object} object{status=string, message=string}
// @Failure      404   {object} object{status=string, message=string}
// @Failure      500   {object} object{status=string, message=string}
// @Router       /alpinist/{id} [delete]
func (a *Application) deleteAlpinist(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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

	alpinist, err := a.repository.GetAlpinistByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	if err = deleteObjectMinio(alpinist.ImageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
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
// @Description  creates expedition and/or adds an alpinist to
// @Tags         alpinists, expeditions
// @Accept       json
// @Produce      json
// @Param        id path uint true "alpinists id"
// @Success      204
// @Failure      400  {json} object{err=string}
// @Failure      404  {json} object{err=string}
// @Failure      500  {json} object{err=string}
// @Router       /alpinist/expedition/{id} [post]
func (a *Application) addAlpinistToLastExpedition(c *gin.Context) {
	alpinistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid alpinistID param",
		})
		return
	}
	if alpinistID < 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "negative alpinistID param",
		})
		return
	}

	//if err := c.ShouldBindJSON(&expedition); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status":  "fail",
	//		"message": "invalid request body",
	//	})
	//	return
	//}

	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	expedition := ds.Expedition{
		Status:    ds.StatusDraft,
		Alpinists: []ds.Alpinist{ds.Alpinist{ID: uint(alpinistID)}},
	}
	if sc.Role == ds.Moderator {
		expedition.ModeratorID = uint(sc.UserID)
	}
	if sc.Role == ds.Usr {
		expedition.UserID = uint(sc.UserID)
	}

	//setTime(&expedition)
	//expedition.Alpinists =
	if err = a.repository.AddExpedition(expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t post expedition into db",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// modifyAlpinist godoc
// @Summary      modify an alpinist
// @Description  modify an alpinist data
// @Tags         alpinists
// @Accept       json
// @Produce      json
// @Failure      400  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
// @Success      200  {object} ds.Alpinist
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
// @Param			body	body		object{id=uint,name=string,year=int}	true	"user credentials"
// @Accept       json
// @Produce      json
// @Failure      400  {object} object{status=string, message=string}
// @Failure      404  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
// @Success      204
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

	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	expedition.UserID = uint(sc.UserID)
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
		c.JSON(ds.GetHttpStatusCode(err), gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// formExpedition godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status to formed
// @Tags         expeditions
// @Accept       json
// @Param        id path uint true "expedition id"
// @Success      204
// @Failure      400   {object} object{status=string, message=string}
// @Failure      403   {object} object{status=string, message=string}
// @Failure      404   {object} object{status=string, message=string}
// @Failure      500   {object} object{status=string, message=string}
// @Router       /expedition/status/form/{id} [put]
func (a *Application) formExpedition(c *gin.Context) {
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

	a.changeStatus(
		c,
		ds.Expedition{
			ID:     uint(id),
			Status: ds.StatusFormed},
		checkUserStatus,
	)
}

// changeExpeditionModeratorStatus godoc
// @Summary      changes an expedition status
// @Description  changes an expedition status with that one witch can be changed by a moderator (deleted or canceled)
// @Tags         expeditions
// @Accept       json
// @Param        id path uint true "expedition id"
//
//	@Param			body	body	object{status=string}	true	"expedition status"
//
// @Success      204
// @Failure      400   {object} object{status=string, message=string}
// @Failure      403   {object} object{status=string, message=string}
// @Failure      404   {object} object{status=string, message=string}
// @Failure      500   {object} object{status=string, message=string}
// @Router       /expedition/{id}/status [put]
func (a *Application) changeExpeditionModeratorStatus(c *gin.Context) {
	var expedition ds.Expedition

	if err := c.ShouldBindJSON(&expedition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status",
		})
		return
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
	expedition.ID = uint(id)

	a.changeStatus(c, expedition, checkModeratorStatus)
}

// filterExpeditionsByStatusAndFormedTime godoc
// @Summary      returns the page with a filtered expeditions
// @Description  returns the page with an expeditions that had been filtered by a status or/and formed time
// @Param        status query string false "new status of the expedition"
// @Param        startTime query string false "start time of interval for filter to formed time"
// @Param        endTime query string false "start time of interval for filter to formed time"
// @Tags         expeditions
// @Produce      json
// @Success      200  {object} object{draft=int, expedition=[]ds.Expedition}
// @Failure      400  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
// @Router       /expedition/filter/ [get]
func (a *Application) filterExpeditionsByStatusAndFormedTime(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	startTime := c.DefaultQuery("startTime", "")
	endTime := c.DefaultQuery("endTime", "")

	//if startTime != "" && endTime == "" || startTime == "" && endTime != "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status":  "fail",
	//		"message": "missing times parameter",
	//	})
	//	return
	//}

	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	var foundExpeditions *[]ds.Expedition
	var err error
	if status == "" && startTime == "" && endTime == "" {
		foundExpeditions, err = a.repository.GetExpeditions(sc)
	}
	if status != "" && startTime == "" && endTime == "" {
		foundExpeditions, err = a.repository.FilterByStatus(status, sc)
	}
	if status == "" && startTime != "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTime(startTime, endTime, sc)
	}

	if status == "" && startTime != "" && endTime == "" {
		foundExpeditions, err = a.repository.FilterByFormedTime(startTime, time.Now().Add(8760*time.Hour).Format("2006-01-02"), sc)
	}
	if status == "" && startTime == "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTime(ds.MinDate, endTime, sc)
	}

	if status != "" && startTime != "" && endTime == "" {
		foundExpeditions, err = a.repository.FilterByFormedTimeAndStatus(startTime, time.Now().Add(8760*time.Hour).Format("2006-01-02"), status, sc)
	}
	if status != "" && startTime == "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTimeAndStatus(ds.MinDate, endTime, status, sc)
	}

	if status != "" && startTime != "" && endTime != "" {
		foundExpeditions, err = a.repository.FilterByFormedTimeAndStatus(startTime, endTime, status, sc)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	draft, err := a.repository.GetDraft(sc.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			draft = ds.Expedition{}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"expedition": *foundExpeditions,
		"draft":      draft.ID,
	})
}

// addAlpinist godoc
// @Summary      adds the alpinist
// @Description  creates the alpinist and puts it to db
// @Tags         alpinists
// @Accept       json
// @Produce      json
// @Success      200  {object} object{id=int}
// @Failure      400  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
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
// @Success      200  {object} object{expedition=ds.Expedition}
// @Failure      500  {object} object{status=string, message=string}
// @Failure      400  {object} object{status=string, message=string}
// @Failure      404  {object} object{status=string, message=string}
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

	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	expedition, err := a.repository.GetExpeditionByID(expeditionID, sc.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "can`t find expedition with such id for the user",
				"message": err.Error(),
			})
			return
		}

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
// @Param        id path uint true "expedition id"
// @Success      204
// @Failure      500   {object} object{status=string, message=string}
// @Failure      400   {object} object{status=string, message=string}
// @Failure      404   {object} object{status=string, message=string}
// @Failure      500   {object} object{status=string, message=string}
// @Router       /expedition/{id} [delete]
func (a *Application) deleteExpedition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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

	//expedition, err := a.repository.GetExpeditionByID(id)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"status":  "fail",
	//		"message": err.Error(),
	//	})
	//	return
	//}

	expedition := ds.Expedition{
		ID:       uint(id),
		Status:   ds.StatusDeleted,
		ClosedAt: time.Now(),
	}
	err = a.repository.DeleteExpedition(expedition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// uploadImage godoc
// @Summary      uploads image
// @Description  uploads image to minio and modifies image data in db
// @Tags         alpinists
// @Accept       json
// @Produce      json
// @Success      200  {object} object{status=string, message=string}
// @Failure      400  {object} object{status=string, message=string}
// @Failure      404  {object} object{status=string, message=string}
// @Failure      500  {object} object{status=string, message=string}
// @Router       /alpinist/image [post]
func (a *Application) uploadImage(c *gin.Context) {
	strId := c.DefaultQuery("id", "")
	id, err := strconv.Atoi(strId)
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

	alpinist, err := a.repository.GetAlpinistByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	if alpinist.ImageName != "" {
		if err = deleteObjectMinio(alpinist.ImageName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "fail",
				"message": err.Error(),
			})
			return
		}
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}
	log.Println(file.Filename)

	objectName := strId + "/" + strings.ReplaceAll(file.Filename, " ", "")
	err = uploadToMinio(objectName, file, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}
	filePathMinio := "http://" + os.Getenv("E3_ENDPOINT") + "/" + os.Getenv("E3_BUCKET") + "/" + objectName

	alpinist.ImageName = objectName
	alpinist.ImageRef = filePathMinio
	if err = a.repository.UpdateAlpinist(*alpinist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "uploaded",
		"message": filePathMinio,
	})
}

func deleteObjectMinio(objectName string) error {
	minioClient, err := minio.New(os.Getenv("E3_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("E3_ID"), os.Getenv("E3_SECRET"), ""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}

	err = minioClient.RemoveObject(context.Background(), os.Getenv("E3_BUCKET"), objectName, opts)
	if err != nil {
		return err
	}

	return nil
}

func uploadToMinio(objectName string, file *multipart.FileHeader, contentType string) error {
	// Initialize minio client object
	//minioClient, err := minio.New("localhost:9001", "minio", "minio124", true)
	minioClient, err := minio.New(os.Getenv("E3_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("E3_ID"), os.Getenv("E3_SECRET"), ""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, os.Getenv("E3_BUCKET"), minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		exists, err := minioClient.BucketExists(ctx, os.Getenv("E3_BUCKET"))
		if err == nil && exists {
			log.Printf("Bucket:%s is already exist\n", os.Getenv("E3_BUCKET"))
		} else {
			return err
		}
	}
	log.Printf("Successfully created bucket: %s\n", os.Getenv("E3_BUCKET"))

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = minioClient.PutObject(ctx, os.Getenv("E3_BUCKET"), objectName, src, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

func setTime(expedition *ds.Expedition) {
	if expedition.Status == ds.StatusFormed {
		expedition.FormedAt = time.Now()
	}
	if expedition.Status == ds.StatusCanceled || expedition.Status == ds.StatusDeleted {
		expedition.ClosedAt = time.Now()
	}
}

func (a *Application) changeStatus(c *gin.Context, expedition ds.Expedition, checkStatus func(ds.Expedition, ds.SessionContext) bool) {
	value, exists := c.Get("sessionContext")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "must be authorized",
		})
		return
	}
	sc := value.(ds.SessionContext)

	expeditionWithStatus, err := a.repository.GetExpeditionById(expedition.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "fail",
			"message": "id is out of range",
		})
		return
	}
	expedition.UserID = expeditionWithStatus.UserID
	expedition.ModeratorID = expeditionWithStatus.ModeratorID

	if !checkStatus(expedition, sc) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "invalid status or user"})
		return
	}

	expedition.FormedAt = expeditionWithStatus.FormedAt
	expedition.ClosedAt = expeditionWithStatus.ClosedAt
	setTime(&expedition)

	if expedition.Status == ds.StatusCanceled || expedition.Status == ds.StatusDenied {
		if expeditionWithStatus.Status != ds.StatusFormed {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "fail",
				"message": "can`t close order that isn`t open",
			})

			return
		}
		expedition.ClosedAt = time.Now()
	}

	if err := a.repository.UpdateStatus(expedition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "can`t update status in db",
		})
		return
	}

	c.Status(http.StatusNoContent)
	return
}

func checkUserStatus(expedition ds.Expedition, sc ds.SessionContext) bool {
	if expedition.Status != ds.StatusFormed {
		return false
	}
	if uint(sc.UserID) != expedition.UserID || sc.Role != ds.Usr {
		return false
	}
	return true
}

func checkModeratorStatus(expedition ds.Expedition, sc ds.SessionContext) bool {
	if expedition.Status != ds.StatusCanceled && expedition.Status != ds.StatusDenied {
		return false
	}
	if uint(sc.UserID) != expedition.ModeratorID && sc.Role != ds.Moderator {
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
