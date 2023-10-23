package gox

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func RunE(f func(ctx context.Context) error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// todo
			}
		}()
		if err := f(context.Background()); err != nil {
			log.Errorf("gorountine run error: %v", err)
		}
	}()
}

func Run(f func()) {
	RunE(func(ctx context.Context) error {
		f()
		return nil
	})
}

