package api

import (
	"errors"
	"net/url"

	"gihub.com/vlean/oneway/config"
	"gihub.com/vlean/oneway/tool/oauth"
	"github.com/gin-gonic/gin"
	"github.com/go-session/session/v3"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

func SystemConfig(ctx *gin.Context) (data any, err error) {
	data = config.Global()
	return
}

func SystemConfigUpdate(ctx *gin.Context) (data any, err error) {
	req, err := Bind[config.App](ctx)
	if err != nil {
		return
	}
	log.Tracef("获取config req: %v", req)
	data = &req
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

func OAuthUrl(ctx *gin.Context) (data any, err error) {
	from := ctx.GetString("from")
	cfg := config.Global()
	if from != "" {
		from = cfg.GatewayDomain()
	}
	fromU, err := url.Parse(from)
	if err != nil {
		return
	}

	redirect := oauth.NewClient(config.Global()).AuthURL(fromU)
	data = map[string]string{
		"redirect": redirect,
	}
	return
}

func OAuth(ctx *gin.Context) (data any, err error) {
	code := ctx.GetString("code")
	cfg := config.Global()
	res, err := oauth.NewClient(cfg).User(code, ctx.Request)
	if err != nil {
		log.Errorf("oauth fail err: %v", err)
		return
	}
	stroe, err := session.Start(ctx.Request.Context(), ctx.Writer, ctx.Request)
	if err != nil {
		return
	}
	if res.Email == "" {
		err = errors.New("proxy user email empty")
		return
	}
	stroe.Set("email", res.Email)
	err = stroe.Save()
	if err != nil {
		log.Errorf("store session err: %v", err)
		return
	}
	if lo.Contains(cfg.Auth.Email, res.Email) {
		redirect := ctx.GetString("redirect_uri")
		data = map[string]any{
			"redirect": redirect,
			"auth":     true,
		}
	} else {
		data = map[string]any{
			"auth": false,
		}
	}
	return
}
