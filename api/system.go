package api

import (
	"errors"

	"gihub.com/vlean/oneway/config"
	"github.com/gin-gonic/gin"
	"github.com/go-session/session/v3"
	log "github.com/sirupsen/logrus"
)

func SystemConfig(ctx *gin.Context) (data any, err error) {
	data = config.Global()
	return
}

func Userinfo(ctx *gin.Context) (data any, err error) {
	store, err := session.Start(ctx.Request.Context(), ctx.Writer, ctx.Request)
	if err != nil {
		log.Errorf("session start err: %v", err)
		return
	}

	email, ok := store.Get("email")
	if !ok {
		err = errors.New("not login")
		return
	}
	data = map[string]any{
		"email": email,
	}
	return
}
