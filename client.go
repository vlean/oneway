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

func (c *client) wsproxy(req *http.Request, proxyConn *netx.Conn) (err error) {
	// 白名单header
	ws, resp, err := websocket.DefaultDialer.Dial(req.URL.String(), req.Header)
	log.Infof("build connection err: %v ws: %v", err, ws)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// ws proxy
	bf := &bytes.Buffer{}
	resp.TransferEncoding = nil
	if err = resp.Write(bf); err != nil {
		return
	}
	log.Tracef("ws proxy url:%v resp: %v", req.URL.String(), bf.Len())
	proxyConn.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})

	cliConn := netx.NewConn(ws)
	gox.Run(func() {
		defer cliConn.Close()
		for v := range cliConn.ReadC() {
			proxyConn.Write(v)
		}
	})
	gox.Run(func() {
		defer cliConn.Close()
		for v := range proxyConn.ReadC() {
			cliConn.Write(v)
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
	if lo.Contains([]string{"ws", "wss"}, req.Header.Get("proxy_schema")) {
		err = c.wsproxy(req, conn)
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
