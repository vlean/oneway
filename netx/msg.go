package netx

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type MsgType int

const (
	MsgSystem  MsgType = 1 // 系统
	MsgForward MsgType = 2 // 转发

)

type Msg struct {
	Type int
	Cont []byte
}

func (m *Msg) System() *PoolCtl {
	pc := &PoolCtl{}
	json.Unmarshal(m.Cont, pc)
	return pc
}

func (m *Msg) TracerWrite(c *Conn) {
	log.Tracef("tracer %s write type:%d cont: %v", c, m.Type, string(m.Cont))
}
func (m *Msg) TracerRead(c *Conn) {
	log.Tracef("tracer %s read type:%d cont: %v", c, m.Type, string(m.Cont))
}
