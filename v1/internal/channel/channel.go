package channel

import "sync"

func New[T any](options ...func(*Channel[T])) *Channel[T] {
	ch := &Channel[T]{c: make(chan T)}
	for _, o := range options {
		o(ch)
	}

	return ch
}

// Channel provides methods for safely reading from and writing to channels.
type Channel[T any] struct {
	mu     sync.Mutex
	c      chan T
	closed bool
}

// WithBuffer sets the buffer size of the channel.
func WithBuffer[T any](i int) func(*Channel[T]) {
	return func(s *Channel[T]) {
		s.c = make(chan T, i)
	}
}

// Closes the channel. If the channel is already closed, then this is a no-op.
func (c *Channel[T]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		close(c.c)
		c.closed = true
	}
}

// Sends a value to the channel. If the channel is closed, then this is a no-op.
func (c *Channel[T]) Send(t T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.c <- t
}

// Recv returns a read-only channel.
func (c *Channel[T]) Recv() <-chan T {
	return c.c
}
