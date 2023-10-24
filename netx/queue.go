package netx

import (
	"context"
	"errors"
)

type Queue[T any] struct {
	data chan T
}

func NewQueue[T any](size int) *Queue[T] {
	return &Queue[T]{
		data: make(chan T, size),
	}
}

func (q *Queue[T]) Enqueue(v T) {
	q.data <- v
}

func (q *Queue[T]) Dequene() (v T, err error) {
	select {
	case v = <-q.data:
	default:
		err = errors.New("not found")
	}
	return
}

func (q *Queue[T]) DequeueOrWaitForNextElementContext(ctx context.Context) (v T, err error) {
	select {
	case <-ctx.Done():
		err = errors.New("not found")
	case v = <-q.data:
	}
	return
}

func (q *Queue[T]) GetLen() int {
	return len(q.data)
}
