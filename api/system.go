package api

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"gihub.com/vlean/oneway/config"
	"gihub.com/vlean/oneway/tool/oauth"
	"github.com/BurntSushi/toml"
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
	fs, err := os.Create(fmt.Sprintf("config-%s.toml", time.Now().Format("200102150405")))
	if err != nil {
		return
	}
	defer fs.Close()

	err = toml.NewEncoder(fs).Encode(req)
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
	from := ctx.Query("from")
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
	code := ctx.Query("code")
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
		redirect := ctx.Query("redirect_uri")
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
