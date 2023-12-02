package middleware

import (
	"RIpPeakBack/internal/app/ds"
	"RIpPeakBack/internal/app/repository"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Middleware struct {
	rr repository.Redis
}

func New(redisRepo repository.Redis) *Middleware {
	return &Middleware{
		rr: redisRepo,
	}
}

func (m *Middleware) IsAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, err := ctx.Request.Cookie("session_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			}

			ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		if c.Expires.After(time.Now()) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		sessionToken := c.Value
		sc, err := m.rr.SessionExists(sessionToken)
		if err != nil {
			ctx.AbortWithStatusJSON(ds.GetHttpStatusCode(err), err.Error())
			return
		}
		if sc.UserID == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("sessionContext", sc)
		ctx.Next()
	}
}
