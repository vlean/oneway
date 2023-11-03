package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/log"
	"github.com/samber/lo"
)

type App struct {
	System     System     `toml:"system"`
	Client     Client     `toml:"client"`
	Cloudflare Cloudflare `toml:"dns"`
	Auth       Auth       `toml:"auth"`
	Server     Server     `toml:"server"`
}

func (a *App) StrictMode() bool {
	return a.System.Mode == "" || a.System.Mode == "strict"
}

func (a *App) RootDomain() string {
	ss := strings.Split(a.System.Domain, ".")
	return "." + strings.Join(ss[1:], ".")
}

func (a *App) GatewayDomain() string {
	u := "http"
	if a.StrictMode() {
		u = "https"
	}
	u += "://" + a.System.Domain
	return u
}

func (a *App) AuthUrl(to *url.URL) string {
	redirect := lo.If(a.StrictMode(), "https://").Else("http://") + a.System.Domain + "/auth/callback"
	ru, _ := url.Parse(redirect)
	q2 := ru.Query()
	q2.Add("redirect_uri", to.String())
	ru.RawQuery = q2.Encode()

	u, _ := url.Parse("https://github.com/login/oauth/authorize?scope=user:email")
	q := u.Query()
	q.Add("client_id", a.Auth.ClientId)
	q.Add("redirect_uri", ru.String())

	u.RawQuery = q.Encode()
	return u.String()
}

type System struct {
	Host   string `toml:"host"`
	Port   int    `toml:"port"`
	Domain string `toml:"domain"`
	Mode   string `toml:"mode"`  // 模式 strict严格https默认
	Token  string `toml:"token"` // 验证token
}

type Client struct {
	Remote string `toml:"remote"`
	Name   string `toml:"name"`
}

type Server struct {
	Domain     string `toml:"domain"`
	ForceHttps bool   `toml:"force_https"`
}

type Cloudflare struct {
	Email        string `toml:"email"`
	ApiKey       string `toml:"api_key"`
	DnsApiToken  string `toml:"dns_api_token"`
	ZoneApiToken string `toml:"zone_api_token"`
}

type Auth struct {
	Mode     string   `toml:"mode"`   // 默认github
	Expire   int64    `toml:"expire"` // 过期时间 h
	Email    []string `toml:"email"`  // 允许邮箱
	Token    string   `toml:"token"`
	ClientId string   `toml:"client_id"`
}

var (
	_global *App
)

func init() {
	// 获取
	name, _ := os.Hostname()
	_global = &App{
		System: System{
			Host: "0.0.0.0",
			Port: 443,
			Mode: "strict",
		},
		Auth: Auth{
			Mode:   "github",
			Expire: 24,
		},
		Server: Server{
			ForceHttps: true,
		},
		Client: Client{
			Name: name,
		},
	}
}

func Global() *App {
	return _global
}

func Token() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}
	hash := sha256.New()
	hash.Write(randomBytes)
	tokenH := hash.Sum(nil)
	tk := hex.EncodeToString(tokenH)
	log.Infof("init connect token: %s", tk)
	return tk
}
