package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"gihub.com/vlean/oneway/config"
	"gihub.com/vlean/oneway/gox"
	"gihub.com/vlean/oneway/netx"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := &client{
				App: config.Global(),
				cli: &http.Client{
					Transport: http.DefaultTransport,
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				},
			}
			c.buildConn()
			c.buildConn()
			c.Run()
			return nil
		},
	}
	root.AddCommand(cmd)
}

type client struct {
	*config.App
	cli *http.Client
}

func (c *client) Run() {
	gox.Retry(func() error {
		conn, err := c.buildConn()
		if err != nil {
			return err
		}
		log.Infof("connect success %v", conn.String())

		<-conn.Context().Done()
		return nil
	}, gox.RetryAlways())
}

func (c *client) buildConn() (conn *netx.Conn, err error) {
	schema := lo.If(c.StrictMode(), "wss").Else("ws")
	remote := fmt.Sprintf("%s://%s/connect?name=%s&token=%s", schema, c.Client.Remote, c.Client.Name, c.System.Token)
	log.Tracef("remote_connect %v", remote)
	ws, _, err := websocket.DefaultDialer.Dial(remote, nil)
	if err != nil {
		return
	}
	conn = netx.NewConn(ws)
	gox.Run(func() {
		c.handle(conn)
	})
	return
}

func (c *client) handle(conn *netx.Conn) {
	for {
		select {
		case <-conn.Context().Done():
			return
		case msg := <-conn.ReadC():
			c.handleMsg(msg, conn)
		}
	}
}

func (c *client) handleMsg(msg *netx.Msg, conn *netx.Conn) {
	if s := netx.ParseSystem(msg.Cont); s != nil {
		for i := 0; i < s.Want; i++ {
			c.buildConn()
		}
		return
	}
	if err := c.transport(msg.Cont, conn); err != nil {
		log.Errorf("transport err:%v", err)
	}
}

var (
	headers = []string{"Cookie", "User-Agent"}
)

func (c *client) wsproxy(req *http.Request, conn *netx.Conn) (err error) {
	req.URL.Scheme = "ws"
	if req.Header.Get("proxy_schema") == "https" {
		req.URL.Scheme = "wss"
	}
	h := http.Header{}
	for _, k := range headers {
		if v := req.Header.Get(k); v != "" {
			h.Add(k, v)
		}
	}
	// 白名单header
	ws, _, err := websocket.DefaultDialer.Dial(req.URL.String(), h)
	if err != nil {
		return
	}
	to := netx.NewConn(ws)
	gox.Run(func() {
		defer to.Close()
		for {
			select {
			case toMsg := <-to.ReadC():
				conn.Write(toMsg)
			case fMsg := <-conn.ReadC():
				to.Write(fMsg)
			}
		}
	})
	return nil
}

func (c *client) transport(cont []byte, conn *netx.Conn) (err error) {
	reader := bufio.NewReaderSize(bytes.NewReader(cont), len(cont))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Errorf("forward body err: %v", err)
		return
	}
	req.URL.Host = req.Host
	req.URL.Scheme = req.Header.Get("proxy_schema")
	if req.URL.Scheme == "" {
		req.URL.Scheme = "https"
	}
	req.RequestURI = ""
	// 判断是否升级为wss
	if req.Header.Get("Connection") == "Upgrade" &&
		req.Header.Get("Upgrade") == "websocket" {
		_ = c.wsproxy(req, conn)
		return
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		log.Errorf("query body err: %v", err)
		return
	}
	bf := &bytes.Buffer{}
	resp.TransferEncoding = nil
	if err = resp.Write(bf); err != nil {
		return
	}
	log.Tracef("http_redirect url:%v resp: %v", req.URL.String(), bf.Len())
	conn.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})
	return
}
