package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBroker_Subscribe_ReturnsChannel(t *testing.T) {
	b := New[string]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := b.Subscribe(ctx)
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
}

func TestBroker_Publish_SingleSubscriber(t *testing.T) {
	b := New[string]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := b.Subscribe(ctx)
	b.Publish("hello")

	select {
	case msg := <-ch:
		if msg != "hello" {
			t.Errorf("expected 'hello', got %q", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestBroker_Publish_MultipleSubscribers(t *testing.T) {
	b := New[int]()

	const subscriberCount = 5
	channels := make([]<-chan int, subscriberCount)
	cancels := make([]context.CancelFunc, subscriberCount)

	for i := 0; i < subscriberCount; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancels[i] = cancel
		channels[i] = b.Subscribe(ctx)
	}
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	b.Publish(42)

	for i, ch := range channels {
		select {
		case msg := <-ch:
			if msg != 42 {
				t.Errorf("subscriber %d: expected 42, got %d", i, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("subscriber %d: timeout waiting for event", i)
		}
	}
}

func TestBroker_ContextCancellation(t *testing.T) {
	b := New[string]()
	ctx, cancel := context.WithCancel(context.Background())

	ch := b.Subscribe(ctx)

	if b.SubscriberCount() != 1 {
		t.Errorf("expected 1 subscriber, got %d", b.SubscriberCount())
	}

	cancel()

	// Wait for cleanup goroutine
	time.Sleep(50 * time.Millisecond)

	if b.SubscriberCount() != 0 {
		t.Errorf("expected 0 subscribers after cancel, got %d", b.SubscriberCount())
	}

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel should be closed")
	}
}

func TestBroker_SlowSubscriber(t *testing.T) {
	b := New[int](WithBufferSize[int](2))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := b.Subscribe(ctx)

	// Fill buffer
	b.Publish(1)
	b.Publish(2)

	// This should be dropped (buffer full, non-blocking)
	b.Publish(3)

	// Should only receive first two
	received := make([]int, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch:
			received = append(received, msg)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	}

	if len(received) != 2 || received[0] != 1 || received[1] != 2 {
		t.Errorf("expected [1, 2], got %v", received)
	}

	// No more events should be available
	select {
	case msg := <-ch:
		t.Errorf("unexpected message: %d", msg)
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestBroker_SubscriberCount(t *testing.T) {
	b := New[string]()

	if b.SubscriberCount() != 0 {
		t.Errorf("expected 0, got %d", b.SubscriberCount())
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	b.Subscribe(ctx1)

	if b.SubscriberCount() != 1 {
		t.Errorf("expected 1, got %d", b.SubscriberCount())
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	b.Subscribe(ctx2)

	if b.SubscriberCount() != 2 {
		t.Errorf("expected 2, got %d", b.SubscriberCount())
	}

	cancel1()
	time.Sleep(50 * time.Millisecond)

	if b.SubscriberCount() != 1 {
		t.Errorf("expected 1 after cancel, got %d", b.SubscriberCount())
	}

	cancel2()
	time.Sleep(50 * time.Millisecond)

	if b.SubscriberCount() != 0 {
		t.Errorf("expected 0 after all cancelled, got %d", b.SubscriberCount())
	}
}

func TestBroker_WithBufferSize(t *testing.T) {
	b := New[int](WithBufferSize[int](3))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := b.Subscribe(ctx)

	// Fill buffer (size 3)
	b.Publish(1)
	b.Publish(2)
	b.Publish(3)
	b.Publish(4) // Should be dropped

	received := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		select {
		case msg := <-ch:
			received = append(received, msg)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	}

	if len(received) != 3 {
		t.Errorf("expected 3 messages, got %d", len(received))
	}
}

func TestBroker_ConcurrentPublish(t *testing.T) {
	b := New[int](WithBufferSize[int](1000))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := b.Subscribe(ctx)

	const publishers = 10
	const messagesPerPublisher = 100

	var wg sync.WaitGroup
	wg.Add(publishers)

	for i := 0; i < publishers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerPublisher; j++ {
				b.Publish(id*messagesPerPublisher + j)
			}
		}(i)
	}

	wg.Wait()

	// Drain channel and count
	received := 0
drainLoop:
	for {
		select {
		case <-ch:
			received++
		case <-time.After(100 * time.Millisecond):
			break drainLoop
		}
	}

	// Should receive at least some messages (buffer was large enough)
	if received == 0 {
		t.Error("expected to receive some messages")
	}
}

func TestBroker_ConcurrentSubscribe(t *testing.T) {
	b := New[int]()

	const subscribers = 20
	var wg sync.WaitGroup
	wg.Add(subscribers)

	cancels := make([]context.CancelFunc, subscribers)

	for i := 0; i < subscribers; i++ {
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			cancels[idx] = cancel
			b.Subscribe(ctx)
		}(i)
	}

	wg.Wait()

	if b.SubscriberCount() != subscribers {
		t.Errorf("expected %d subscribers, got %d", subscribers, b.SubscriberCount())
	}

	// Cleanup
	for _, cancel := range cancels {
		if cancel != nil {
			cancel()
		}
	}
}

func TestBroker_ConcurrentPublishSubscribe(t *testing.T) {
	b := New[int](WithBufferSize[int](100))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start subscribers
	const subCount = 5
	wg.Add(subCount)
	for i := 0; i < subCount; i++ {
		go func() {
			defer wg.Done()
			ch := b.Subscribe(ctx)
			for range ch {
				// Consume messages
			}
		}()
	}

	// Wait for subscribers to register
	time.Sleep(50 * time.Millisecond)

	// Publish concurrently
	const pubCount = 5
	const msgCount = 50
	var pubWg sync.WaitGroup
	pubWg.Add(pubCount)
	for i := 0; i < pubCount; i++ {
		go func() {
			defer pubWg.Done()
			for j := 0; j < msgCount; j++ {
				b.Publish(j)
			}
		}()
	}

	pubWg.Wait()
	cancel()
	wg.Wait()

	// If we get here without deadlock or panic, test passed
}
