// Package httphandlers exposes REST handlers shared by gateway and microservices.
package httphandlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/InnoSure-Platform/pc-platform/internal/product"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-platform/internal/usecase"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
)

type API struct {
	M *usecase.Motor
}

func (a *API) MountPublic(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]string{"status": "ok"})
	})
	a.MountParty(r)
	a.MountPolicy(r)
	a.MountBilling(r)
	a.MountClaims(r)
	a.MountAudit(r)
	r.Get("/demo/kpis", a.kpis)
}

func (a *API) MountParty(r chi.Router) {
	r.Post("/parties", a.registerParty)
	r.Get("/parties/{partyId}", a.getParty)
}

func (a *API) MountPolicy(r chi.Router) {
	r.Get("/products/{productCode}", a.getProduct)
	r.Get("/products/{productCode}/risk-schema", a.riskSchema)
	r.Post("/quotes", a.createQuote)
	r.Get("/quotes/{quoteId}", a.getQuote)
	r.Post("/quotes/{quoteId}/bind", a.bindQuote)
	r.Get("/policies/{policyId}", a.getPolicy)
	r.Get("/policies/{policyId}/documents", a.listDocs)
	r.Get("/demo/kpis", a.kpis)
}

func (a *API) MountBilling(r chi.Router) {
	r.Post("/billing/invoices/{invoiceId}/pay", a.payInvoice)
}

func (a *API) MountClaims(r chi.Router) {
	r.Post("/claims", a.submitFNOL)
	r.Post("/claims/{claimId}/settle", a.settleClaim)
}

func (a *API) MountAudit(r chi.Router) {
	r.Get("/audit", a.queryAudit)
}

func (a *API) MountAll(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, 200, map[string]string{"status": "ok"})
	})
	a.MountParty(r)
	a.MountPolicy(r)
	a.MountBilling(r)
	a.MountClaims(r)
	a.MountAudit(r)
}

func idem(r *http.Request) string {
	if k := httpx.IdempotencyKey(r); k != "" {
		return k
	}
	return uuid.NewString()
}

func (a *API) registerParty(w http.ResponseWriter, r *http.Request) {
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
	p, err := a.M.RegisterParty(r.Context(), usecase.RegisterPartyCmd{
		FullName: body.FullName, FullNameAm: body.FullNameAm, PhoneE164: body.PhoneE164,
		Email: body.Email, Address: body.Address, Actor: tc.UserID, IdemKey: idem(r),
	})
	writeResult(w, err, 201, p)
}

func (a *API) getParty(w http.ResponseWriter, r *http.Request) {
	p, err := a.M.GetParty(r.Context(), chi.URLParam(r, "partyId"))
	writeResult(w, err, 200, p)
}

func (a *API) getProduct(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r, "productCode") != a.M.Product.Code {
		httpx.WriteJSON(w, 404, map[string]string{"message": "not found"})
		return
	}
	httpx.WriteJSON(w, 200, a.M.Product)
}

func (a *API) riskSchema(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, 200, product.RiskSchema())
}

func (a *API) createQuote(w http.ResponseWriter, r *http.Request) {
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
	q, err := a.M.CreateQuote(r.Context(), usecase.CreateQuoteCmd{
		PartyID: body.PartyID, ProductCode: body.ProductCode, Risk: body.Risk,
		Locale: i18n.ParseLocale(tc.Locale), Actor: tc.UserID, IdemKey: idem(r),
	})
	writeResult(w, err, 201, q)
}

func (a *API) getQuote(w http.ResponseWriter, r *http.Request) {
	q, err := a.M.GetQuote(r.Context(), chi.URLParam(r, "quoteId"))
	writeResult(w, err, 200, q)
}

func (a *API) bindQuote(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	br, err := a.M.BindQuote(r.Context(), chi.URLParam(r, "quoteId"), tc.UserID, idem(r))
	writeResult(w, err, 200, br)
}

func (a *API) payInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Channel string `json:"channel"`
		Phone   string `json:"phone"`
	}
	_ = httpx.DecodeJSON(r, &body)
	tc := httpx.TenantFromRequest(r)
	pr, err := a.M.PayInvoice(r.Context(), chi.URLParam(r, "invoiceId"), body.Channel, body.Phone, tc.UserID, idem(r))
	writeResult(w, err, 200, pr)
}

func (a *API) getPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := a.M.GetPolicy(r.Context(), chi.URLParam(r, "policyId"))
	writeResult(w, err, 200, p)
}

func (a *API) listDocs(w http.ResponseWriter, r *http.Request) {
	d, err := a.M.ListPolicyDocuments(r.Context(), chi.URLParam(r, "policyId"))
	writeResult(w, err, 200, d)
}

func (a *API) submitFNOL(w http.ResponseWriter, r *http.Request) {
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
	cl, err := a.M.SubmitFNOL(r.Context(), usecase.FNOLCmd{
		PolicyID: body.PolicyID, LossDate: ld, Description: body.Description,
		Latitude: body.Latitude, Longitude: body.Longitude, EstimatedAmountMinor: body.EstimatedAmountMinor,
		PhotoObjectKeys: body.PhotoObjectKeys, Actor: tc.UserID, IdemKey: idem(r),
	})
	writeResult(w, err, 201, cl)
}

func (a *API) settleClaim(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	cl, err := a.M.SettleFastTrack(r.Context(), chi.URLParam(r, "claimId"), tc.UserID, idem(r))
	writeResult(w, err, 200, cl)
}

func (a *API) queryAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := a.M.QueryAudit(r.Context(), r.URL.Query().Get("entityType"), r.URL.Query().Get("entityId"), limit)
	writeResult(w, err, 200, rows)
}

func (a *API) kpis(w http.ResponseWriter, r *http.Request) {
	k, err := a.M.KPIs(r.Context())
	writeResult(w, err, 200, k)
}

func writeResult(w http.ResponseWriter, err error, ok int, body any) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, ok, body)
}
