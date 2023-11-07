package api

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gihub.com/vlean/oneway/config"
	"github.com/gin-gonic/gin"
)

var (
	tokenCache = &sync.Map{}
	authFail   = errors.New("fail client_id or host or token or redirect_uri")
)

type AuthReq struct {
	ClientID     string `form::"client_id"`
	ResponseType string `form:"response_type"`
	State        string `form:"state"`
	RedirectUri  string `form:"redirect_uri"`
	Scope        string `form:"scope"`
}

func Auth(ctx *gin.Context) (data any, err error) {
	req, err := Bind[AuthReq](ctx)
	if err != nil {
		return
	}
	if req.ClientID == "" || !strings.HasPrefix(req.ClientID, "oauth-") || req.RedirectUri == "" {
		err = authFail
		return
	}

	toUrl, err := url.Parse(req.RedirectUri)
	if err != nil {
		return
	}

	// 生成随机token
	tk := config.Token()
	if tk == "" {
		err = authFail
		return
	}
	tk = tk[:32]

	// tokenCache.Range(func(key, value any) bool {
	// 	vt, ok := value.(time.Time)
	// 	if !ok {
	// 		return true
	// 	}
	// 	if time.Since(vt) > time.Minute*10 {
	// 		tokenCache.Delete(key)
	// 	}
	// 	return true
	// })
	tokenCache.Store(tk, time.Now())

	q := toUrl.Query()
	q.Add("code", tk)
	q.Add("client_id", req.ClientID)
	if req.State != "" {
		q.Add("state", req.State)
	}
	if req.ResponseType != "" {
		q.Add("resposne_type", req.ResponseType)
	}
	if req.Scope != "" {
		q.Add("scope", req.Scope)
	}

	toUrl.RawQuery = q.Encode()
	ctx.Redirect(http.StatusTemporaryRedirect, toUrl.String())
	return
}

type CodeReq struct {
	Code      string `form:"code"`
	GrantType string `form:"grant_type"`
}

func Code(ctx *gin.Context) {
	req, err := Bind[CodeReq](ctx)
	if err != nil {
		return
	}
	v, ok := tokenCache.LoadAndDelete(req.Code)
	if !ok {
		err = errors.New("invalid token")
		return
	}
	if time.Since(v.(time.Time)) > time.Minute*10 {
		err = errors.New("invalid token")
		return
	}

	tk := config.Token()
	tokenCache.Store(tk, nil)
	data := map[string]any{
		"access_token":  tk,
		"token_type":    "bearer",
		"expires_in":    86400 * 7,
		"refresh_token": tk,
		"email":         "vlean",
		"id":            "vlean",
	}
	ctx.JSON(http.StatusOK, data)
}

func User(ctx *gin.Context) (data any, err error) {

	return
}
