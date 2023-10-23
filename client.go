package main

import (
	"bufio"
	"bytes"
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
	root.AddCommand(&cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := &client{
				App: config.Global(),
			}
			c.Run()
			return nil
		},
	})
}

type client struct {
	*config.App
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
	ws, _, err := websocket.DefaultDialer.Dial(lo.If(c.StrictMode(), "wss://").Else("ws://")+
		c.Client.Remote+"/connect", nil)
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("query body err: %v", err)
		return
	}
	bf := &bytes.Buffer{}
	if err = resp.Write(bf); err != nil {
		return
	}
	log.Tracef("http_redirect url:%v resp: %v", req.URL.String(), bf.String())
	conn.Write(&netx.Msg{
		Type: websocket.TextMessage,
		Cont: bf.Bytes(),
	})
	return
}

