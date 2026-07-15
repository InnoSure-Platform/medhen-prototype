"use client";

import { useState } from "react";
import { useLocale } from "@/components/Shell";
import { api, docLabel, type PaymentResult, type Quote } from "@/lib/api";
import { formatETB, t } from "@/lib/i18n";

type Step = "party" | "risk" | "quote" | "pay" | "done";

export default function QuotePage() {
  const { locale } = useLocale();
  const [step, setStep] = useState<Step>("party");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const [partyId, setPartyId] = useState("");
  const [phone, setPhone] = useState("+251911234567");
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
      const bind = await api.bindQuote(locale, quote.id);
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

  return (
    <div className="section">
      <h1>{t("quoteTitle", locale)}</h1>
      <div className="steps">
        {(["party", "risk", "quote", "pay", "done"] as Step[]).map((s) => (
          <div key={s} className={`step ${step === s ? "on" : ""}`}>{s}</div>
        ))}
      </div>
      {err && <div className="banner-err">{err}</div>}

      {step === "party" && (
        <div className="stack">
          <div className="grid-2">
            <div className="field"><label>{t("fullName", locale)}</label><input value={form.fullName} onChange={(e) => setForm({ ...form, fullName: e.target.value })} /></div>
            <div className="field"><label>{t("fullNameAm", locale)}</label><input value={form.fullNameAm} onChange={(e) => setForm({ ...form, fullNameAm: e.target.value })} /></div>
            <div className="field"><label>{t("phone", locale)}</label><input value={form.phoneE164} onChange={(e) => setForm({ ...form, phoneE164: e.target.value })} /></div>
          </div>
          <button className="btn btn-primary" type="button" disabled={busy} onClick={registerAndContinue}>
            {busy ? "…" : t("quoteContinue", locale)}
          </button>
        </div>
      )}

      {step === "risk" && (
        <div className="stack">
          <div className="grid-2">
            <div className="field"><label>{t("plate", locale)}</label><input value={form.plateNumber} onChange={(e) => setForm({ ...form, plateNumber: e.target.value })} /></div>
            <div className="field"><label>{t("make", locale)}</label><input value={form.make} onChange={(e) => setForm({ ...form, make: e.target.value })} /></div>
            <div className="field"><label>{t("model", locale)}</label><input value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} /></div>
            <div className="field"><label>{t("year", locale)}</label><input type="number" value={form.year} onChange={(e) => setForm({ ...form, year: Number(e.target.value) })} /></div>
            <div className="field">
              <label>{t("cover", locale)}</label>
              <select value={form.coverType} onChange={(e) => setForm({ ...form, coverType: e.target.value })}>
                <option value="comprehensive">{t("coverComprehensive", locale)}</option>
                <option value="third_party">{t("coverThirdParty", locale)}</option>
              </select>
            </div>
            <div className="field"><label>{t("sumInsured", locale)}</label><input type="number" value={form.sumInsuredETB} onChange={(e) => setForm({ ...form, sumInsuredETB: Number(e.target.value) })} /></div>
          </div>
          <button className="btn btn-primary" type="button" disabled={busy} onClick={price}>
            {busy ? "…" : t("quoteCalculate", locale)}
          </button>
        </div>
      )}

      {step === "quote" && quote && (
        <div className="stack">
          <div className="banner-ok">{t("quoteStp", locale)}: <strong>{quote.uwDecision}</strong></div>
          <table className="lines">
            <thead><tr><th>{t("quoteLine", locale)}</th><th>{t("quoteAmount", locale)}</th></tr></thead>
            <tbody>
              {quote.lines.map((l) => (
                <tr key={l.code}>
                  <td>{locale === "am" ? l.labelAm : l.label}</td>
                  <td>{formatETB(l.amountMinor, locale)}</td>
                </tr>
              ))}
              <tr className="total-row">
                <td>{t("quoteTotal", locale)}</td>
                <td>{formatETB(quote.totalMinor, locale)}</td>
              </tr>
            </tbody>
          </table>
          <button className="btn btn-primary" type="button" disabled={busy} onClick={bindAndPay}>
            {busy ? "…" : t("quotePay", locale)}
          </button>
        </div>
      )}

      {step === "done" && result && (
        <div className="stack">
          <div className="banner-ok">
            {t("quoteIssued", locale)}: <strong>{result.policy.policyNumber}</strong>
            <div>Receipt {result.receiptId}</div>
          </div>
          <div className="doc-header">
            <h2>{t("quoteDocs", locale)}</h2>
            <button type="button" className="btn btn-ghost" onClick={downloadAll}>{t("quoteDownloadAll", locale)}</button>
          </div>
          <div className="doc-grid">
            {result.documents.map((d) => (
              <a key={d.id} className="doc-card" href={api.fileUrl(d.url)} target="_blank" rel="noreferrer" download>
                <span className="doc-type">{docLabel(d.type, locale)}</span>
                <span className="doc-locale">{d.locale === "am" ? "አማርኛ" : "English"}</span>
                <span className="doc-action">{t("quoteDownload", locale)} · PDF</span>
              </a>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
