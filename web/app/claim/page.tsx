"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/Shell";
import { api, type Claim } from "@/lib/api";
import { formatETB, t } from "@/lib/i18n";

export default function ClaimPage() {
  const { locale } = useLocale();
  const [policyId, setPolicyId] = useState("");
  const [policyNumber, setPolicyNumber] = useState("");
  const [description, setDescription] = useState("Rear bumper scrape in Bole");
  const [amount, setAmount] = useState(25000);
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
        policyId,
        lossDate: new Date().toISOString(),
        description,
        latitude: 8.9806,
        longitude: 38.7578,
        estimatedAmountMinor: Math.round(amount * 100),
        photoObjectKeys: ["fnol/demo.jpg"],
      });
      setClaim(cl);
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  async function settle() {
    if (!claim) return;
    setBusy(true); setErr("");
    try {
      setClaim(await api.settle(locale, claim.id));
    } catch (e) {
      setErr(String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="section">
      <h1>{t("claimTitle", locale)}</h1>
      <p>{t("claimSub", locale)}</p>
      {err && <div className="banner-err">{err}</div>}
      <div className="stack" style={{ marginTop: "1.25rem" }}>
        {policyNumber && (
          <div className="banner-ok">{locale === "am" ? "ፖሊሲ" : "Policy"}: {policyNumber}</div>
        )}
        <div className="field">
          <label>{t("claimPolicyId", locale)}</label>
          <input value={policyId} onChange={(e) => setPolicyId(e.target.value)} placeholder="From quote flow or paste ID" />
        </div>
        <div className="field">
          <label>{t("claimDescription", locale)}</label>
          <textarea rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        <div className="field">
          <label>{t("claimAmount", locale)}</label>
          <input type="number" value={amount} onChange={(e) => setAmount(Number(e.target.value))} />
        </div>
        <button className="btn btn-primary" type="button" disabled={busy || !policyId} onClick={submit}>
          {busy ? "…" : t("claimSubmit", locale)}
        </button>
        {claim && (
          <div className="banner-ok">
            <div>{claim.claimNumber} · {claim.track} · {claim.status}</div>
            {claim.status !== "SETTLED" && claim.track === "FAST_TRACK" && (
              <button className="btn btn-primary" style={{ marginTop: "0.75rem" }} type="button" disabled={busy} onClick={settle}>
                {t("claimSettle", locale)}
              </button>
            )}
            {claim.status === "SETTLED" && (
              <div style={{ marginTop: "0.5rem" }}>
                {t("claimSettled", locale)} · {formatETB(claim.settlementMinor ?? 0, locale)}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
