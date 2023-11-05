package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gihub.com/vlean/oneway/api"
	"gihub.com/vlean/oneway/config"
	"gihub.com/vlean/oneway/gox"
	"gihub.com/vlean/oneway/model"
	"gihub.com/vlean/oneway/netx"
	"gihub.com/vlean/oneway/netx/httpx"
	"gihub.com/vlean/oneway/tool/oauth"
	"gihub.com/vlean/oneway/tool/stat"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/gin-gonic/gin"
	"github.com/go-session/session/v3"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(&cobra.Command{
		Use:     "server",
		Aliases: []string{"server"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &server{
				App:    config.Global(),
				pg:     netx.NewGroupPool(),
				oauth:  oauth.NewClient(config.Global()),
				engine: gin.Default(),
			}
			if s.App.System.Token == "" {
				s.App.System.Token = config.Token()
			}
			netx.SetGloablGP(s.pg)
			return s.Run()
		},
	})
}

type server struct {
	*config.App
	server *http.Server
	pg     *netx.GroupPool
	oauth  oauth.OAuth
	engine *gin.Engine
}

//go:embed bin/dist
var feResource embed.FS

func (s *server) Run() (err error) {
	model.InitDB()

	// session
	session.InitManager(
		session.SetDomain(s.RootDomain()),
		session.SetEnableSetCookie(true),
		session.SetExpired(3600*s.App.Auth.Expire),
		session.SetSecure(s.App.StrictMode()),
		session.SetStore(model.NewSessionManager()),
	)

	// server
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.System.Port),
		Handler: s.router(),
	}
	// tls config
	if s.StrictMode() {
		dns := s.App.Cloudflare
		_ = os.Setenv("CLOUDFLARE_EMAIL", dns.Email)
		_ = os.Setenv("CLOUDFLARE_API_KEY", dns.ApiKey)
		_ = os.Setenv("CLOUDFLARE_DNS_API_TOKEN", dns.DnsApiToken)
		_ = os.Setenv("CLOUDFLARE_ZONE_API_TOKEN", dns.ZoneApiToken)
		cfg := simplecert.Default
		cfg.Domains = append([]string{s.App.System.Domain}, model.NewForwardDao().Domains()...)
		cfg.SSLEmail = dns.Email
		cfg.DNSProvider = "cloudflare"
		certReload, err2 := simplecert.Init(cfg, nil)
		if err2 != nil {
			return err2
		}
		tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)
		tlsConf.GetCertificate = certReload.GetCertificateFunc()
		s.server.TLSConfig = tlsConf
		gox.Run(func() {
			s.redirectHttps()
		})
		log.Infof("http server listen %v", s.server.Addr)
		return s.server.ListenAndServeTLS("", "")
	}
	log.Infof("http server listen %v", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *server) redirectHttps() {
	mx := &http.ServeMux{}
	mx.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 拦截
		if !s.canProxy(w, r) {
			return
		}
		r.URL.Host = r.Host
		log.Tracef("https force redirect from %v", r.URL.String())
		switch r.URL.Scheme {
		case "ws":
			r.URL.Scheme = "wss"
		case "":
			if r.Header.Get("Connection") == "Upgrade" &&
				r.Header.Get("Upgrade") == "websocket" {
				r.URL.Scheme = "wss"
			} else {
				r.URL.Scheme = "https"
			}
		default:
			r.URL.Scheme = "https"
		}
		log.Tracef("https force redirect to %v", r.URL.String())
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	})
	log.Info("http force http server listen :80")
	_ = http.ListenAndServe(":80", mx)
}

func (s *server) router() http.Handler {
	mx := &http.ServeMux{}
	mx.HandleFunc("/", s.handle)
	mx.HandleFunc(s.App.System.Domain+"/connect", s.connect)
	mx.HandleFunc(s.App.System.Domain+"/auth/callback", s.callback)
	api.Register(s.engine)
	return mx
}

func (s *server) callback(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = r.Host
	if r.TLS != nil && r.URL.Scheme == "" {
		r.URL.Scheme = "https"
	}

	q := r.URL.Query()
	code := q.Get("code")
	res, err := s.oauth.User(code, r)
	if err != nil {
		log.Errorf("oauth fail err: %v", err)
		return
	}
	stroe, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Errorf("session start err: %v", err)
		return
	}
	if res.Email == "" {
		log.Tracef("proxy user email empty")
		return
	}
	stroe.Set("email", res.Email)
	err = stroe.Save()
	if err != nil {
		log.Errorf("store session err: %v", err)
		return
	}
	if lo.Contains(s.App.Auth.Email, res.Email) {
		redirect := q.Get("redirect_uri")
		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

func (s *server) connect(w http.ResponseWriter, r *http.Request) {
	// 判断是否升级为wss
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		conn, err := netx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		if r.URL.Query().Get("token") != s.App.System.Token {
			log.Warnf("err connection %v", r.URL.String())
			w.WriteHeader(http.StatusForbidden)
			return
		}

		key := r.URL.Query().Get("name")
		if key == "" {
			key = "default"
		}
		pool := s.pg.Get(key)
		pool.Add(conn)
		log.Tracef("connect success group:%s addr:%s size:%v", key, conn.RemoteAddr(), pool.Len())
		return
	}
}

func (s *server) canProxy(w http.ResponseWriter, r *http.Request) bool {
	ok := model.NewForwardDao().Proxy(r.Host) != nil
	if !ok {
		ok = r.Host == s.App.System.Domain
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	return ok
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	// 拦截
	if !s.canProxy(w, r) {
		return
	}
	r.URL.Host = r.Host
	if r.TLS != nil && r.URL.Scheme == "" {
		r.URL.Scheme = "https"
	}
	stat.HttpIncr(stat.Request)
	// 测试

	// 鉴权
	if s.auth(w, r) != nil {
		log.Tracef("auth fail stdout")
		_ = r.Write(os.Stdout)

		stat.HttpIncr(stat.AuthFail)
		r.URL.Host = r.Host
		http.Redirect(w, r, s.oauth.AuthURL(r.URL), http.StatusTemporaryRedirect)
		return
	}

	// 拦截
	if r.Host == s.System.Domain {
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/") {
			s.engine.ServeHTTP(w, r)
			return
		}

		switch path {
		case "/connect":
			s.connect(w, r)
		case "/auth/callback":
			s.callback(w, r)
		default:
			if !(strings.HasSuffix(path, ".js") ||
				strings.HasSuffix(path, ".css") ||
				strings.HasSuffix(path, ".ico")) {
				path = "/index.html"
			}
			cont, err := feResource.ReadFile("bin/dist" + path)
			if err != nil {
				log.Warnf("read file err: %v path: %v", err, path)
				return
			}
			h := w.Header()
			h.Add("Cache-Control", "public, max-age=86400")
			// 增加cache
			_, _ = w.Write(cont)
		}
		return
	}

	// 判断是否升级为wss
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		conn, err := netx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = s.wsproxy(w, r, netx.NewConn(conn))
		return
	}

	// 转发请求
	if err := s.proxy(w, r); err != nil {
		log.Errorf("proxy error %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func (s *server) auth(w http.ResponseWriter, r *http.Request) (err error) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		return err
	}

	email, ok := store.Get("email")
	if !ok {
		return errors.New("not found email")
	}
	if lo.Contains(s.App.Auth.Email, email.(string)) {
		return nil
	}
	return errors.New("not authed")
}

func (s *server) wsproxy(w http.ResponseWriter, r *http.Request, conn *netx.Conn) (err error) {
	fw, nr, ok := s.rewrite(r)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	pool := s.pg.Get(fw.Client)
	if pool == nil {
		return
	}
	pc := pool.Get()
	if pc == nil {
		return
	}
	// build conn
	bf := &bytes.Buffer{}
	if err = nr.Write(bf); err != nil {
		return
	}
	pc.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})

	// read&write
	gox.Run(func() {
		defer pool.Put(pc)
		defer conn.Close()
		for {
			select {
			case orgMsg := <-conn.ReadC():
				pc.Write(orgMsg)
			case toMsg := <-pc.ReadC():
				conn.Write(toMsg)
			}
		}
	})
	return nil
}

func (s *server) proxy(w http.ResponseWriter, r *http.Request) (err error) {
	fw, nr, ok := s.rewrite(r)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if nr.Header.Get("Content-Encoding") != "" {
		nr.Header.Set("Content-Encoding", "gzip")
	}
	//if ck, err := nr.Cookie("go_session_id"); err == nil && ck != nil {
	//	ss := nr.Header.Get("go_session_id")
	//	nr.Header.Set("go_session_id", strings.ReplaceAll(ss, ck.Value, "session_id"))
	//}

	// proxy
	pool := s.pg.Get(fw.Client)
	if pool == nil {
		return
	}

	conn := pool.Get()
	if conn == nil {
		return
	}
	defer pool.Put(conn)

	bf := &bytes.Buffer{}
	if err = nr.Write(bf); err != nil {
		return
	}
	conn.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})

	toMsg := conn.Read()
	bff := &bytes.Buffer{}
	bff.Write(toMsg.Cont)
	resp, err := httpx.ReadResponse(bufio.NewReader(bff))
	if err != nil {
		log.Errorf("parser response err: %v", err)
		return
	}
	log.Tracef("redirect length %v from %v to %v ", resp.Body.Len(),
		nr.URL.Host,
		nr.URL.EscapedPath())
	h := w.Header()
	resp.Header.Del("Connection")
	log.Tracef("tracer header %v", resp.Header)
	for k, v := range resp.Header {
		for _, v1 := range v {
			if strings.Contains(v1, fw.To) {
				v1 = strings.ReplaceAll(v1, "http://"+fw.To, "https://"+fw.From)
				v1 = strings.ReplaceAll(v1, fw.To, fw.From)
			}
			h.Add(k, v1)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	return
}

func (s *server) copyHeader(dest, src http.Header) {
	//copy header
	for k, v := range src {
		for _, v1 := range v {
			dest.Add(k, v1)
		}
	}
}

func (s *server) rewrite(r *http.Request) (p *model.Forward, nr *http.Request, ok bool) {
	// rewrite
	p = model.NewForwardDao().Proxy(r.Host)
	if ok = p != nil; !ok {
		return
	}
	nr = r.Clone(context.Background())
	nr.RequestURI = ""
	nr.Header.Add("proxy_schema", p.Schema)
	nr.Host = p.To
	nr.URL.Host = p.To

	return
}
