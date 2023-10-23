package netx

import (
	"bytes"
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

const (
	StageVersion MessageVersion = iota + 1
	StageLength
	StageContent
	StageFin
)

type Message struct {
	Version MessageVersion // 版本
	Type    MessageType
	Length  uint
	Content any
	stage   MessageVersion // 1=> version 2
	raw     []byte
	conn    *Conn
}

func NewSendMessage(tp MessageType, body any) *Message {
	bd, ok := body.([]byte)
	if !ok {
		var err error
		bd, err = json.Marshal(body)
		if err != nil {
			return nil
		}
	}

	raw := &bytes.Buffer{}
	raw.WriteByte(byte(VersionV1<<4 | tp))
	l := len(bd)
	for i := 3; i >= 0; i-- {
		raw.WriteByte(byte(l >> (8 * i)))
	}
	raw.Write(bd)
	return &Message{
		Version: VersionV1,
		Type:    tp,
		Length:  uint(len(bd)),
		Content: body,
		stage:   StageFin,
		raw:     raw.Bytes(),
	}
}
func NewSendHeader(tp MessageType, l int) *Message {
	raw := &bytes.Buffer{}
	raw.WriteByte(byte(VersionV1<<4 | tp))
	for i := 3; i >= 0; i-- {
		raw.WriteByte(byte(l >> (8 * i)))
	}
	return &Message{
		Version: VersionV1,
		Type:    tp,
		Length:  uint(l),
		stage:   StageContent,
		raw:     raw.Bytes(),
	}
}

func (msg *Message) BodySlice() []byte {
	return msg.raw[5:]
}

func (msg *Message) Read(cont []byte) (tail []byte, ok bool) {
	if msg.raw == nil {
		msg.raw = append([]byte{}, cont...)
	} else {
		msg.raw = append(msg.raw, cont...)
	}
	length := len(msg.raw)
	if msg.stage < StageVersion && length > 0 {
		msg.stage++
		msg.Version = MessageVersion(msg.raw[0] >> 4)
		msg.Type = MessageType(msg.raw[0] & 15)
	}
	if msg.stage < StageLength && length >= 5 {
		msg.Length = uint(msg.raw[1])<<24 + uint(msg.raw[2])<<16 + uint(msg.raw[3])<<8 + uint(msg.raw[4])
		msg.stage++
	}
	if msg.stage < StageContent && uint(length) >= 5+msg.Length {
		msg.stage = StageFin
		msg.Content = msg.raw[5 : 5+msg.Length]
		tail = msg.raw[5+msg.Length:]
		msg.raw = msg.raw[:5+msg.Length]
	}
	return tail, msg.stage == StageFin
}

func (msg *Message) Body() any {
	return msg.Content
}

func (msg *Message) RawBody() []byte {
	return msg.raw
}

func (msg *Message) BodyLength() int {
	return len(msg.raw) - 5
}

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
	msg  *Message        `json:"-"`
}

func (s *ServerCall) SetMsg(msg *Message) {
	s.msg = msg
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

