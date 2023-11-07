package netx

import (
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"gihub.com/vlean/oneway/gox"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

type Pool struct {
	*sync.RWMutex
	q      *Queue[*Conn]
	size   int
	extent atomic.Bool
}

func NewPool() *Pool {
	pool := &Pool{
		RWMutex: &sync.RWMutex{},
		q:       NewQueue[*Conn](1e3),
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
func (c *Pool) Len() int {
	c.Lock()
	defer c.Unlock()
	return c.size
}

// Put 将conn放回连接池
func (c *Pool) Put(conn *Conn) {
	c.q.Enqueue(conn)
}

func (c *Pool) Get() (val *Conn) {
	tm, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := c.q.DequeueOrWaitForNextElementContext(tm)
	log.Tracef("fetch conn from pool conn: %v err: %v last: %v", conn, err, c.q.GetLen())
	if err != nil {
		return nil
	}
	if conn.Closed() {
		return c.Get()
	}

	// 判断是否需要扩容
	if c.q.GetLen() < 5 && c.size < 1e3 && !c.extent.Load() {
		log.Tracef("pool need extend size: %d", c.size)
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

func (g *GroupPool) GetConn(key string) *Conn {
	p := g.Get(key)
	if p == nil {
		return nil
	}
	return p.Get()
}

func (g *GroupPool) PutConn(k string, c *Conn) {
	p := g.Get(k)
	if p == nil {
		return
	}
	p.Put(c)
}

func (g *GroupPool) Add(group string, pool *Pool) {
	g.Lock()
	defer g.Unlock()
	g.pool[group] = pool
}

func (g *GroupPool) Stat() []Stat {
	g.Lock()
	defer g.Unlock()
	ret := make([]Stat, 0)
	for name, pool := range g.pool {
		ret = append(ret, Stat{
			Name: name,
			Size: pool.size,
			Use:  pool.q.GetLen(),
		})
	}
	return ret
}

type Stat struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Use  int    `json:"use"`
}

var _gloabl *GroupPool

func SetGloablGP(c *GroupPool) {
	_gloabl = c
}

func GlobalGP() *GroupPool {
	return _gloabl
}
