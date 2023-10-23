package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"

	"gihub.com/vlean/oneway/config"
	"gihub.com/vlean/oneway/gox"
	"gihub.com/vlean/oneway/model"
	"gihub.com/vlean/oneway/netx"
	"gihub.com/vlean/oneway/tool/github"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/go-session/session/v3"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func init() {
	root.AddCommand(&cobra.Command{
		Use:     "server",
		Aliases: []string{"server"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &server{
				App: config.Global(),
				pg:  netx.NewGroupPool(),
			}
			return s.Run()
		},
	})
}

type server struct {
	*config.App
	server *http.Server
	pg     *netx.GroupPool
}

func (s *server) Run() (err error) {
	// session
	session.InitManager(
		session.SetDomain(s.RootDomain()),
		session.SetEnableSetCookie(true),
		session.SetExpired(3600*24),
		session.SetSecure(s.App.StrictMode()),
		session.SetStore(session.NewMemoryStore()),
	)

	// server
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.System.Port),
		Handler: s.router(),
	}
	// tls config
	log.Infof("http server listen %v", s.server.Addr)
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
		return s.server.ListenAndServeTLS("", "")
	}
	return s.server.ListenAndServe()
}

func (s *server) redirectHttps() {
	mx := &http.ServeMux{}
	mx.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "http" {
			r.URL.Scheme = "https"
		} else if r.URL.Scheme == "ws" {
			r.URL.Scheme = "wss"
		}
		http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
	})
	_ = http.ListenAndServe(":80", mx)
}

func (s *server) router() *http.ServeMux {
	mx := &http.ServeMux{}
	mx.HandleFunc("/", s.handle)
	mx.HandleFunc(s.App.System.Domain+"/connect", s.connect)
	mx.HandleFunc(s.App.System.Domain+"/auth/callback", s.callback)
	mx.HandleFunc("/mock", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	return mx
}

func (s *server) callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	res, err := github.Email(github.AccessTokenReq{
		ClientId:     s.Auth.ClientId,
		ClientSecret: s.Auth.Token,
		Code:         code,
	})
	if err != nil {
		return
	}
	stroe, err := session.Start(context.Background(), w, r)
	if err != nil {
		return
	}
	stroe.Set("email", res.Email)
	err = stroe.Save()
	if err != nil {
		return
	}
	if lo.Contains(s.App.Auth.Email, res.Email) {
		redirect := q.Get("redirect_uri")
		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	}
}

func (s *server) connect(w http.ResponseWriter, r *http.Request) {
	// 判断是否升级为wss
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		conn, err := netx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		key := r.Header.Get("name")
		if key == "" {
			key = "default"
		}
		log.Tracef("connect success group:%s addr:%s", key, conn.RemoteAddr())
		pool := s.pg.Get(key)
		pool.Add(conn)
		return
	}
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	if s.auth(w, r) != nil {
		http.Redirect(w, r, s.App.AuthUrl(r.URL), http.StatusTemporaryRedirect)
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
		return
	}

	// 默认处理
	w.WriteHeader(http.StatusForbidden)
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
	group, nr, ok := s.rewrite(r)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	pool := s.pg.Get(group)
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
	conn.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})

	// read&write
	gox.Run(func() {
		defer pool.Put(pc)
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
	group, nr, ok := s.rewrite(r)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// proxy
	pool := s.pg.Get(group)
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
	br := bufio.NewReader(bff)
	response, err := http.ReadResponse(br, nr)
	if err != nil {
		log.Errorf("parser response err: %v", err)
		return
	}
	s.copyHeader(w.Header(), response.Header)
	response.Write(w)
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

func (s *server) rewrite(r *http.Request) (group string, nr *http.Request, ok bool) {
	// rewrite
	p := model.NewForwardDao().Proxy(r.Host)
	if ok = p != nil; !ok {
		return
	}
	nr = r.Clone(context.Background())
	nr.RequestURI = ""
	nr.Header.Add("proxy_schema", p.Schema)
	nr.Host = p.To
	nr.URL.Host = p.To
	group = p.Client

	return
}
