package oauth

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type AccessTokenReq struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type github struct {
	*base
}

type AccessTokenRep struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func (g *github) User(code string, r *http.Request) (user User, err error) {
	req := AccessTokenReq{
		ClientId:     g.cfg.Auth.ClientId,
		ClientSecret: g.cfg.Auth.Token,
		Code:         code,
	}
	res := AccessTokenRep{}
	rep, err := g.cli.R().
		SetBody(req).
		SetHeader("Accept", "application/json").
		SetResult(&res).
		Post("https://github.com/login/oauth/access_token")
	log.Tracef("access_token err:%v, resp: %v", err, rep)
	if err != nil {
		return
	}
	if res.AccessToken == "" {
		return
	}

	_, err = g.cli.R().
		SetHeader("Accept", "application/json").
		SetResult(&user).
		SetAuthToken(res.AccessToken).
		Get("https://api.github.com/user")
	return
}
