// Package httphandlers exposes REST handlers shared by gateway and microservices.
package httphandlers

import (
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"
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
	r.Post("/parties/{partyId}/kyc-verify", a.verifyKYC)
	r.Get("/parties/{partyId}", a.getParty)
	MountMetrics(r)
}

func (a *API) MountPolicy(r chi.Router) {
	r.Get("/products/{productCode}", a.getProduct)
	r.Get("/products/{productCode}/risk-schema", a.riskSchema)
	r.Post("/quotes", a.createQuote)
	r.Get("/quotes", a.listQuotes)
	r.Get("/quotes/{quoteId}", a.getQuote)
	r.Post("/quotes/{quoteId}/bind", a.bindQuote)
	r.Post("/quotes/{quoteId}/approve", a.approveQuote)
	r.Post("/quotes/{quoteId}/decline", a.declineQuote)
	r.Post("/billing/eod-reconciliation", a.eodReconciliation)
	r.Get("/policies/{policyId}", a.getPolicy)
	r.Post("/policies/{policyId}/endorse", a.endorsePolicy)
	r.Post("/policies/{policyId}/renew", a.renewPolicy)
	r.Post("/policies/{policyId}/cancel", a.cancelPolicy)
	r.Get("/policies/{policyId}/documents", a.listDocs)
	r.Get("/demo/kpis", a.kpis)
}

func (a *API) MountBilling(r chi.Router) {
	r.Post("/billing/invoices/{invoiceId}/pay", a.payInvoice)
}

func (a *API) MountClaims(r chi.Router) {
	r.Post("/claims", a.submitFNOL)
	r.Get("/claims", a.listClaims)
	r.Post("/claims/{claimId}/settle", a.settleClaim)
	r.Post("/claims/{claimId}/reserve", a.adjustReserve)
	r.Post("/claims/{claimId}/recovery", a.recordRecovery)
}

func (a *API) MountFincrime(r chi.Router) {
	r.Post("/fincrime/screen", a.screenParty)
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

func (a *API) verifyKYC(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	var body struct {
		FaydaID string `json:"faydaId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, pcerr.E(pcerr.CodeValidation, "invalid json"))
		return
	}
	err := a.M.VerifyKYC(r.Context(), chi.URLParam(r, "partyId"), body.FaydaID, tc.UserID)
	writeResult(w, err, 200, map[string]string{"status": "ok"})
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

func (a *API) listQuotes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("status") == "REFERRED" {
		qs, err := a.M.ListReferredQuotes(r.Context())
		writeResult(w, err, 200, qs)
		return
	}
	writeResult(w, nil, 200, []any{}) // Other statuses not implemented for listing
}

func (a *API) approveQuote(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	err := a.M.ApproveQuote(r.Context(), chi.URLParam(r, "quoteId"), tc.UserID)
	writeResult(w, err, 200, map[string]string{"status": "ok"})
}

func (a *API) declineQuote(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	err := a.M.DeclineQuote(r.Context(), chi.URLParam(r, "quoteId"), tc.UserID)
	writeResult(w, err, 200, map[string]string{"status": "ok"})
}

func (a *API) bindQuote(w http.ResponseWriter, r *http.Request) {
	var body struct {
		InstallmentPlan string `json:"installmentPlan"`
	}
	_ = httpx.DecodeJSON(r, &body)
	if body.InstallmentPlan == "" {
		body.InstallmentPlan = "100_UPFRONT"
	}

	tc := httpx.TenantFromRequest(r)
	br, err := a.M.BindQuote(r.Context(), chi.URLParam(r, "quoteId"), body.InstallmentPlan, tc.UserID, idem(r))
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

func (a *API) eodReconciliation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Date string `json:"date"`
	}
	_ = httpx.DecodeJSON(r, &body)
	res, err := a.M.RunEndOfDayReconciliation(r.Context(), body.Date)
	writeResult(w, err, 200, res)
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
	var body struct {
		SettlementMinor int64 `json:"settlementMinor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, pcerr.E(pcerr.CodeValidation, "invalid json"))
		return
	}
	cl, err := a.M.SettleClaim(r.Context(), chi.URLParam(r, "claimId"), tc.UserID, idem(r), body.SettlementMinor)
	writeResult(w, err, 200, cl)
}

func (a *API) adjustReserve(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	var body struct {
		AmountMinor int64 `json:"amountMinor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, pcerr.E(pcerr.CodeValidation, "invalid json"))
		return
	}
	err := a.M.AdjustReserve(r.Context(), chi.URLParam(r, "claimId"), tc.UserID, body.AmountMinor)
	writeResult(w, err, 200, map[string]string{"status": "ok"})
}

func (a *API) recordRecovery(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	var body struct {
		AmountMinor int64 `json:"amountMinor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, pcerr.E(pcerr.CodeValidation, "invalid json"))
		return
	}
	err := a.M.RecordRecovery(r.Context(), chi.URLParam(r, "claimId"), tc.UserID, body.AmountMinor)
	writeResult(w, err, 200, map[string]string{"status": "ok"})
}

func (a *API) listClaims(w http.ResponseWriter, r *http.Request) {
	cls, err := a.M.Repo.ListClaims(r.Context())
	writeResult(w, err, 200, cls)
}

func (a *API) screenParty(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PartyID string `json:"partyId"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	// Stub for InnoGuard integration
	httpx.WriteJSON(w, 200, map[string]string{"status": "CLEARED"})
}

func (a *API) endorsePolicy(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	var body struct {
		Risk store.MotorRisk `json:"risk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, pcerr.E(pcerr.CodeValidation, "invalid json"))
		return
	}
	pol, err := a.M.EndorsePolicy(r.Context(), usecase.EndorsePolicyCmd{
		PolicyID: chi.URLParam(r, "policyId"),
		Risk:     body.Risk,
		Actor:    tc.UserID,
		IdemKey:  idem(r),
	})
	writeResult(w, err, 200, pol)
}

func (a *API) renewPolicy(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	pol, err := a.M.RenewPolicy(r.Context(), usecase.RenewPolicyCmd{
		PolicyID: chi.URLParam(r, "policyId"),
		Actor:    tc.UserID,
		IdemKey:  idem(r),
	})
	writeResult(w, err, 200, pol)
}

func (a *API) cancelPolicy(w http.ResponseWriter, r *http.Request) {
	tc := httpx.TenantFromRequest(r)
	err := a.M.CancelPolicy(r.Context(), usecase.CancelPolicyCmd{
		PolicyID: chi.URLParam(r, "policyId"),
		Actor:    tc.UserID,
		IdemKey:  idem(r),
	})
	writeResult(w, err, 204, nil)
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
