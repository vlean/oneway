package oauth

type gitee struct {
	*base
}


type giteeAuthRep struct{
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
}

func (g *gitee) User(code string) (user User, err error) {
	authRep := giteeAuthRep{}
	_, err = g.cli.R().
		SetResult(&authRep).
		SetQueryParams(map[string]string{
			"grant_type":"authorization_code",
			"client_id": g.cfg.Auth.ClientId,
			"client_secret": g.cfg.Auth.Token,
			"code": code,
		}).Get("https://gitee.com/oauth/token")
	if err != nil || authRep.AccessToken == "" {
		return
	}
	users := make([]User, 0)
	// GET ?access_token=e082440a95c0f0667b0d8a4c55fb7d30
	_, err = g.cli.R().
		SetResult(&users). 
		SetQueryParam("", authRep.AccessToken). 
		Get("https://gitee.com/api/v5/emails")
	if err != nil || len(users) == 0 {
		return 
	}
	user.Email = users[0].Email
	return 
}
