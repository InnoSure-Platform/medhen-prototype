package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/notification/domain"
)

func TestNewSMS(t *testing.T) {
	n := domain.NewSMS("eic", "+251911", "hello")
	if n.Channel != domain.ChannelSMS || n.Status != domain.StatusQueued || n.ID == "" {
		t.Fatalf("unexpected notification: %+v", n)
	}
}

func TestMarkSent(t *testing.T) {
	n := domain.NewSMS("eic", "+251911", "hi")
	n.MarkSent()
	if n.Status != domain.StatusSent || n.Attempts != 1 || n.SentAt == nil {
		t.Fatalf("unexpected after MarkSent: %+v", n)
	}
}

func TestMarkFailed(t *testing.T) {
	n := domain.NewSMS("eic", "+251911", "hi")
	n.MarkFailed()
	if n.Status != domain.StatusFailed || n.Attempts != 1 {
		t.Fatalf("unexpected after MarkFailed: %+v", n)
	}
}
