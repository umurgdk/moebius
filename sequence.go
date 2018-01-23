package main

import (
	"context"
	"sync"
)

type AsyncSequence struct {
	err    error
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func NewAsyncSequence() *AsyncSequence {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncSequence{nil, ctx, cancel, wg}
}

func (s *AsyncSequence) Add(num int) *AsyncSequence {
	if s.err == nil {
		s.wg.Add(num)
	}

	return s
}

func (s *AsyncSequence) Done() {
	s.wg.Done()
}

func (s *AsyncSequence) Fail(err error) {
	if err == nil {
		return
	}

	s.err = err
	s.cancel()
}

func (s *AsyncSequence) Err() error {
	return s.err
}

func (s *AsyncSequence) Wait() error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-s.ctx.Done():
		return s.err
	case <-done:
		return nil
	}
}
