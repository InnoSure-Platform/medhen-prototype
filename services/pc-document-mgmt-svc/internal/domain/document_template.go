package domain

import (
	"time"
)

type DocumentTemplate struct {
	ID          string
	Code        string
	Version     int
	Locale      string
	HtmlContent string
	CssContent  string
	MergeSchema map[string]interface{}
	CreatedAt   time.Time
}

func NewDocumentTemplate(id, code string, version int, locale, html, css string, schema map[string]interface{}) *DocumentTemplate {
	return &DocumentTemplate{
		ID:          id,
		Code:        code,
		Version:     version,
		Locale:      locale,
		HtmlContent: html,
		CssContent:  css,
		MergeSchema: schema,
		CreatedAt:   time.Now().UTC(),
	}
}
