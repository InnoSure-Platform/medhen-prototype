package eventbus

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type testEvent struct{ name string }

func (e testEvent) EventName() string { return e.name }

func TestPublishDeliversToAllSubscribers(t *testing.T) {
	b := New(nil)
	var got int
	b.Subscribe("policy.bound", func(context.Context, Event) error { got++; return nil })
	b.Subscribe("policy.bound", func(context.Context, Event) error { got++; return nil })

	if err := b.Publish(context.Background(), testEvent{"policy.bound"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got != 2 {
		t.Fatalf("handlers invoked = %d, want 2", got)
	}
}

func TestPublishRoutesByName(t *testing.T) {
	b := New(nil)
	var a, c int
	b.Subscribe("a", func(context.Context, Event) error { a++; return nil })
	b.Subscribe("c", func(context.Context, Event) error { c++; return nil })

	_ = b.Publish(context.Background(), testEvent{"a"})
	if a != 1 || c != 0 {
		t.Fatalf("routing wrong: a=%d c=%d", a, c)
	}
}

func TestPublishAggregatesErrorsAndIsolatesHandlers(t *testing.T) {
	b := New(nil)
	sentinel := errors.New("boom")
	var third bool
	b.Subscribe("e", func(context.Context, Event) error { return sentinel })
	b.Subscribe("e", func(context.Context, Event) error { panic("bad handler") })
	b.Subscribe("e", func(context.Context, Event) error { third = true; return nil })

	err := b.Publish(context.Background(), testEvent{"e"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected aggregated error to contain sentinel, got %v", err)
	}
	if !third {
		t.Fatal("third handler must still run after a prior error/panic")
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	if err := New(nil).Publish(context.Background(), testEvent{"none"}); err != nil {
		t.Fatalf("publish with no subscribers should be nil, got %v", err)
	}
}

func TestConcurrentSubscribeAndPublish(t *testing.T) {
	b := New(nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Subscribe("x", func(context.Context, Event) error { return nil })
			_ = b.Publish(context.Background(), testEvent{"x"})
		}()
	}
	wg.Wait() // -race will catch data races here
}
