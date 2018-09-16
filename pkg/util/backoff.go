package util

import "time"

type Backoff struct {
	StartMillis, MaxIntervalMillis, MaxMillis int
	c                                         chan bool
	done                                      chan bool
}

func (b *Backoff) Start() <-chan bool {
	b.c = make(chan bool)
	b.done = make(chan bool)
	go b.backoff()
	return b.c
}

func (b *Backoff) Close() error {
	b.done <- true
	return nil
}

func (b *Backoff) backoff() {
	if b.StartMillis == 0 && b.MaxIntervalMillis == 0 {
		b.StartMillis = 100
		b.MaxIntervalMillis = 2000
		b.MaxMillis = 300000
	}

	start := time.Now()
	currentMillis := b.StartMillis

	for {
		writeVal := true
		if time.Now().Sub(start) > (time.Duration(b.MaxMillis) * time.Millisecond) {
			b.c <- false
		}

		select {
		case <-b.done:
			close(b.done)
			close(b.c)
			return
		case b.c <- writeVal:
		}

		time.Sleep(time.Duration(currentMillis) * time.Millisecond)

		currentMillis *= 2
		if currentMillis > b.MaxIntervalMillis {
			currentMillis = b.MaxIntervalMillis
		}
	}
}
