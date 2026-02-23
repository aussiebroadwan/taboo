package pubsub

import (
	"context"
	"sync"
)

// Option configures a Broker.
type Option[T any] func(*Broker[T])

// WithBufferSize sets the buffer size for subscriber channels.
func WithBufferSize[T any](size int) Option[T] {
	return func(b *Broker[T]) {
		b.bufferSize = size
	}
}

// Broker is a generic publish/subscribe message broker.
type Broker[T any] struct {
	mu          sync.RWMutex
	subscribers map[chan T]struct{}
	bufferSize  int
}

// New creates a new Broker with the given options.
func New[T any](opts ...Option[T]) *Broker[T] {
	b := &Broker[T]{
		subscribers: make(map[chan T]struct{}),
		bufferSize:  16,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Subscribe returns a channel that receives published events.
// The channel is closed when the context is cancelled.
func (b *Broker[T]) Subscribe(ctx context.Context) <-chan T {
	ch := make(chan T, b.bufferSize)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	// Cleanup when context is cancelled
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers, ch)
		close(ch)
		b.mu.Unlock()
	}()

	return ch
}

// Publish sends an event to all subscribers.
// Events are dropped for slow subscribers (non-blocking).
func (b *Broker[T]) Publish(event T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Drop event if subscriber is slow
		}
	}
}

// SubscriberCount returns the current number of subscribers.
func (b *Broker[T]) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
