// Domain response shapes matching the monolith's JSON. The OpenAPI spec types
// some responses loosely; these are the app's source of truth. Money on reads is
// a MAJOR-unit number (e.g. 2160.00); writes use *_minor integers. Policy/quote/
// claim structs serialise with Go field names (PascalCase).

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
  Status: "FILED" | "SETTLED" | "REJECTED" | "REFERRED";
  Description: string;
  Latitude: number;
  Longitude: number;
  Reserve: number;
  SettledAmount: number;
  CreatedAt: string;
  UpdatedAt: string;
  Version: number;
};

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

export type Party = {
  ID: string;
  FullName: string;
  FullNameAmharic: string;
  PhoneE164: string;
  NationalID: string;
  Status: string;
  Type: string;
};

export type PaymentIntent = {
  invoice_id: string;
  reference: string;
  amount_minor: number;
  checkout_url: string;
  status: string;
};

export type IamUser = {
  id: string;
  subject: string;
  email?: string;
  full_name?: string;
  roles: string[];
};

export type FNOLInput = {
  policy_id: string;
  description: string;
  latitude: number;
  longitude: number;
  reserve_minor: number;
};

export type RegisterUserInput = {
  subject: string;
  email?: string;
  full_name?: string;
  roles: string[];
};
