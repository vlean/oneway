package netx

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

var (
	prefix = []byte("CCMAX")
)

func ParseSystem(msg []byte) *PoolCtl {
	if len(msg) < len(prefix) {
		return nil
	}
	for i, b := range prefix {
		if msg[i] != b {
			return nil
		}
	}
	m := msg[len(prefix):]
	p := &PoolCtl{}
	json.Unmarshal(m, p)
	return p
}

// 传输协议
// |  4(协议)+4(类型)| 32b(内容长度)  | xxx内容 |

type MessageVersion int
type MessageType int

const (
	TypeServerCall MessageType = iota
	TypeForwardCall
)
const (
	VersionV1 = 1
)

const (
	CtlPool = iota + 11
	CtlRestart
	CtlStop
)

type PoolCtl struct {
	Want int `json:"want"`
}

type IServerCall[T any] struct {
	Type int `json:"type"`
	Body T   `json:"body"`
}

func NewServerCall[T any](tp int, body T) *IServerCall[T] {
	return &IServerCall[T]{
		Type: tp,
		Body: body,
	}
}

type ServerCall struct {
	Type int             `json:"type"`
	Body json.RawMessage `json:"body"`
}

func (s *ServerCall) HandleCall() {
	switch s.Type {
	case CtlPool:

	}
}

func (s *ServerCall) CtlPool() *PoolCtl {
	data := &PoolCtl{}
	err := json.Unmarshal(s.Body, data)
	if err != nil {
		log.Errorf("parse poolctl err:%v", err)
		return nil
	}
	return data
}
