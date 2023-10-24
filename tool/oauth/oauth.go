package oauth

import (
	"net/url"

	"gihub.com/vlean/oneway/config"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

type OAuth interface {
	AuthURL(to *url.URL) string
	User(code string) (User, error)
}

type User struct {
	Login     string `json:"login"`
	Id        int    `json:"id"`
	AvatarUrl string `json:"avatar_url"`
	Email     string `json:"email"`
}

type base struct {
	cli  *resty.Client
	cfg  *config.App
	auth string
}

func (b *base) AuthURL(to *url.URL) string {
	redirect := lo.If(b.cfg.StrictMode(), "https://").Else("http://") + b.cfg.System.Domain + "/auth/callback"
	ru, _ := url.Parse(redirect)
	q2 := ru.Query()
	q2.Add("redirect_uri", to.String())
	ru.RawQuery = q2.Encode()

	u, _ := url.Parse(b.auth)
	q := u.Query()
	q.Add("client_id", b.cfg.Auth.ClientId)
	q.Add("redirect_uri", ru.String())

	u.RawQuery = q.Encode()
	return u.String()
}

func NewClient(cfg *config.App) OAuth {
	b := &base{
		cli: resty.New(),
		cfg: cfg,
	}
	switch cfg.Auth.Mode {
	case "github":
		b.auth = "https://github.com/login/oauth/authorize?scope=user:email"
		return &github{base: b}
	case "gitee":
		b.auth = "https://gitee.com/oauth/authorize?response_type=code&scope=user_info%20emails"
		return &gitee{base: b}
	default:
		panic("not found auth")
	}
}
