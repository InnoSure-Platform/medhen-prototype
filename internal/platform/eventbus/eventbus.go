// Package eventbus is the in-process publish/subscribe backbone for domain
// events. In the modular monolith, a command handler writes events to the
// transactional outbox in the same DB transaction; after commit the relay
// publishes them here, where subscribers (audit, notification, reporting, …)
// handle them. Keeping this seam means the same handler code works unchanged if
// a module is later extracted to its own service (the relay simply targets a
// broker instead of this bus).
package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// Event is a domain event. EventName is the stable topic key used for routing.
type Event interface {
	EventName() string
}

// Handler consumes an event. Returning an error marks the delivery as failed
// (the relay may retry); a panic is recovered and treated as an error so one
// bad subscriber cannot take down the publisher or its siblings.
type Handler func(ctx context.Context, e Event) error

// Bus is a concurrency-safe in-process event dispatcher.
type Bus struct {
	mu          sync.RWMutex
	handlers    map[string][]Handler
	allHandlers []Handler
	logger      *slog.Logger
}

// New creates an empty bus.
func New(logger *slog.Logger) *Bus {
	if logger == nil {
		logger = slog.Default()
	}
	return &Bus{handlers: make(map[string][]Handler), logger: logger}
}

// Subscribe registers a handler for the named event. Multiple handlers per event
// are allowed and are invoked in registration order.
func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// SubscribeAll registers a handler invoked for every published event, regardless
// of topic (used by the audit module to record an immutable trail).
func (b *Bus) SubscribeAll(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.allHandlers = append(b.allHandlers, h)
}

// Publish delivers an event synchronously to all subscribers, aggregating their
// errors. Delivery to one handler never prevents delivery to the others.
func (b *Bus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[e.EventName()]...)
	handlers = append(handlers, b.allHandlers...)
	b.mu.RUnlock()

	var errs []error
	for _, h := range handlers {
		if err := b.safeInvoke(ctx, h, e); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b *Bus) safeInvoke(ctx context.Context, h Handler, e Event) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			b.logger.Error("event handler panic", "event", e.EventName(), "panic", rec)
			err = fmt.Errorf("handler panic on %q: %v", e.EventName(), rec)
		}
	}()
	return h(ctx, e)
}
