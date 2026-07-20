package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	notifapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/domain"
)

// fakeRepo is an in-memory Repository for unit tests.
type fakeRepo struct{ items map[string]*domain.Notification }

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[string]*domain.Notification{}} }

func (r *fakeRepo) Save(_ context.Context, n *domain.Notification) error {
	r.items[n.ID] = n
	return nil
}
func (r *fakeRepo) ListQueued(_ context.Context, limit int) ([]*domain.Notification, error) {
	var out []*domain.Notification
	for _, n := range r.items {
		if n.Status == domain.StatusQueued && len(out) < limit {
			out = append(out, n)
		}
	}
	return out, nil
}

// fakeSender records or fails sends.
type fakeSender struct {
	sent []string
	fail bool
}

func (s *fakeSender) SendSMS(_ context.Context, to, _ string) error {
	if s.fail {
		return errors.New("gateway down")
	}
	s.sent = append(s.sent, to)
	return nil
}

func logger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestEnqueueSkipsEmptyRecipient(t *testing.T) {
	repo := newFakeRepo()
	svc := notifapp.NewService(repo, &fakeSender{}, logger())
	if err := svc.EnqueueSMS(context.Background(), "eic", "", "hi"); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if len(repo.items) != 0 {
		t.Fatal("empty recipient must not be queued")
	}
}

func TestDispatchSendsAndMarksSent(t *testing.T) {
	repo := newFakeRepo()
	sender := &fakeSender{}
	svc := notifapp.NewService(repo, sender, logger())
	_ = svc.EnqueueSMS(context.Background(), "eic", "+251911", "policy issued")

	n, err := svc.Dispatch(context.Background(), 10)
	if err != nil || n != 1 {
		t.Fatalf("dispatch = %d (%v), want 1", n, err)
	}
	if len(sender.sent) != 1 || sender.sent[0] != "+251911" {
		t.Fatalf("sender got %v", sender.sent)
	}
	// Marked SENT → not dispatched again.
	n2, _ := svc.Dispatch(context.Background(), 10)
	if n2 != 0 {
		t.Fatalf("second dispatch sent %d, want 0", n2)
	}
}

func TestDispatchMarksFailedOnError(t *testing.T) {
	repo := newFakeRepo()
	svc := notifapp.NewService(repo, &fakeSender{fail: true}, logger())
	_ = svc.EnqueueSMS(context.Background(), "eic", "+251911", "x")

	n, _ := svc.Dispatch(context.Background(), 10)
	if n != 0 {
		t.Fatalf("failed send should count 0 sent, got %d", n)
	}
	for _, item := range repo.items {
		if item.Status != domain.StatusFailed || item.Attempts != 1 {
			t.Fatalf("expected FAILED with 1 attempt, got %s/%d", item.Status, item.Attempts)
		}
	}
}
