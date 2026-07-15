// Package httpx provides shared HTTP helpers for edge and service adapters.
package httpx

import (
	"encoding/json"
	"net/http"

	pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
	"github.com/InnoSure-Platform/pc-shared-go/tenant"
)

type ctxKey int

const (
	KeyTenant ctxKey = iota + 1
	KeyIdempotency
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err error) {
	if pe, ok := err.(*pcerr.Error); ok {
		WriteJSON(w, pcerr.HTTPStatus(pe.Code), map[string]any{
			"code":    pe.Code,
			"message": pe.Message,
		})
		return
	}
	WriteJSON(w, 500, map[string]any{"code": pcerr.CodeInternal, "message": err.Error()})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func TenantFromRequest(r *http.Request) tenant.Context {
	tid := r.Header.Get("X-Tenant-ID")
	if tid == "" {
		tid = tenant.EIC
	}
	return tenant.Context{
		TenantID: tid,
		UserID:   r.Header.Get("X-User-ID"),
		Locale:   string(i18n.ParseLocale(r.Header.Get("Accept-Language"))),
	}
}

func IdempotencyKey(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, Accept-Language, X-Tenant-ID, X-User-ID, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = r.Header.Get("Traceparent")
		}
		if id != "" {
			w.Header().Set("X-Request-ID", id)
		}
		next.ServeHTTP(w, r)
	})
}
