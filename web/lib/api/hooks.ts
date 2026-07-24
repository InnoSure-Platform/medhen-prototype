"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useLocale } from "next-intl";
import { apiRequest } from "./client";
import type {
  Audit,
  Claim,
  CreateQuoteInput,
  FNOLInput,
  IamUser,
  Invoice,
  KPI,
  Party,
  PaymentIntent,
  Policy,
  Product,
  Quote,
  RegisterPartyInput,
  RegisterResponse,
  RegisterUserInput,
} from "./types";

// Centralised query keys.
export const qk = {
  products: ["products"] as const,
  product: (code: string) => ["product", code] as const,
  quote: (id: string) => ["quote", id] as const,
  policy: (id: string) => ["policy", id] as const,
  invoice: (id: string) => ["invoice", id] as const,
  claim: (id: string) => ["claim", id] as const,
  kpis: ["kpis"] as const,
  audit: (limit: number) => ["audit", limit] as const,
  users: ["users"] as const,
  policies: (limit: number, offset: number) => ["policies", limit, offset] as const,
  claims: (status: string, limit: number, offset: number) => ["claims", status, limit, offset] as const,
  quotes: (limit: number, offset: number) => ["quotes", limit, offset] as const,
  parties: (limit: number, offset: number) => ["parties", limit, offset] as const,
  invoiceByPolicy: (policyId: string) => ["invoice-by-policy", policyId] as const,
};

// ---- Queries ---------------------------------------------------------------

export function useProducts() {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.products,
    queryFn: () => apiRequest<Product[]>("/product/products", { locale }),
  });
}

export function useProduct(code: string) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.product(code),
    queryFn: () => apiRequest<Product>(`/product/products/${code}`, { locale }),
    enabled: !!code,
  });
}

export function useQuoteQuery(id: string | undefined) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.quote(id ?? ""),
    queryFn: () => apiRequest<Quote>(`/policy/quotes/${id}`, { locale }),
    enabled: !!id,
  });
}

export function usePolicies(limit = 50, offset = 0) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.policies(limit, offset),
    queryFn: () => apiRequest<Policy[]>(`/policy/policies?limit=${limit}&offset=${offset}`, { locale }),
  });
}

export function useClaims(status = "", limit = 50, offset = 0) {
  const locale = useLocale();
  const q = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  if (status) q.set("status", status);
  return useQuery({
    queryKey: qk.claims(status, limit, offset),
    queryFn: () => apiRequest<Claim[]>(`/claims/claims?${q.toString()}`, { locale }),
  });
}

export function useQuotes(limit = 50, offset = 0) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.quotes(limit, offset),
    queryFn: () => apiRequest<Quote[]>(`/policy/quotes?limit=${limit}&offset=${offset}`, { locale }),
  });
}

export function useParties(limit = 50, offset = 0) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.parties(limit, offset),
    queryFn: () => apiRequest<Party[]>(`/party/parties?limit=${limit}&offset=${offset}`, { locale }),
  });
}

export function useInvoiceByPolicy(policyId: string | undefined) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.invoiceByPolicy(policyId ?? ""),
    queryFn: () => apiRequest<Invoice>(`/billing/invoices/by-policy/${policyId}`, { locale }),
    enabled: !!policyId,
    retry: false,
  });
}

export function usePolicy(id: string | undefined) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.policy(id ?? ""),
    queryFn: () => apiRequest<Policy>(`/policy/policies/${id}`, { locale }),
    enabled: !!id,
  });
}

export function useInvoice(id: string | undefined) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.invoice(id ?? ""),
    queryFn: () => apiRequest<Invoice>(`/billing/invoices/${id}`, { locale }),
    enabled: !!id,
  });
}

export function useClaim(id: string | undefined) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.claim(id ?? ""),
    queryFn: () => apiRequest<Claim>(`/claims/claims/${id}`, { locale }),
    enabled: !!id,
  });
}

export function useKpis() {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.kpis,
    queryFn: () => apiRequest<KPI>("/reporting/kpis", { locale }),
  });
}

export function useAudit(limit = 30) {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.audit(limit),
    queryFn: () => apiRequest<Audit[]>(`/audit/logs?limit=${limit}`, { locale }),
    refetchInterval: 15_000,
  });
}

export function useUsers() {
  const locale = useLocale();
  return useQuery({
    queryKey: qk.users,
    queryFn: () => apiRequest<IamUser[]>("/iam/users", { locale }),
  });
}

export function useCreateUser() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: RegisterUserInput) =>
      apiRequest<{ id: string }>("/iam/users", { method: "POST", body, locale }),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.users }),
  });
}

// ---- Mutations -------------------------------------------------------------

export function useRegisterParty() {
  const locale = useLocale();
  return useMutation({
    mutationFn: (body: RegisterPartyInput) =>
      apiRequest<RegisterResponse>("/party/parties", { method: "POST", body, locale }),
  });
}

export function useCreateQuote() {
  const locale = useLocale();
  return useMutation({
    mutationFn: (body: CreateQuoteInput) =>
      apiRequest<Quote>("/policy/quotes", { method: "POST", body, locale }),
  });
}

export function useBindQuote() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (quoteId: string) =>
      apiRequest<Policy>(`/policy/quotes/${quoteId}/bind`, { method: "POST", body: {}, locale }),
    onSuccess: (policy) => {
      qc.setQueryData(qk.policy(policy.ID), policy);
      qc.invalidateQueries({ queryKey: qk.kpis });
      qc.invalidateQueries({ queryKey: ["policies"] });
    },
  });
}

export function useEndorsePolicy() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ policyId, deltaMinor, reason }: { policyId: string; deltaMinor: number; reason: string }) =>
      apiRequest<Policy>(`/policy/policies/${policyId}/endorse`, {
        method: "POST",
        body: { premium_delta_minor: deltaMinor, reason },
        locale,
      }),
    onSuccess: (p) => {
      qc.setQueryData(qk.policy(p.ID), p);
      qc.invalidateQueries({ queryKey: ["policies"] });
    },
  });
}

export function useCancelPolicy() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ policyId, reason }: { policyId: string; reason: string }) =>
      apiRequest<{ policy: Policy; refund_minor: number }>(`/policy/policies/${policyId}/cancel`, {
        method: "POST",
        body: { reason },
        locale,
      }),
    onSuccess: (res) => {
      qc.setQueryData(qk.policy(res.policy.ID), res.policy);
      qc.invalidateQueries({ queryKey: ["policies"] });
    },
  });
}

export function useRenewPolicy() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (policyId: string) =>
      apiRequest<Policy>(`/policy/policies/${policyId}/renew`, { method: "POST", body: {}, locale }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["policies"] });
      qc.invalidateQueries({ queryKey: qk.kpis });
    },
  });
}

export function useInitiatePayment() {
  const locale = useLocale();
  return useMutation({
    mutationFn: (invoiceId: string) =>
      apiRequest<PaymentIntent>(`/billing/invoices/${invoiceId}/pay-init`, { method: "POST", body: {}, locale }),
  });
}

export function useSubmitFNOL() {
  const locale = useLocale();
  return useMutation({
    mutationFn: (body: FNOLInput) =>
      apiRequest<Claim>("/claims/claims", { method: "POST", body, locale }),
  });
}

export function useSettleClaim() {
  const locale = useLocale();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ claimId, amountMinor }: { claimId: string; amountMinor: number }) =>
      apiRequest<Claim>(`/claims/claims/${claimId}/settle`, {
        method: "POST",
        body: { amount_minor: amountMinor },
        locale,
      }),
    onSuccess: (claim) => {
      qc.setQueryData(qk.claim(claim.ID), claim);
      qc.invalidateQueries({ queryKey: qk.kpis });
      qc.invalidateQueries({ queryKey: ["claims"] });
    },
  });
}
