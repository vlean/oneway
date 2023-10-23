package gox

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type RetryOption func(r *retry)

type retry struct {
	max      int
	sleep    time.Duration
	maxSleep time.Duration
	onlyErr  bool
}

func RetryAlways() RetryOption {
	return func(r *retry) {
		r.max = -1
	}
}

func RetryTime(t int) RetryOption {
	return func(r *retry) {
		r.max = t
	}
}

func RetrySleep(start, max time.Duration) RetryOption {
	return func(r *retry) {
		r.sleep = start
		r.maxSleep = max
	}
}

func RetryOnlyErr(b bool) RetryOption {
	return func(r *retry) {
		r.onlyErr = b
	}
}

func Retry(f func() error, opt ...RetryOption) {
	r := &retry{
		max:      3,
		sleep:    time.Second,
		maxSleep: time.Minute,
	}

	for _, o := range opt {
		o(r)
	}

	sleep := r.sleep
	rt := 0
	for {
		err := f()
		if err == nil && r.onlyErr {
			return
		}

		log.Warnf("exec error: %v, retry: %d, sleep: %s", err, rt, sleep.String())

		// 是否停止重试
		if r.max != -1 && rt > r.max {
			return
		}

		time.Sleep(sleep)

		rt++
		sleep *= 2 // 倍增
		if sleep > r.maxSleep {
			sleep = r.maxSleep
		}
	}
}

