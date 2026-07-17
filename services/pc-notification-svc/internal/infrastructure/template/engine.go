package template

import (
	"bytes"
	gotpl "text/template"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Render(templateString string, data map[string]interface{}) (string, error) {
	t, err := gotpl.New("msg").Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
