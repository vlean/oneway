package netx

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Conn struct {
	ctx   context.Context
	ws    *websocket.Conn // websocket连接
	Close func()          // 关闭
	read  chan *Msg       // 消息同步
	write chan *Msg
	once  sync.Once
	// 消息计数
	closed bool
}

func NewConn(conn *websocket.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		ctx:   ctx,
		ws:    conn,
		read:  make(chan *Msg),
		write: make(chan *Msg),
		once:  sync.Once{},
	}
	c.Close = func() {
		c.once.Do(func() {
			c.closed = true
			cancel()
			c.ws.Close()
		})
	}
	go c.readMsg()
	go c.writeMsg()
	return c
}

func (c *Conn) Closed() bool {
	return c.closed
}

func (c *Conn) String() string {
	return c.ws.RemoteAddr().String()
}

func (c *Conn) Write(msg *Msg) {
	c.write <- msg
}

func (c *Conn) Context() context.Context {
	return c.ctx
}

// Pull 从Conn里读取信息
func (c *Conn) Read() *Msg {
	return <-c.read
}

func (c *Conn) ReadC() <-chan *Msg {
	return c.read
}

func (c *Conn) readMsg() {
	defer c.Close()
	var err error
	for {
		msg := &Msg{}
		msg.Type, msg.Cont, err = c.ws.ReadMessage()
		if err != nil {
			log.WithError(err).Debugf("read message fail client: %v type: %v", c.String(), msg.Type)
			break
		}
		msg.TracerRead()
		c.read <- msg
	}
}

func (c *Conn) writeMsg() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Close()
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.write:
			msg.TracerWrite()
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				log.Errorf("write msg fail, msg not ok")
				return
			}

			w, err := c.ws.NextWriter(msg.Type)
			if err != nil {
				log.WithError(err).Errorf("write msg fail build writer client: %v", c.String())
				return
			}
			w.Write(msg.Cont)
			if err = w.Close(); err != nil {
				log.WithError(err).Errorf("write msg fail close err client: %v", c.String())
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
