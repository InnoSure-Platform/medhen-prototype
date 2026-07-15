package svcboot

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/InnoSure-Platform/pc-platform/internal/auditrelay"
	"github.com/InnoSure-Platform/pc-platform/internal/gatewayproxy"
	"github.com/InnoSure-Platform/pc-platform/internal/httphandlers"
	"github.com/InnoSure-Platform/pc-platform/internal/integration"
	"github.com/InnoSure-Platform/pc-platform/internal/runtime"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
)

func Start(name, addr, mode string) {
	ctx := context.Background()
	switch mode {
	case "integration":
		startIntegration(ctx, addr, name)
		return
	case "gateway":
		if gatewayproxy.Enabled() {
			startGatewayProxy(addr, name)
			return
		}
	case "audit":
		startAudit(ctx, addr, name)
		return
	}
	startStandard(ctx, addr, name, mode)
}

func startStandard(ctx context.Context, addr, name, mode string) {
	m := runtime.BuildMotor(ctx)
	api := &httphandlers.API{M: m}
	r := baseRouter(name)
	mount := func(router chi.Router) {
		switch mode {
		case "party":
			api.MountParty(router)
		case "policy":
			api.MountPolicy(router)
		case "billing":
			api.MountBilling(router)
		case "claims":
			api.MountClaims(router)
		case "gateway":
			api.MountPublic(router)
		default:
			api.MountAll(router)
		}
	}
	r.Route("/api/v1", mount)
	r.Route("/internal/v1", mount)
	runtime.Listen(name, addr, r)
}

func startAudit(ctx context.Context, addr, name string) {
	repo := runtime.OpenStore(ctx)
	_ = runtime.OpenKafka(ctx, repo)
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		var b []string
		for _, p := range strings.Split(brokers, ",") {
			if t := strings.TrimSpace(p); t != "" {
				b = append(b, t)
			}
		}
		go auditrelay.ConsumeDomainEvents(ctx, b, repo)
	}
	m := runtime.MotorFromRepo(repo)
	api := &httphandlers.API{M: m}
	r := baseRouter(name)
	r.Route("/api/v1", api.MountAudit)
	r.Route("/internal/v1", api.MountAudit)
	runtime.Listen(name, addr, r)
}

func startGatewayProxy(addr, name string) {
	r := runtime.BaseRouter()
	r.Get("/health", runtime.Health(name))
	docs := runtime.FileServer("")
	r.Get("/files/*", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(docs, filepath.Base(chi.URLParam(req, "*"))))
	})
	gatewayproxy.Mount(r)
	runtime.Listen(name, addr, r)
}

func startIntegration(ctx context.Context, addr, name string) {
	pay := integration.NewTelebirrFromEnv()
	sms := &integration.MockSMS{}
	r := runtime.BaseRouter()
	r.Get("/health", runtime.Health(name))
	r.Route("/internal/v1", func(r chi.Router) {
		r.Post("/telebirr/charge", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				Phone       string `json:"phone"`
				AmountMinor int64  `json:"amountMinor"`
				Reference   string `json:"reference"`
			}
			if err := httpx.DecodeJSON(req, &body); err != nil {
				httpx.WriteError(w, err)
				return
			}
		id, err := pay.Charge(body.Phone, body.AmountMinor, body.Reference)
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		mode := "mock"
		if os.Getenv("TELEBIRR_APP_ID") != "" {
			mode = "sandbox"
		}
		httpx.WriteJSON(w, 200, map[string]string{"receiptId": id, "mode": mode})
		})
		r.Post("/sms/send", func(w http.ResponseWriter, req *http.Request) {
			var body struct{ To, Body string }
			_ = httpx.DecodeJSON(req, &body)
			_ = sms.Send(body.To, body.Body)
			httpx.WriteJSON(w, 200, map[string]string{"status": "sent"})
		})
	})
	runtime.Listen(name, addr, r)
}

func baseRouter(name string) chi.Router {
	r := runtime.BaseRouter()
	r.Get("/health", runtime.Health(name))
	docs := runtime.FileServer("")
	r.Get("/files/*", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(docs, filepath.Base(chi.URLParam(req, "*"))))
	})
	return r
}

func Main(name, defaultAddr, mode string) {
	addr := runtime.Env("MEDHEN_ADDR", defaultAddr)
	Start(name, addr, mode)
}
