package onetomany

import "sync"

type OneToManyChan[T any] struct {
	chans     map[int]chan T
	bufLen    int
	idCounter int
	mu        sync.RWMutex
}

func (o *OneToManyChan[T]) Init(bufLen int) {
	o.chans = make(map[int]chan T)
	o.bufLen = bufLen
}

func (o *OneToManyChan[T]) Subscribe() (<-chan T, func()) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.idCounter++
	id := o.idCounter

	ch := make(chan T, o.bufLen)
	o.chans[id] = ch

	unsubscribe := func() {
		o.mu.Lock()
		defer o.mu.Unlock()

		delete(o.chans, id)
		close(ch)
	}

	return ch, unsubscribe
}

func (o *OneToManyChan[T]) Push(v T) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, ch := range o.chans {
		ch <- v
	}
}
