"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/Shell";
import { MoneyInput } from "@/components/ui/MoneyInput";
import { api, errorMessage, type Claim } from "@/lib/api";
import { formatBirr, t } from "@/lib/i18n";

export default function ClaimPage() {
  const { locale } = useLocale();
  const [policyId, setPolicyId] = useState("");
  const [policyNumber, setPolicyNumber] = useState("");
  const [description, setDescription] = useState("Rear bumper scrape in Bole");
  const [reserveETB, setReserveETB] = useState(25000);
  const [settleETB, setSettleETB] = useState(25000);
  const [claim, setClaim] = useState<Claim | null>(null);
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const last = api.loadLastPolicy();
    if (last) {
      setPolicyId(last.policyId);
      setPolicyNumber(last.policyNumber);
    }
  }, []);

  async function submit() {
    setBusy(true); setErr("");
    try {
      const cl = await api.submitFNOL(locale, {
        policy_id: policyId,
        description,
        latitude: 8.9806,
        longitude: 38.7578,
        reserve_minor: Math.round(reserveETB * 100),
      });
      setClaim(cl);
      setSettleETB(cl.Reserve);
    } catch (e) {
      setErr(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  async function settle() {
    if (!claim) return;
    setBusy(true); setErr("");
    try {
      const updated = await api.settleClaim(locale, claim.ID, Math.round(settleETB * 100));
      setClaim(updated);
    } catch (e) {
      setErr(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="animate-rise mx-auto max-w-3xl px-6 py-12">

      <div className="mb-10 text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-blue-50 text-brand-blue-600 mb-6 border border-blue-100 shadow-sm">
          <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
        </div>
        <h1 className="text-4xl font-display font-bold text-slate-900 tracking-tight">{t("claimTitle", locale)}</h1>
        <p className="mt-3 text-lg text-slate-500 max-w-lg mx-auto">{t("claimSub", locale)}</p>
      </div>

      {err && (
        <div className="mb-8 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 flex items-start gap-3 animate-rise" role="alert">
          <svg className="w-5 h-5 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <div className="text-sm font-medium">{err}</div>
        </div>
      )}

      <div className="card overflow-hidden relative">
        {busy && (
          <div className="absolute inset-0 bg-white/60 backdrop-blur-sm z-50 flex flex-col items-center justify-center">
            <div className="w-10 h-10 border-4 border-brand-blue-600 border-t-transparent rounded-full animate-spin"></div>
            <p className="mt-4 font-semibold text-brand-blue-600 tracking-wide animate-pulse">Processing claim...</p>
          </div>
        )}

        <div className="p-8 sm:p-10">
          {!claim ? (
            <div className="space-y-6">

              {policyNumber && (
                <div className="p-4 bg-emerald-50 border border-emerald-100 rounded-xl flex items-center gap-3">
                  <span className="flex w-8 h-8 rounded-full bg-emerald-100 text-emerald-600 items-center justify-center shrink-0">
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" /></svg>
                  </span>
                  <div>
                    <span className="text-emerald-800 font-medium block text-sm">Active Policy Detected</span>
                    <span className="text-emerald-900 font-mono font-bold">{policyNumber}</span>
                  </div>
                </div>
              )}

              <div className="space-y-5">
                <div>
                  <label htmlFor="c-policy-id" className="label-text">{t("claimPolicyId", locale)}</label>
                  <input id="c-policy-id" className="input-field font-mono text-sm" value={policyId} onChange={(e) => setPolicyId(e.target.value)} placeholder="e.g. pol_123456789" />
                </div>

                <div>
                  <label htmlFor="c-description" className="label-text">{t("claimDescription", locale)}</label>
                  <textarea id="c-description" className="input-field resize-none" rows={4} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Describe what happened in detail..." />
                </div>

                <MoneyInput
                  label={t("claimAmount", locale)}
                  value={reserveETB}
                  min={0}
                  step={1000}
                  onChange={setReserveETB}
                />
              </div>

              <div className="pt-6 border-t border-slate-100 mt-8">
                <button className="btn btn-primary w-full sm:w-auto px-8" type="button" disabled={busy || !policyId} onClick={submit}>
                  {t("claimSubmit", locale)}
                </button>
              </div>

            </div>
          ) : (
            <div className="space-y-8 animate-rise">

              <div className="text-center">
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-50 text-brand-blue-600 mb-4 ring-8 ring-blue-50/50">
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                </div>
                <h3 className="text-2xl font-bold text-slate-900 font-display">Claim Registered</h3>
                <p className="text-slate-500 mt-1">Your First Notice of Loss (FNOL) has been recorded successfully.</p>
              </div>

              <div className="bg-slate-50 border border-slate-200 rounded-xl p-6">
                <div className="grid grid-cols-2 gap-y-6 gap-x-4">
                  <div>
                    <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Claim Reference</div>
                    <div className="text-lg font-mono font-bold text-slate-900">{claim.ID.split("-")[0]}</div>
                  </div>
                  <div>
                    <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Reserve</div>
                    <div className="text-lg font-mono font-bold text-slate-900">{formatBirr(claim.Reserve, locale)}</div>
                  </div>
                  <div className="col-span-2">
                    <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Current Status</div>
                    <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-bold border ${claim.Status === "SETTLED" ? "bg-emerald-100 text-emerald-700 border-emerald-200" : "bg-amber-100 text-amber-700 border-amber-200"}`}>
                      {claim.Status}
                    </div>
                  </div>
                </div>
              </div>

              {claim.Status !== "SETTLED" && (
                <div className="p-5 bg-white border border-brand-blue-200 shadow-sm shadow-brand-blue-100 rounded-xl space-y-4">
                  <div>
                    <h4 className="font-semibold text-slate-900 mb-1">Fast-Track Settlement</h4>
                    <p className="text-sm text-slate-600">Claims within the fast-track authority settle instantly; larger amounts are referred for manual review.</p>
                  </div>
                  <MoneyInput
                    label="Settlement amount"
                    value={settleETB}
                    min={0}
                    step={1000}
                    onChange={setSettleETB}
                  />
                  <button className="btn btn-primary w-full" type="button" disabled={busy} onClick={settle}>
                    {t("claimSettle", locale)}
                  </button>
                </div>
              )}

              {claim.Status === "SETTLED" && (
                <div className="p-6 bg-emerald-50 border border-emerald-200 rounded-xl text-center">
                  <div className="text-emerald-800 font-bold mb-1">{t("claimSettled", locale)}</div>
                  <div className="text-3xl font-mono font-bold text-emerald-600">{formatBirr(claim.SettledAmount, locale)}</div>
                  <p className="text-sm text-emerald-700 mt-2">Funds have been disbursed to your linked account.</p>
                </div>
              )}

            </div>
          )}
        </div>
      </div>
    </div>
  );
}
