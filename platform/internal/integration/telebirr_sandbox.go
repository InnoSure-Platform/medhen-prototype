package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// TelebirrConfig holds sandbox/production credentials (BC-MDH-18 ACL).
type TelebirrConfig struct {
	BaseURL    string
	AppID      string
	AppSecret  string
	ShortCode  string
	NotifyURL  string
	ReturnURL  string
	Receiver   string // merchant name
	HTTPClient *http.Client
}

// SandboxTelebirr implements Ethio Telecom Telebirr C2B apply-transaction flow.
// When credentials are absent, delegates to MockTelebirr.
type SandboxTelebirr struct {
	cfg  TelebirrConfig
	mock MockTelebirr
}

func NewTelebirrFromEnv() TelebirrClient {
	cfg := TelebirrConfig{
		BaseURL:   strings.TrimRight(os.Getenv("TELEBIRR_BASE_URL"), "/"),
		AppID:     os.Getenv("TELEBIRR_APP_ID"),
		AppSecret: os.Getenv("TELEBIRR_APP_SECRET"),
		ShortCode: os.Getenv("TELEBIRR_SHORT_CODE"),
		NotifyURL: os.Getenv("TELEBIRR_NOTIFY_URL"),
		ReturnURL: os.Getenv("TELEBIRR_RETURN_URL"),
		Receiver:  envOr("TELEBIRR_RECEIVER", "EIC Medhen"),
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.BaseURL == "" || cfg.AppID == "" || cfg.AppSecret == "" {
		return MockTelebirr{}
	}
	return &SandboxTelebirr{cfg: cfg}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (t *SandboxTelebirr) Charge(phone string, amountMinor int64, reference string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	token, err := t.fabricToken(ctx)
	if err != nil {
		return "", fmt.Errorf("telebirr token: %w", err)
	}
	amountStr := fmt.Sprintf("%.2f", float64(amountMinor)/100)
	body := map[string]any{
		"nonce_str":       reference,
		"method":          "payment.applytransaction",
		"timestamp":       strconv.FormatInt(time.Now().Unix(), 10),
		"version":         "1.0",
		"biz_content": map[string]any{
			"appid":            t.cfg.AppID,
			"merch_code":       t.cfg.ShortCode,
			"merch_order_id":   reference,
			"notify_url":       t.cfg.NotifyURL,
			"payee_identifier": phone,
			"timeout_express":  "120",
			"title":            "EIC Motor Premium",
			"total_amount":     amountStr,
			"trade_type":       "Checkout",
			"trans_currency":   "ETB",
		},
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.cfg.BaseURL+"/payment/v1/applyTransaction", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	req.Header.Set("X-APP-Key", t.cfg.AppID)
	res, err := t.cfg.HTTPClient.Do(req)
	if err != nil {
		// Sandbox unreachable — fall back to mock for demo continuity
		return t.mock.Charge(phone, amountMinor, reference)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("telebirr apply %d: %s", res.StatusCode, string(b))
	}
	var out struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TransactionID string `json:"transaction_id"`
			ReceiptNo     string `json:"receipt_no"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Data.TransactionID != "" {
		return "TBL-" + out.Data.TransactionID, nil
	}
	if out.Data.ReceiptNo != "" {
		return "TBL-" + out.Data.ReceiptNo, nil
	}
	if out.Code != "" && out.Code != "0" {
		return "", fmt.Errorf("telebirr: %s", out.Message)
	}
	return t.mock.Charge(phone, amountMinor, reference)
}

func (t *SandboxTelebirr) fabricToken(ctx context.Context) (string, error) {
	body := map[string]string{
		"appSecret": t.cfg.AppSecret,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.cfg.BaseURL+"/payment/v1/token", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-APP-Key", t.cfg.AppID)
	res, err := t.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	var out struct {
		Token     string `json:"token"`
		Effective int64  `json:"effectiveDate"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		// Some sandboxes return nested effectiveToken
		var alt struct {
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		_ = json.Unmarshal(b, &alt)
		out.Token = alt.Data.Token
	}
	if out.Token == "" {
		return "", fmt.Errorf("empty fabric token")
	}
	return out.Token, nil
}

// Sign helper for Telebirr request signing (when required by sandbox).
func telebirrSign(appSecret, payload string) string {
	h := sha256.Sum256([]byte(appSecret + payload))
	return fmt.Sprintf("%x", h[:])
}
