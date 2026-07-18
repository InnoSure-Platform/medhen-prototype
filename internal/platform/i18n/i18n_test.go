package i18n

import (
	"sync"
	"testing"
)

func TestTranslateAndFallback(t *testing.T) {
	tr := New()
	tr.Add(English, "policy.bound", "Policy bound")
	tr.Add(Amharic, "policy.bound", "ፖሊሲ ተያይዟል")

	if got := tr.Translate(Amharic, "policy.bound"); got != "ፖሊሲ ተያይዟል" {
		t.Fatalf("am = %q", got)
	}
	if got := tr.Translate(English, "policy.bound"); got != "Policy bound" {
		t.Fatalf("en = %q", got)
	}
	// Missing Amharic falls back to English.
	tr.Add(English, "only.en", "English only")
	if got := tr.Translate(Amharic, "only.en"); got != "English only" {
		t.Fatalf("fallback = %q, want English only", got)
	}
	// Missing entirely returns the key.
	if got := tr.Translate(English, "missing"); got != "missing" {
		t.Fatalf("missing = %q, want key echo", got)
	}
}

func TestParseLang(t *testing.T) {
	if ParseLang("am") != Amharic || ParseLang("en") != English || ParseLang("fr") != Default {
		t.Fatal("ParseLang mapping wrong")
	}
}

func TestConcurrentAddTranslate(t *testing.T) {
	tr := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); tr.Add(English, "k", "v") }()
		go func() { defer wg.Done(); _ = tr.Translate(English, "k") }()
	}
	wg.Wait() // -race asserts no data race
}
