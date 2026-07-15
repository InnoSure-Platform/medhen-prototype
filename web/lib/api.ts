const API_BASE = process.env.NEXT_PUBLIC_MEDHEN_API ?? "http://localhost:8080";
const API = `${API_BASE}/api/v1`;

let tokenGetter: () => string | null = () => null;

export function setTokenGetter(fn: () => string | null) {
  tokenGetter = fn;
}

function headers(locale: string, idem?: string): HeadersInit {
  const h: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Tenant-ID": "eic",
    "Accept-Language": locale,
  };
  const tok = tokenGetter();
  if (tok) {
    h.Authorization = `Bearer ${tok}`;
  } else {
    h["X-User-ID"] = "demo-agent";
  }
  if (idem) h["Idempotency-Key"] = idem;
  return h;
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, init);
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || res.statusText);
  }
  return res.json() as Promise<T>;
}

export type Party = { id: string; fullName: string; fullNameAm?: string; phoneE164: string };
export type PremiumLine = { code: string; label: string; labelAm: string; amountMinor: number };
export type Quote = {
  id: string;
  status: string;
  lines: PremiumLine[];
  totalMinor: number;
  currency: string;
  uwDecision: string;
  risk: { plateNumber: string; make: string; model: string; year: number; coverType: string; sumInsuredMinor: number };
};
export type BindResponse = {
  policy: { id: string; status: string; totalMinor: number; policyNumber?: string };
  invoice: { id: string; amountMinor: number; status: string };
};
export type PaymentResult = {
  receiptId: string;
  policy: { id: string; policyNumber: string; status: string; totalMinor: number };
  documents: { id: string; type: string; locale: string; url: string; objectKey?: string }[];
};
export type Claim = {
  id: string;
  claimNumber: string;
  status: string;
  track: string;
  settlementMinor?: number;
};
export type KPI = {
  policiesInForce: number;
  gwpMinor: number;
  claimsSettled: number;
  productCode: string;
};
export type Audit = {
  id: string;
  entityType: string;
  entityId: string;
  action: string;
  actor: string;
  detail?: string;
  at: string;
};

export const api = {
  registerParty: (locale: string, body: object) =>
    req<Party>("/parties", { method: "POST", headers: headers(locale, crypto.randomUUID()), body: JSON.stringify(body) }),
  createQuote: (locale: string, body: object) =>
    req<Quote>("/quotes", { method: "POST", headers: headers(locale, crypto.randomUUID()), body: JSON.stringify(body) }),
  bindQuote: (locale: string, quoteId: string) =>
    req<BindResponse>(`/quotes/${quoteId}/bind`, { method: "POST", headers: headers(locale, crypto.randomUUID()), body: "{}" }),
  pay: (locale: string, invoiceId: string, phone: string) =>
    req<PaymentResult>(`/billing/invoices/${invoiceId}/pay`, {
      method: "POST",
      headers: headers(locale, crypto.randomUUID()),
      body: JSON.stringify({ channel: "telebirr", phone }),
    }),
  submitFNOL: (locale: string, body: object) =>
    req<Claim>("/claims", { method: "POST", headers: headers(locale, crypto.randomUUID()), body: JSON.stringify(body) }),
  settle: (locale: string, claimId: string) =>
    req<Claim>(`/claims/${claimId}/settle`, { method: "POST", headers: headers(locale, crypto.randomUUID()), body: "{}" }),
  kpis: (locale: string) => req<KPI>("/demo/kpis", { headers: headers(locale) }),
  audit: (locale: string) => req<Audit[]>("/audit?limit=30", { headers: headers(locale) }),
  riskSchema: (locale: string) => req<Record<string, unknown>>("/products/MOTOR-PRIVATE-COMP/risk-schema", { headers: headers(locale) }),
  fileUrl: (path: string) => `${API_BASE}${path}`,
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

export function docLabel(type: string, locale: string): string {
  const en: Record<string, string> = {
    schedule: "Policy Schedule",
    coi: "Certificate of Insurance",
    sticker: "Windshield QR Sticker",
  };
  const am: Record<string, string> = {
    schedule: "የፖሊሲ መርሃ ግብር",
    coi: "የኢንሹራንስ የምስክር ወረቀት",
    sticker: "የመኪና መስታወት QR ስቲከር",
  };
  return locale === "am" ? am[type] ?? type : en[type] ?? type;
}
