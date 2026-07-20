// Package i18n provides a small, concurrency-safe translator for the platform's
// bilingual (Amharic / English) content. The pre-refactor translator documented
// thread-safety but had no lock; this one guards its dictionary with an RWMutex.
package i18n

import "sync"

// Lang is a supported language code.
type Lang string

const (
	Amharic Lang = "am"
	English Lang = "en"
	// Default is used when a requested language is unknown.
	Default = English
)

// Translator holds keyed translations per language.
type Translator struct {
	mu       sync.RWMutex
	dict     map[Lang]map[string]string
	fallback Lang
}

// New creates an empty translator falling back to English.
func New() *Translator {
	return &Translator{dict: make(map[Lang]map[string]string), fallback: Default}
}

// Add registers a translation for key in the given language. Safe for concurrent use.
func (t *Translator) Add(lang Lang, key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dict[lang] == nil {
		t.dict[lang] = make(map[string]string)
	}
	t.dict[lang][key] = value
}

// Translate returns the value for key in lang, falling back to the fallback
// language and finally to the key itself. Safe for concurrent use.
func (t *Translator) Translate(lang Lang, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if m, ok := t.dict[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := t.dict[t.fallback]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

// ParseLang maps an Accept-Language-ish value to a supported Lang, defaulting to English.
func ParseLang(s string) Lang {
	switch Lang(s) {
	case Amharic:
		return Amharic
	case English:
		return English
	default:
		return Default
	}
}
