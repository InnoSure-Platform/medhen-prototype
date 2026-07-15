package app

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/InnoSure-Platform/pc-platform/internal/product"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
	mw "github.com/InnoSure-Platform/pc-shared-go/middleware"
	"github.com/InnoSure-Platform/pc-shared-go/tenant"
)

func NewRouter(p *Platform) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(httpx.CORS)
	r.Use(httpx.RequestID)
	r.Use(mw.Recover)
	r.Use(mw.Logging)
	r.Use(mw.DemoAuth)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]any{
			"status": "ok", "service": "medhen-api", "tenant": tenant.EIC, "product": "Medhen Platform Phase 0",
		})
	})

	r.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(p.DocsDir, filepath.Base(chi.URLParam(r, "*"))))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			httpx.WriteJSON(w, 200, map[string]string{"status": "ok"})
		})

		r.Post("/parties", p.handleRegisterParty)
		r.Get("/parties/{partyId}", p.handleGetParty)
		r.Get("/products/{productCode}", p.handleGetProduct)
		r.Get("/products/{productCode}/risk-schema", p.handleRiskSchema)
		r.Post("/quotes", p.handleCreateQuote)
		r.Get("/quotes/{quoteId}", p.handleGetQuote)
		r.Post("/quotes/{quoteId}/bind", p.handleBind)
		r.Post("/billing/invoices/{invoiceId}/pay", p.handlePay)
		r.Get("/policies/{policyId}", p.handleGetPolicy)
		r.Get("/policies/{policyId}/documents", p.handleListDocs)
		r.Post("/claims", p.handleFNOL)
		r.Post("/claims/{claimId}/settle", p.handleSettle)
		r.Get("/audit", p.handleAudit)
		r.Get("/demo/kpis", p.handleKPIs)
	})

	_ = os.MkdirAll(p.DocsDir, 0o755)
	return r
}

func idemOrNew(r *http.Request) string {
	if k := httpx.IdempotencyKey(r); k != "" {
		return k
	}
	return uuid.NewString()
}

func (p *Platform) handleRegisterParty(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FullName   string         `json:"fullName"`
		FullNameAm string         `json:"fullNameAm"`
		PhoneE164  string         `json:"phoneE164"`
		Email      string         `json:"email"`
		Address    *store.Address `json:"address"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tc := httpx.TenantFromRequest(r)
	party, err := p.RegisterParty(RegisterPartyCmd{
		FullName: body.FullName, FullNameAm: body.FullNameAm, PhoneE164: body.PhoneE164,
		Email: body.Email, Address: body.Address, Actor: tc.UserID, IdemKey: idemOrNew(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 201, party)
}

func (p *Platform) handleGetParty(w http.ResponseWriter, r *http.Request) {
	party, err := p.GetParty(chi.URLParam(r, "partyId"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, party)
}

func (p *Platform) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r, "productCode") != p.Product.Code {
		httpx.WriteJSON(w, 404, map[string]string{"code": "NOT_FOUND", "message": "product not found"})
		return
	}
	httpx.WriteJSON(w, 200, p.Product)
}

func (p *Platform) handleRiskSchema(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, 200, product.RiskSchema())
}

func (p *Platform) handleCreateQuote(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PartyID     string          `json:"partyId"`
		ProductCode string          `json:"productCode"`
		Risk        store.MotorRisk `json:"risk"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tc := httpx.TenantFromRequest(r)
	q, err := p.CreateQuote(CreateQuoteCmd{
		PartyID: body.PartyID, ProductCode: body.ProductCode, Risk: body.Risk,
		Locale: i18n.ParseLocale(tc.Locale), Actor: tc.UserID, IdemKey: idemOrNew(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 201, q)
}

func (p *Platform) handleGetQuote(w http.ResponseWriter, r *http.Request) {
	q, err := p.GetQuote(chi.URLParam(r, "quoteId"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, q)
}

func (p *Platform) handleBind(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	br, err := p.BindQuote(chi.URLParam(r, "quoteId"), tc.UserID, idemOrNew(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, br)
}

func (p *Platform) handlePay(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Channel string `json:"channel"`
		Phone   string `json:"phone"`
	}
	_ = httpx.DecodeJSON(r, &body)
	tc := httpx.TenantFromRequest(r)
	pr, err := p.PayInvoice(chi.URLParam(r, "invoiceId"), body.Channel, body.Phone, tc.UserID, idemOrNew(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, pr)
}

func (p *Platform) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	pol, err := p.GetPolicy(chi.URLParam(r, "policyId"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, pol)
}

func (p *Platform) handleListDocs(w http.ResponseWriter, r *http.Request) {
	docs, err := p.ListPolicyDocuments(chi.URLParam(r, "policyId"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, docs)
}

func (p *Platform) handleFNOL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PolicyID             string   `json:"policyId"`
		LossDate             string   `json:"lossDate"`
		Description          string   `json:"description"`
		Latitude             float64  `json:"latitude"`
		Longitude            float64  `json:"longitude"`
		EstimatedAmountMinor int64    `json:"estimatedAmountMinor"`
		PhotoObjectKeys      []string `json:"photoObjectKeys"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	ld := time.Now().UTC()
	if body.LossDate != "" {
		if t, err := time.Parse(time.RFC3339, body.LossDate); err == nil {
			ld = t
		}
	}
	tc := httpx.TenantFromRequest(r)
	cl, err := p.SubmitFNOL(FNOLCmd{
		PolicyID: body.PolicyID, LossDate: ld, Description: body.Description,
		Latitude: body.Latitude, Longitude: body.Longitude, EstimatedAmountMinor: body.EstimatedAmountMinor,
		PhotoObjectKeys: body.PhotoObjectKeys, Actor: tc.UserID, IdemKey: idemOrNew(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 201, cl)
}

func (p *Platform) handleSettle(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	cl, err := p.SettleFastTrack(chi.URLParam(r, "claimId"), tc.UserID, idemOrNew(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, 200, cl)
}

func (p *Platform) handleAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	httpx.WriteJSON(w, 200, p.QueryAudit(r.URL.Query().Get("entityType"), r.URL.Query().Get("entityId"), limit))
}

func (p *Platform) handleKPIs(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, 200, p.KPIs())
}
