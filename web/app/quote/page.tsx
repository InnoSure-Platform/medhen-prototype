"use client";

import { useState } from "react";
import { useLocale } from "@/components/Shell";
import { api, docLabel, type PaymentResult, type Quote } from "@/lib/api";
import { formatETB, t } from "@/lib/i18n";

type Step = "party" | "kyc" | "risk" | "quote" | "pay" | "done";

const STEPS: { id: Step; label: string; number: number }[] = [
  { id: "party", label: "Identity", number: 1 },
  { id: "kyc", label: "Verification", number: 2 },
  { id: "risk", label: "Asset details", number: 3 },
  { id: "quote", label: "Quotation", number: 4 },
  { id: "pay", label: "Payment", number: 5 },
  { id: "done", label: "Complete", number: 6 },
];

export default function QuotePage() {
  const { locale } = useLocale();
  const [step, setStep] = useState<Step>("party");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  
  // State
  const [partyId, setPartyId] = useState("");
  const [faydaId, setFaydaId] = useState("1234567890");
  const [phone, setPhone] = useState("+251911234567");
  const [installmentPlan, setInstallmentPlan] = useState("100_UPFRONT");
  const [quote, setQuote] = useState<Quote | null>(null);
  const [invoiceId, setInvoiceId] = useState("");
  const [result, setResult] = useState<PaymentResult | null>(null);
  
  const [form, setForm] = useState({
    fullName: "Abebe Kebede",
    fullNameAm: "አበበ ከበደ",
    phoneE164: "+251911234567",
    plateNumber: "AA-3-12345",
    make: "Toyota",
    model: "Corolla",
    year: 2021,
    coverType: "comprehensive",
    sumInsuredETB: 1500000,
  });

  async function registerAndContinue() {
    setBusy(true); setErr("");
    try {
      const party = await api.registerParty(locale, {
        fullName: form.fullName,
        fullNameAm: form.fullNameAm,
        phoneE164: form.phoneE164,
        address: { region: "Addis Ababa", zone: "Kirkos", woreda: "08", kebele: "15", city: "Addis Ababa", line1: "Bole Road" },
      });
      setPartyId(party.id);
      setPhone(form.phoneE164);
      setStep("kyc");
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function verifyIdentity() {
    setBusy(true); setErr("");
    try {
      await api.verifyKYC(locale, partyId, faydaId);
      setStep("risk");
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function price() {
    setBusy(true); setErr("");
    try {
      const q = await api.createQuote(locale, {
        partyId,
        productCode: "MOTOR-PRIVATE-COMP",
        risk: {
          plateNumber: form.plateNumber,
          make: form.make,
          model: form.model,
          year: form.year,
          usage: "private",
          coverType: form.coverType,
          sumInsuredMinor: Math.round(form.sumInsuredETB * 100),
        },
      });
      setQuote(q);
      setStep("quote");
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function bindAndPay() {
    if (!quote) return;
    setBusy(true); setErr("");
    try {
      const bind = await api.bindQuote(locale, quote.id, installmentPlan);
      setInvoiceId(bind.invoice.id);
      const pay = await api.pay(locale, bind.invoice.id, phone);
      setResult(pay);
      api.saveLastPolicy(pay.policy.id, pay.policy.policyNumber);
      setStep("done");
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  function downloadAll() {
    result?.documents.forEach((d) => {
      window.open(api.fileUrl(d.url), "_blank", "noopener,noreferrer");
    });
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
                  {/* Line connecting steps */}
                  {stepIdx !== STEPS.length - 1 && (
                    <div className="absolute inset-0 flex items-center" aria-hidden="true">
                      <div className={`h-0.5 w-full ${isCompleted ? 'bg-brand-blue-600' : 'bg-slate-200'}`} />
                    </div>
                  )}
                  {/* Step circle */}
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
        <div className="mb-8 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 flex items-start gap-3 animate-in fade-in slide-in-from-top-2">
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
                <label className="label-text">{t("fullName", locale)}</label>
                <input className="input-field" value={form.fullName} onChange={(e) => setForm({ ...form, fullName: e.target.value })} />
              </div>
              <div>
                <label className="label-text">{t("fullNameAm", locale)}</label>
                <input className="input-field font-sans" value={form.fullNameAm} onChange={(e) => setForm({ ...form, fullNameAm: e.target.value })} />
              </div>
              <div className="md:col-span-2">
                <label className="label-text">{t("phone", locale)}</label>
                <input className="input-field font-mono" value={form.phoneE164} onChange={(e) => setForm({ ...form, phoneE164: e.target.value })} />
              </div>
            </div>
            
            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy} onClick={registerAndContinue}>
                {t("quoteContinue", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 2: KYC */}
        {step === "kyc" && (
          <div className="space-y-8 animate-in slide-in-from-right-4">
            <div className="flex items-start gap-5">
              <div className="w-12 h-12 rounded-xl bg-blue-50 text-brand-blue-600 flex items-center justify-center shrink-0 border border-blue-100">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2" /></svg>
              </div>
              <div>
                <h2 className="text-2xl font-bold text-slate-900 font-display">Fayda Identity Verification</h2>
                <p className="text-slate-500 mt-1 leading-relaxed">
                  As part of the National ID initiative, we require your 10-digit Fayda ID to instantly verify your identity against the central registry.
                </p>
              </div>
            </div>
            
            <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
              <label className="label-text">Fayda ID Number</label>
              <input className="input-field text-lg font-mono tracking-widest text-center" value={faydaId} onChange={(e) => setFaydaId(e.target.value)} placeholder="1234567890" maxLength={10} />
            </div>
            
            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy} onClick={verifyIdentity}>
                Verify & Continue &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 3: RISK */}
        {step === "risk" && (
          <div className="space-y-8 animate-in slide-in-from-right-4">
            <div>
              <h2 className="text-2xl font-bold text-slate-900 font-display">Asset Information</h2>
              <p className="text-slate-500 mt-1">Provide the details of the vehicle you want to insure.</p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="label-text">{t("plate", locale)}</label>
                <input className="input-field font-mono uppercase" value={form.plateNumber} onChange={(e) => setForm({ ...form, plateNumber: e.target.value })} />
              </div>
              <div>
                <label className="label-text">{t("make", locale)}</label>
                <input className="input-field" value={form.make} onChange={(e) => setForm({ ...form, make: e.target.value })} />
              </div>
              <div>
                <label className="label-text">{t("model", locale)}</label>
                <input className="input-field" value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} />
              </div>
              <div>
                <label className="label-text">{t("year", locale)}</label>
                <input className="input-field" type="number" value={form.year} onChange={(e) => setForm({ ...form, year: Number(e.target.value) })} />
              </div>
              <div>
                <label className="label-text">{t("cover", locale)}</label>
                <select className="input-field bg-white" value={form.coverType} onChange={(e) => setForm({ ...form, coverType: e.target.value })}>
                  <option value="comprehensive">{t("coverComprehensive", locale)}</option>
                  <option value="third_party">{t("coverThirdParty", locale)}</option>
                </select>
              </div>
              <div>
                <label className="label-text">{t("sumInsured", locale)}</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <span className="text-slate-500 font-semibold text-sm">ETB</span>
                  </div>
                  <input className="input-field pl-12 font-mono font-medium" type="number" value={form.sumInsuredETB} onChange={(e) => setForm({ ...form, sumInsuredETB: Number(e.target.value) })} />
                </div>
              </div>
            </div>
            
            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy} onClick={price}>
                {t("quoteCalculate", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 4: QUOTE SUMMARY */}
        {step === "quote" && quote && (
          <div className="space-y-8 animate-in slide-in-from-right-4">
            
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 bg-emerald-50 border border-emerald-200 rounded-xl">
              <div>
                <div className="text-emerald-800 font-bold text-lg">{t("quoteStp", locale)}</div>
                <div className="text-emerald-600 text-sm mt-1">Decision: <span className="font-semibold uppercase">{quote.uwDecision}</span></div>
              </div>
              <div className="bg-white px-4 py-2 rounded-lg border border-emerald-100 shadow-sm font-mono font-bold text-xl text-slate-900">
                {formatETB(quote.totalMinor, locale)}
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
                  {quote.lines.map((l) => (
                    <tr key={l.code} className="hover:bg-slate-50/50">
                      <td className="px-6 py-4 text-slate-700">{locale === "am" ? l.labelAm : l.label}</td>
                      <td className="px-6 py-4 text-right font-mono font-medium text-slate-900">{formatETB(l.amountMinor, locale)}</td>
                    </tr>
                  ))}
                  <tr className="bg-slate-50/50 font-bold text-slate-900">
                    <td className="px-6 py-4 uppercase tracking-wider">{t("quoteTotal", locale)}</td>
                    <td className="px-6 py-4 text-right text-lg">{formatETB(quote.totalMinor, locale)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {quote.status === "REFERRED" ? (
              <div className="p-5 rounded-xl bg-amber-50 border border-amber-200 text-amber-800 flex items-start gap-3">
                <svg className="w-6 h-6 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                <div>
                  <strong className="block mb-1">Pending Underwriting Review</strong>
                  Your quote has been referred to an underwriter due to specific risk factors. You will be able to proceed with payment once it is approved.
                </div>
              </div>
            ) : (
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <label className="label-text mb-3 text-base">Select Payment Plan</label>
                <select className="input-field bg-white text-base py-3" value={installmentPlan} onChange={(e) => setInstallmentPlan(e.target.value)}>
                  <option value="100_UPFRONT">Pay 100% Upfront</option>
                  <option value="40_30_30">Installments (40% Now, 30% in 30 Days, 30% in 60 Days)</option>
                </select>
                {installmentPlan === "40_30_30" && (
                  <div className="mt-4 p-4 bg-brand-blue-50 border border-brand-blue-100 rounded-lg flex justify-between items-center text-sm font-medium">
                    <span className="text-brand-blue-800">Due Today (40% Downpayment)</span>
                    <span className="text-brand-blue-900 font-mono font-bold text-lg">{formatETB(quote.totalMinor * 0.4, locale)}</span>
                  </div>
                )}
              </div>
            )}

            <div className="pt-4 border-t border-slate-100 flex justify-end">
              <button className="btn btn-primary px-8" type="button" disabled={busy || quote.status === "REFERRED"} onClick={bindAndPay}>
                {t("quotePay", locale)} &rarr;
              </button>
            </div>
          </div>
        )}

        {/* STEP 5: DONE / SUCCESS */}
        {step === "done" && result && (
          <div className="space-y-8 animate-in slide-in-from-bottom-4">
            
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <div className="w-20 h-20 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mb-6 ring-8 ring-emerald-50">
                <svg className="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="3" d="M5 13l4 4L19 7"></path></svg>
              </div>
              <h2 className="text-3xl font-display font-bold text-slate-900 mb-2">Policy Issued Successfully</h2>
              <p className="text-slate-500 max-w-md mx-auto">Your payment was confirmed via Telebirr. Your policy documents have been generated and securely stored.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Policy Number</div>
                <div className="text-xl font-mono font-bold text-slate-900">{result.policy.policyNumber}</div>
              </div>
              <div className="bg-slate-50 p-6 rounded-xl border border-slate-200">
                <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Telebirr Receipt ID</div>
                <div className="text-xl font-mono font-bold text-slate-900">{result.receiptId}</div>
              </div>
            </div>

            <div className="mt-8">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-bold text-slate-900 font-display">Policy Documents</h3>
                <button type="button" className="text-brand-blue-600 font-semibold text-sm hover:text-brand-blue-800 transition-colors" onClick={downloadAll}>
                  Download All .PDFs
                </button>
              </div>
              
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {result.documents.map((d) => (
                  <a key={d.id} href={api.fileUrl(d.url)} target="_blank" rel="noreferrer" download className="flex items-start gap-4 p-4 rounded-xl border border-slate-200 hover:border-brand-blue-300 hover:shadow-md transition-all group bg-white">
                    <div className="w-10 h-10 rounded-lg bg-red-50 text-red-500 flex items-center justify-center shrink-0">
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="font-semibold text-slate-900 truncate group-hover:text-brand-blue-600 transition-colors">
                        {docLabel(d.type, locale)}
                      </h4>
                      <p className="text-xs font-medium text-slate-500 mt-0.5 uppercase tracking-wider">
                        {d.locale === "am" ? "Amharic" : "English"} · PDF
                      </p>
                    </div>
                    <svg className="w-5 h-5 text-slate-300 group-hover:text-brand-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
                  </a>
                ))}
              </div>
            </div>

            <div className="pt-8 flex justify-center">
              <a href="/customer" className="btn btn-ghost px-8">Return to Dashboard</a>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
