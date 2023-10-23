package netx

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"

	"gihub.com/vlean/oneway/gox"
	queue "github.com/enriquebris/goconcurrentqueue"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"golang.org/x/net/context"
)

type Pool struct {
	*sync.RWMutex
	q      *queue.FIFO
	size   int
	extent atomic.Bool
}

func NewPool() *Pool {
	pool := &Pool{
		RWMutex: &sync.RWMutex{},
		q:       queue.NewFIFO(),
	}
	return pool
}

// Add 新增connect
func (c *Pool) Add(conn *websocket.Conn) {
	cx := NewConn(conn)
	c.q.Enqueue(cx)
	c.Lock()
	defer c.Unlock()
	c.size++

	gox.Run(func() {
		<-cx.Context().Done()
		c.Lock()
		defer c.Unlock()
		c.size--
		log.Tracef("conn close addr: %v last: %v", conn.RemoteAddr(), c.size)
	})
}

// Put 将conn放回连接池
func (c *Pool) Put(conn *Conn) {
	c.q.Enqueue(conn)
}

func (c *Pool) Get() (val *Conn) {
	tm, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cx, err := c.q.DequeueOrWaitForNextElementContext(tm)
	if err != nil {
		return nil
	}

	conn := cx.(*Conn)

	// 判断是否需要扩容
	if c.q.GetLen() < 5 && c.size < 1e3 && !c.extent.Load() {
		c.extent.Store(true)
		gox.Run(func() {
			defer func() {
				c.Put(conn)
				time.Sleep(time.Second / 3)
				c.extent.Store(false)
			}()
			conn.Write(&Msg{
				Type: websocket.TextMessage,
				Cont: []byte("CCMAX{\"want\": 5, \"type\": 11}"),
			})
		})
		return c.Get()
	}
	return conn
}

type GroupPool struct {
	*sync.RWMutex
	pool map[string]*Pool
}

func NewGroupPool() *GroupPool {
	return &GroupPool{
		RWMutex: &sync.RWMutex{},
		pool:    make(map[string]*Pool),
	}
}

func (g *GroupPool) Get(key string) *Pool {
	g.RLock()
	defer g.RUnlock()
	_, ok := g.pool[key]
	if !ok {
		g.pool[key] = NewPool()
	}
	return g.pool[key]
}

func (g *GroupPool) Add(group string, pool *Pool) {
	g.Lock()
	defer g.Unlock()
	g.pool[group] = pool
}
