package tftp

import (
	"math/rand"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
	defaultRetries = 5
)

type backoffFunc func(int) time.Duration

type backoff struct {
	attempt int
	handler backoffFunc
}

func (b *backoff) reset() {
	b.attempt = 0
}

func (b *backoff) count() int {
	return b.attempt
}

func (b *backoff) backoff() {
	if b.handler == nil {
		time.Sleep(time.Duration(rand.Int63n(int64(time.Second))))
	} else {
		time.Sleep(b.handler(b.attempt))
	}
	b.attempt++
}
