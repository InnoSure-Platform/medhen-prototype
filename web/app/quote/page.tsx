"use client";

import { useState } from "react";
import { useLocale } from "@/components/Shell";
import { MoneyInput } from "@/components/ui/MoneyInput";
import { api, errorMessage, type Policy, type Quote } from "@/lib/api";
import { formatBirr, t } from "@/lib/i18n";

type Step = "party" | "risk" | "quote" | "done";

const STEPS: { id: Step; label: string; number: number }[] = [
  { id: "party", label: "Identity", number: 1 },
  { id: "risk", label: "Asset details", number: 2 },
  { id: "quote", label: "Quotation", number: 3 },
  { id: "done", label: "Complete", number: 4 },
];

export default function QuotePage() {
  const { locale } = useLocale();
  const [step, setStep] = useState<Step>("party");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  const [partyId, setPartyId] = useState("");
  const [quote, setQuote] = useState<Quote | null>(null);
  const [policy, setPolicy] = useState<Policy | null>(null);

  const [form, setForm] = useState({
    fullName: "Abebe Kebede",
    fullNameAm: "አበበ ከበደ",
    phoneE164: "+251911234567",
    nationalId: "1234567890",
    plateNumber: "AA-3-12345",
    make: "Toyota",
    model: "Corolla",
    year: 2021,
    ageBand: "adult",
    coverType: "comprehensive",
    sumInsuredETB: 1500000,
  });

  async function registerAndContinue() {
    setBusy(true); setErr("");
    try {
      const party = await api.registerParty(locale, {
        full_name: form.fullName,
        full_name_amharic: form.fullNameAm,
        phone_e164: form.phoneE164,
        national_id: form.nationalId,
        address: { region: "Addis Ababa", zone: "Kirkos", woreda: "08", kebele: "15", city: "Addis Ababa", line1: "Bole Road" },
      });
      setPartyId(party.id);
      setStep("risk");
    } catch (e) {
      setErr(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  async function price() {
    setBusy(true); setErr("");
    try {
      const coverages = form.coverType === "comprehensive" ? ["OD", "TPL"] : ["TPL"];
      const q = await api.createQuote(locale, {
        party_id: partyId,
        product_code: "MOT",
        coverages,
        risk_dimensions: {
          age_band: form.ageBand,
          plate_number: form.plateNumber,
          make: form.make,
          model: form.model,
          year: String(form.year),
          usage: "private",
          sum_insured: String(form.sumInsuredETB),
        },
      });
      setQuote(q);
      setStep("quote");
    } catch (e) {
      setErr(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  async function bind() {
    if (!quote) return;
    setBusy(true); setErr("");
    try {
      const issued = await api.bindQuote(locale, quote.ID);
      setPolicy(issued);
      api.saveLastPolicy(issued.ID, issued.PolicyNumber);
      setStep("done");
    } catch (e) {
      setErr(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  const currentStepNumber = STEPS.find((s) => s.id === step)?.number || 1;

  return (
    <div className="max-w-4xl mx-auto px-6 py-12 animate-in fade-in slide-in-from-bottom-4 duration-700">

      {/* Header & Step Indicator */}
      <div className="mb-12">
        <h1 className="text-4xl font-display font-bold text-slate-900 mb-8 tracking-tight">{t("quoteTitle", locale)}</h1>

        <nav aria-label="Progress">
          <ol role="list" className="flex items-center">
            {STEPS.map((s, stepIdx) => {
              const isCompleted = s.number < currentStepNumber;
              const isCurrent = s.id === step;

              return (
                <li key={s.id} className={`relative ${stepIdx !== STEPS.length - 1 ? 'pr-8 sm:pr-20' : ''}`}>
                  {stepIdx !== STEPS.length - 1 && (
                    <div className="absolute inset-0 flex items-center" aria-hidden="true">
                      <div className={`h-0.5 w-full ${isCompleted ? 'bg-brand-blue-600' : 'bg-slate-200'}`} />
                    </div>
                  )}
                  <div className="relative flex h-8 items-center justify-center bg-white">
                    <span
                      className={`relative z-10 flex h-8 w-8 items-center justify-center rounded-full border-2 text-xs font-bold transition-colors ${
                        isCompleted
                          ? 'border-brand-blue-600 bg-brand-blue-600 text-white'
                          : isCurrent
                          ? 'border-brand-blue-600 bg-white text-brand-blue-600'
                          : 'border-slate-300 bg-white text-slate-400'
                      }`}
                    >
                      {isCompleted ? (
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" strokeWidth="3" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>
                      ) : (
                        s.number
                      )}
                    </span>
                    <span className={`absolute -bottom-6 w-max text-xs font-semibold ${isCurrent ? 'text-brand-blue-600' : 'text-slate-500 hidden sm:block'}`}>
                      {s.label}
                    </span>
                  </div>
                </li>
              );
            })}
          </ol>
        </nav>
      </div>

      {err && (
        <div className="mb-8 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 flex items-start gap-3 animate-in fade-in slide-in-from-top-2" role="alert">
          <svg className="w-5 h-5 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <div className="text-sm font-medium">{err}</div>
        </div>
      )}

      {/* Main Content Area */}
      <div className="bg-white border border-slate-200 shadow-xl rounded-2xl p-8 sm:p-10 relative overflow-hidden">
        {busy && (
          <div className="absolute inset-0 bg-white/60 backdrop-blur-sm z-50 flex flex-col items-center justify-center">
            <div className="w-10 h-10 border-4 border-brand-blue-600 border-t-transparent rounded-full animate-spin"></div>
            <p className="mt-4 font-semibold text-brand-blue-600 tracking-wide animate-pulse">Processing securely...</p>
          </div>
        )}

        {/* STEP 1: PARTY */}
        {step === "party" && (
          <div className="space-y-8 animate-in slide-in-from-right-4">
            <div>
              <h2 className="text-2xl font-bold text-slate-900 font-display">Personal Information</h2>
              <p className="text-slate-500 mt-1">Please provide your details to begin the quotation process.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label htmlFor="q-fullname" className="label-text">{t("fullName", locale)}</label>
                <input id="q-fullname" className="input-field" value={form.fullName} onChange={(e) => setForm({ ...form, fullName: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-fullname-am" className="label-text">{t("fullNameAm", locale)}</label>
                <input id="q-fullname-am" className="input-field font-sans" value={form.fullNameAm} onChange={(e) => setForm({ ...form, fullNameAm: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-phone" className="label-text">{t("phone", locale)}</label>
                <input id="q-phone" className="input-field font-mono" value={form.phoneE164} onChange={(e) => setForm({ ...form, phoneE164: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-national-id" className="label-text">Fayda / National ID</label>
                <input id="q-national-id" className="input-field font-mono tracking-widest" value={form.nationalId} maxLength={10} onChange={(e) => setForm({ ...form, nationalId: e.target.value })} />
              </div>
            </div>

            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy} onClick={registerAndContinue}>
                {t("quoteContinue", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 2: RISK */}
        {step === "risk" && (
          <div className="space-y-8 animate-in slide-in-from-right-4">
            <div>
              <h2 className="text-2xl font-bold text-slate-900 font-display">Asset Information</h2>
              <p className="text-slate-500 mt-1">Provide the details of the vehicle you want to insure.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label htmlFor="q-plate" className="label-text">{t("plate", locale)}</label>
                <input id="q-plate" className="input-field font-mono uppercase" value={form.plateNumber} onChange={(e) => setForm({ ...form, plateNumber: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-make" className="label-text">{t("make", locale)}</label>
                <input id="q-make" className="input-field" value={form.make} onChange={(e) => setForm({ ...form, make: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-model" className="label-text">{t("model", locale)}</label>
                <input id="q-model" className="input-field" value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} />
              </div>
              <div>
                <label htmlFor="q-year" className="label-text">{t("year", locale)}</label>
                <input id="q-year" className="input-field" type="number" value={form.year} onChange={(e) => setForm({ ...form, year: Number(e.target.value) })} />
              </div>
              <div>
                <label htmlFor="q-age-band" className="label-text">Driver age band</label>
                <select id="q-age-band" className="input-field bg-white" value={form.ageBand} onChange={(e) => setForm({ ...form, ageBand: e.target.value })}>
                  <option value="young">Young (18–25)</option>
                  <option value="adult">Adult (26–59)</option>
                  <option value="senior">Senior (60+)</option>
                </select>
              </div>
              <div>
                <label htmlFor="q-cover" className="label-text">{t("cover", locale)}</label>
                <select id="q-cover" className="input-field bg-white" value={form.coverType} onChange={(e) => setForm({ ...form, coverType: e.target.value })}>
                  <option value="comprehensive">{t("coverComprehensive", locale)}</option>
                  <option value="third_party">{t("coverThirdParty", locale)}</option>
                </select>
              </div>
              <div className="md:col-span-2">
                <MoneyInput
                  label={t("sumInsured", locale)}
                  value={form.sumInsuredETB}
                  min={1000}
                  step={10000}
                  onChange={(v) => setForm({ ...form, sumInsuredETB: v })}
                />
              </div>
            </div>

            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy} onClick={price}>
                {t("quoteCalculate", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 3: QUOTE SUMMARY */}
        {step === "quote" && quote && (
          <div className="space-y-8 animate-in slide-in-from-right-4">

            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 bg-emerald-50 border border-emerald-200 rounded-xl">
              <div>
                <div className="text-emerald-800 font-bold text-lg">{t("quoteStp", locale)}</div>
                <div className="text-emerald-600 text-sm mt-1">Status: <span className="font-semibold uppercase">{quote.Status}</span></div>
              </div>
              <div className="bg-white px-4 py-2 rounded-lg border border-emerald-100 shadow-sm font-mono font-bold text-xl text-slate-900">
                {formatBirr(quote.GrossPremium, locale)}
              </div>
            </div>

            <div className="border border-slate-200 rounded-xl overflow-hidden">
              <table className="w-full text-sm text-left">
                <thead className="bg-slate-50 text-slate-600 font-semibold border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4">{t("quoteLine", locale)}</th>
                    <th className="px-6 py-4 text-right">{t("quoteAmount", locale)}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  <tr className="hover:bg-slate-50/50">
                    <td className="px-6 py-4 text-slate-700">Net premium ({quote.Coverages.join(", ")})</td>
                    <td className="px-6 py-4 text-right font-mono font-medium text-slate-900">{formatBirr(quote.NetPremium, locale)}</td>
                  </tr>
                  <tr className="hover:bg-slate-50/50">
                    <td className="px-6 py-4 text-slate-700">Taxes &amp; duties</td>
                    <td className="px-6 py-4 text-right font-mono font-medium text-slate-900">{formatBirr(quote.TotalTaxes, locale)}</td>
                  </tr>
                  <tr className="bg-slate-50/50 font-bold text-slate-900">
                    <td className="px-6 py-4 uppercase tracking-wider">{t("quoteTotal", locale)}</td>
                    <td className="px-6 py-4 text-right text-lg">{formatBirr(quote.GrossPremium, locale)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy || quote.Status !== "QUOTED"} onClick={bind}>
                {t("quotePay", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 4: DONE / SUCCESS */}
        {step === "done" && policy && (
          <div className="space-y-8 animate-in slide-in-from-bottom-4">

            <div className="flex flex-col items-center justify-center py-8 text-center">
              <div className="w-20 h-20 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mb-6 ring-8 ring-emerald-50">
                <svg className="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="3" d="M5 13l4 4L19 7"></path></svg>
              </div>
              <h2 className="text-3xl font-display font-bold text-slate-900 mb-2">{t("quoteIssued", locale)}</h2>
              <p className="text-slate-500 max-w-md mx-auto">Your policy has been issued. An invoice for the premium has been raised — settle it via Telebirr to activate cover.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Policy Number</div>
                <div className="text-xl font-mono font-bold text-slate-900">{policy.PolicyNumber}</div>
              </div>
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Premium</div>
                <div className="text-xl font-mono font-bold text-slate-900">{formatBirr(policy.GrossPremium, locale)}</div>
              </div>
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Status</div>
                <div className="text-xl font-mono font-bold text-slate-900">{policy.Status}</div>
              </div>
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Cover period</div>
                <div className="text-sm font-mono font-bold text-slate-900">
                  {new Date(policy.EffectiveFrom).toLocaleDateString()} — {new Date(policy.EffectiveTo).toLocaleDateString()}
                </div>
              </div>
            </div>

            <div className="pt-8 flex justify-center gap-4">
              <a href="/claim" className="btn btn-ghost px-8">File a claim</a>
              <a href="/customer" className="btn btn-primary px-8">Return to Dashboard</a>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
