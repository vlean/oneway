package stat

import (
	"sync/atomic"
	"time"
)

type System struct {
	StartAt time.Time `json:"start_at"` // 启动时间
	Http    *Http     `json:"http"`
}

type Http struct {
	Request  int64 `json:"request"`   // 请求次数
	BodySize int64 `json:"body_size"` // 请求体
	AuthFail int64 `json:"auth_fail"` // 鉴权
}

var _sys *System

func init() {
	_sys = &System{
		StartAt: time.Now(),
		Http:    &Http{},
	}
}

func Runtime() *System {
	return _sys
}

const (
	Request = iota
	Body
	AuthFail
)

func HttpRecord(name int, step int64) {
	switch name {
	case Request:
		atomic.AddInt64(&_sys.Http.Request, step)
	case Body:
		atomic.AddInt64(&_sys.Http.BodySize, step)
	case AuthFail:
		atomic.AddInt64(&_sys.Http.AuthFail, step)
	}

}

func HttpIncr(name int) {
	HttpRecord(name, 1)
}
