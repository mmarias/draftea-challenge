package eventbus

import (
	"context"
	"log"
	"sync"
)

// HandlerFunc is the type for functions that handle events.
type HandlerFunc func(ctx context.Context, message []byte)

type Client interface {
	Publish(ctx context.Context, topic string, message []byte) error
	Subscribe(topic string, handler HandlerFunc)
}

// MemoryBus is an in-memory implementation of an event bus for demonstration purposes.
// It implements the publisher.Client interface.
type MemoryBus struct {
	handlers map[string][]HandlerFunc
	mu       sync.RWMutex
}

// New creates a new instance of MemoryBus.
func New() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]HandlerFunc),
	}
}

// Publish sends a message to all registered handlers for a given topic.
// This method makes MemoryBus implement the publisher.Client interface.
func (b *MemoryBus) Publish(ctx context.Context, topic string, message []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, ok := b.handlers[topic]; ok {
		log.Printf("[MemoryBus] Publishing event to topic '%s'", topic)
		for _, handler := range handlers {
			// In a real scenario, you'd likely run this in a goroutine.
			// For simplicity, we run it synchronously.
			go handler(ctx, message)
		}
	} else {
		log.Printf("[MemoryBus] No handlers registered for topic '%s'", topic)
	}
	return nil
}

// Subscribe registers a handler function for a given topic.
func (b *MemoryBus) Subscribe(topic string, handler HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	log.Printf("[MemoryBus] Subscribing a new handler to topic '%s'", topic)
	b.handlers[topic] = append(b.handlers[topic], handler)
}
