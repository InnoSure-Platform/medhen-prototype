package i18n

import (
	"errors"
)

// Translator defines the interface for localizing messages.
type Translator interface {
	Translate(key, lang string) string
}

// InMemoryTranslator implements a basic thread-safe dictionary translator.
type InMemoryTranslator struct {
	dict map[string]map[string]string // lang -> key -> translation
}

// NewInMemoryTranslator initializes an empty translator.
func NewInMemoryTranslator() *InMemoryTranslator {
	return &InMemoryTranslator{
		dict: make(map[string]map[string]string),
	}
}

// AddTranslation adds a key-value translation for a specific language.
func (t *InMemoryTranslator) AddTranslation(lang, key, translation string) error {
	if lang == "" || key == "" {
		return errors.New("lang and key must not be empty")
	}

	if t.dict[lang] == nil {
		t.dict[lang] = make(map[string]string)
	}
	t.dict[lang][key] = translation
	return nil
}

// Translate returns the localized string or the raw key if not found.
// Falls back to "en" if the language is unsupported.
func (t *InMemoryTranslator) Translate(key, lang string) string {
	if langDict, ok := t.dict[lang]; ok {
		if val, found := langDict[key]; found {
			return val
		}
	}
	// Fallback to en
	if enDict, ok := t.dict["en"]; ok {
		if val, found := enDict[key]; found {
			return val
		}
	}
	// Return the key itself as the final fallback
	return key
}
