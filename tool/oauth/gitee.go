package oauth

import (
	"net/http"
	"net/url"
)

type gitee struct {
	*base
}

type giteeAuthRep struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (g *gitee) User(code string, r *http.Request) (user User, err error) {
	ru, _ := url.Parse(r.URL.String())
	if ru.Scheme == "" {
		ru.Scheme = "https"
	}
	q := ru.Query()
	q.Del("code")
	ru.RawQuery = q.Encode()

	authRep := giteeAuthRep{}
	_, err = g.cli.SetDebug(true).R().
		SetResult(&authRep).
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     g.cfg.Auth.ClientId,
			"client_secret": g.cfg.Auth.Token,
			"code":          code,
			"redirect_uri":  ru.String(),
		}).Post("https://gitee.com/oauth/token")
	if err != nil || authRep.AccessToken == "" {
		return
	}
	users := make([]User, 0)
	_, err = g.cli.SetDebug(true).R().
		SetResult(&users).
		SetQueryParam("", authRep.AccessToken).
		Get("https://gitee.com/api/v5/emails")
	if err != nil || len(users) == 0 {
		return
	}
	user.Email = users[0].Email
	return
}
