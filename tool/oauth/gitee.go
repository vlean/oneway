package oauth

import "github.com/samber/lo"

type gitee struct {
	*base
}

type giteeAuthRep struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (g *gitee) User(code string) (user User, err error) {
	authRep := giteeAuthRep{}
	_, err = g.cli.SetDebug(true).R().
		SetResult(&authRep).
		SetQueryParams(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     g.cfg.Auth.ClientId,
			"client_secret": g.cfg.Auth.Token,
			"code":          code,
			"redirect_uri":  lo.If(g.cfg.StrictMode(), "https://").Else("http://") + g.cfg.System.Domain + "/auth/callback",
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
