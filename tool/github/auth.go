package github

import (
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type AccessTokenReq struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type AccessTokenRep struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type User struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravatarId        string `json:"gravatar_id"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
}

func Email(req AccessTokenReq) (user User, err error) {
	client := resty.New()
	res := AccessTokenRep{}
	rep, err := client.R().SetBody(req).
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

	rep, err = client.R().SetHeader("Accept", "application/json").
		SetResult(&user).
		SetAuthToken(res.AccessToken).
		Get("https://api.github.com/user")
	return
}

