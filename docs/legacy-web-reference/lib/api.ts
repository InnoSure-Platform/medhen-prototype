// API client for the Medhen modular monolith.
//
// All calls go same-origin to the Next.js proxy under /api/medhen/* (see
// app/api/medhen/[...path]/route.ts), which attaches the access token and the
// tenant server-side (C5). The browser therefore holds no token and sends no
// X-Tenant-ID. Paths below mirror the monolith's real module routes.
const PROXY_BASE = "/api/medhen";

function headers(locale: string, idem?: string): HeadersInit {
  const h: Record<string, string> = {
    "Content-Type": "application/json",
    "Accept-Language": locale,
  };
  if (idem) h["Idempotency-Key"] = idem;
  return h;
}

// ApiError carries only the HTTP status and a user-safe message. Raw backend
// error bodies are never surfaced to users (M8).
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function safeMessage(status: number): string {
  switch (true) {
    case status === 400:
      return "The request was invalid. Please review the details and try again.";
    case status === 401:
      return "Your session has expired. Please sign in again.";
    case status === 403:
      return "You do not have permission to perform this action.";
    case status === 404:
      return "The requested record could not be found.";
    case status === 409:
      return "This action conflicts with the current state (it may already be processed or require manual review).";
    case status === 422:
      return "The request could not be processed. Please check the values and try again.";
    case status >= 500:
      return "The service is temporarily unavailable. Please try again shortly.";
    default:
      return "Something went wrong. Please try again.";
  }
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${PROXY_BASE}${path}`, init);
  if (!res.ok) {
    // Drain the body so the connection is freed, but do not expose it.
    await res.text().catch(() => "");
    throw new ApiError(res.status, safeMessage(res.status));
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

const idem = () => crypto.randomUUID();

// ---- Response shapes (match the monolith's domain JSON) --------------------
// Money fields are JSON numbers in major units (Birr), e.g. 2160.00.
// Domain structs are serialised with Go field names (PascalCase keys).

export type RegisterResponse = { id: string };

export type Quote = {
  ID: string;
  PartyID: string;
  ProductCode: string;
  Coverages: string[];
  RiskDimensions: Record<string, string>;
  NetPremium: number;
  TotalTaxes: number;
  GrossPremium: number;
  CalculationID: string;
  Status: "QUOTED" | "BOUND" | "EXPIRED";
  Version: number;
  CreatedAt: string;
  UpdatedAt: string;
};

export type Policy = {
  ID: string;
  PolicyNumber: string;
  QuoteID: string;
  PartyID: string;
  ProductCode: string;
  GrossPremium: number;
  Status: "ISSUED" | "CANCELLED";
  EffectiveFrom: string;
  EffectiveTo: string;
  IssuedAt: string;
  Version: number;
};

export type Invoice = {
  ID: string;
  PolicyID: string;
  PartyID: string;
  AmountDue: number;
  AmountPaid: number;
  Status: string;
  CreatedAt: string;
  UpdatedAt: string;
  Version: number;
};

export type Claim = {
  ID: string;
  PolicyID: string;
  PartyID: string;
  Status: "FILED" | "SETTLED" | "REJECTED";
  Description: string;
  Latitude: number;
  Longitude: number;
  Reserve: number;
  SettledAmount: number;
  CreatedAt: string;
  UpdatedAt: string;
  Version: number;
};

// KPI view is serialised with snake_case JSON tags and minor-unit money.
export type KPI = {
  tenant_id: string;
  premium_written_minor: number;
  claims_paid_minor: number;
  policy_count: number;
  claim_count: number;
  loss_ratio: number;
  assumed_expense_ratio: number;
  combined_ratio: number;
};

// Audit records are serialised with snake_case JSON tags.
export type Audit = {
  id: string;
  topic: string;
  tenant_id: string;
  payload: unknown;
  recorded_at: string;
};

export type Product = {
  Code: string;
  LOB: string;
  Name: string;
  NameAmharic: string;
  Status: string;
  RateVersion: string;
  Coverages: { Code: string; Name: string; NameAmharic: string; BaseRate: number }[];
  Factors: unknown[];
};

// ---- Request payloads ------------------------------------------------------

export type RegisterPartyInput = {
  full_name: string;
  full_name_amharic: string;
  phone_e164: string;
  national_id?: string;
  address: {
    region: string;
    zone: string;
    woreda: string;
    kebele: string;
    city: string;
    line1: string;
  };
};

export type CreateQuoteInput = {
  party_id: string;
  product_code: string;
  coverages: string[];
  risk_dimensions: Record<string, string>;
};

export type FNOLInput = {
  policy_id: string;
  description: string;
  latitude: number;
  longitude: number;
  reserve_minor: number;
};

// ---- Operations ------------------------------------------------------------

export const api = {
  registerParty: (locale: string, body: RegisterPartyInput) =>
    req<RegisterResponse>("/party/parties", { method: "POST", headers: headers(locale, idem()), body: JSON.stringify(body) }),
  getParty: (locale: string, id: string) =>
    req<Record<string, unknown>>(`/party/parties/${id}`, { headers: headers(locale) }),

  createQuote: (locale: string, body: CreateQuoteInput) =>
    req<Quote>("/policy/quotes", { method: "POST", headers: headers(locale, idem()), body: JSON.stringify(body) }),
  getQuote: (locale: string, id: string) =>
    req<Quote>(`/policy/quotes/${id}`, { headers: headers(locale) }),
  bindQuote: (locale: string, quoteId: string) =>
    req<Policy>(`/policy/quotes/${quoteId}/bind`, { method: "POST", headers: headers(locale, idem()), body: "{}" }),
  getPolicy: (locale: string, policyId: string) =>
    req<Policy>(`/policy/policies/${policyId}`, { headers: headers(locale) }),

  getInvoice: (locale: string, invoiceId: string) =>
    req<Invoice>(`/billing/invoices/${invoiceId}`, { headers: headers(locale) }),

  submitFNOL: (locale: string, body: FNOLInput) =>
    req<Claim>("/claims/claims", { method: "POST", headers: headers(locale, idem()), body: JSON.stringify(body) }),
  getClaim: (locale: string, id: string) =>
    req<Claim>(`/claims/claims/${id}`, { headers: headers(locale) }),
  settleClaim: (locale: string, claimId: string, amountMinor: number) =>
    req<Claim>(`/claims/claims/${claimId}/settle`, { method: "POST", headers: headers(locale, idem()), body: JSON.stringify({ amount_minor: amountMinor }) }),

  kpis: (locale: string) => req<KPI>("/reporting/kpis", { headers: headers(locale) }),
  audit: (locale: string) => req<Audit[]>("/audit/logs?limit=30", { headers: headers(locale) }),
  listProducts: (locale: string) => req<Product[]>("/product/products", { headers: headers(locale) }),
  getProduct: (locale: string, code: string) => req<Product>(`/product/products/${code}`, { headers: headers(locale) }),

  saveLastPolicy: (policyId: string, policyNumber: string) => {
    if (typeof window !== "undefined") {
      sessionStorage.setItem("medhen_last_policy", JSON.stringify({ policyId, policyNumber }));
    }
  },
  loadLastPolicy: (): { policyId: string; policyNumber: string } | null => {
    if (typeof window === "undefined") return null;
    try {
      const raw = sessionStorage.getItem("medhen_last_policy");
      return raw ? JSON.parse(raw) : null;
    } catch {
      return null;
    }
  },
};

// errorMessage extracts a user-safe string from a thrown error, never a raw
// backend body (M8).
export function errorMessage(e: unknown): string {
  if (e instanceof ApiError) return e.message;
  return "Something went wrong. Please try again.";
}

