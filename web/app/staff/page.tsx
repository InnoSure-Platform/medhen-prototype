"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/Shell";
import { api, type Audit, type KPI } from "@/lib/api";
import { formatETB } from "@/lib/i18n";

export default function StaffPage() {
  const { locale } = useLocale();
  const [kpis, setKpis] = useState<KPI | null>(null);
  const [audit, setAudit] = useState<Audit[]>([]);
  const [schema, setSchema] = useState<Record<string, unknown> | null>(null);
  const [err, setErr] = useState("");
  const [lastPolicy, setLastPolicy] = useState<{ policyId: string; policyNumber: string } | null>(null);
  const [reconciliationResult, setReconciliationResult] = useState<any>(null);
  const [referredQuotes, setReferredQuotes] = useState<any[]>([]);
  const [claims, setClaims] = useState<any[]>([]);

  useEffect(() => {
    setLastPolicy(api.loadLastPolicy());
    (async () => {
      try {
        const [k, a, s] = await Promise.all([api.kpis(locale), api.audit(locale), api.riskSchema(locale)]);
        setKpis(k);
        setAudit(a);
        setSchema(s);
        const refs = await api.listReferredQuotes(locale);
        setReferredQuotes(refs);
        const clms = await api.listClaims(locale);
        setClaims(clms);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [locale]);

  async function triggerEOD() {
    try {
      const res = await api.runEodReconciliation(locale, new Date().toISOString().split("T")[0]);
      setReconciliationResult(res);
      alert("EOD Reconciliation completed successfully!");
    } catch (e) {
      setErr(String(e));
    }
  }

  async function handleApprove(id: string) {
    try {
      await api.approveQuote(locale, id);
      setReferredQuotes(referredQuotes.filter(q => q.id !== id));
    } catch (e) {
      setErr(String(e));
    }
  }

  async function handleDecline(id: string) {
    try {
      await api.declineQuote(locale, id);
      setReferredQuotes(referredQuotes.filter(q => q.id !== id));
    } catch (e) {
      setErr(String(e));
    }
  }

  async function handleAdjustReserve(id: string) {
    const amount = prompt("Enter new reserve amount (in ETB):");
    if (!amount) return;
    try {
      await api.adjustClaimReserve(locale, id, parseInt(amount) * 100);
      setClaims(await api.listClaims(locale));
    } catch (e) {
      setErr(String(e));
    }
  }

  async function handleRecordRecovery(id: string) {
    const amount = prompt("Enter recovery amount (in ETB):");
    if (!amount) return;
    try {
      await api.recordClaimRecovery(locale, id, parseInt(amount) * 100);
      setClaims(await api.listClaims(locale));
    } catch (e) {
      setErr(String(e));
    }
  }

  async function handleSettleClaim(id: string, track: string) {
    const amount = prompt(`Enter final settlement amount (in ETB) for this ${track} claim:`);
    if (!amount) return;
    try {
      await api.settleClaim(locale, id, parseInt(amount) * 100);
      setClaims(await api.listClaims(locale));
    } catch (e) {
      setErr(String(e));
    }
  }

  return (
    <div className="section">
      <h1>{locale === "am" ? "የሰራተኛ ዳሽቦርድ" : "Staff · governance"}</h1>
      <p>{locale === "am" ? "የጋራ መሠረት ማረጋገጫ፣ ኦዲት እና KPI።" : "Shared-core proof, immutable audit, and KPI tile."}</p>
      {err && <div className="banner-err">{err}</div>}

      {kpis && (
        <div className="kpi-row" style={{ marginTop: "1.5rem" }}>
          <div className="kpi">
            <div className="label">Policies in force</div>
            <div className="value">{kpis.policiesInForce}</div>
          </div>
          <div className="kpi">
            <div className="label">GWP</div>
            <div className="value">{formatETB(kpis.gwpMinor, locale)}</div>
          </div>
          <div className="kpi">
            <div className="label">Claims settled</div>
            <div className="value">{kpis.claimsSettled}</div>
          </div>
        </div>
      )}

      {lastPolicy && (
        <div className="card" style={{ marginTop: "1rem" }}>
          <p>You recently bound a policy.</p>
          <a href={`/policy/${lastPolicy.policyId}`} className="btn" style={{ marginTop: "1rem", display: "inline-block", textDecoration: "none" }}>
            View Policy Details ({lastPolicy.policyNumber})
          </a>
        </div>
      )}

      <h2 style={{ marginTop: "2rem" }}>Underwriting Referral Workbench</h2>
      <div className="card" style={{ marginTop: "1rem", overflowX: "auto" }}>
        {referredQuotes.length === 0 ? (
          <p>No quotes pending review.</p>
        ) : (
          <table className="lines" style={{ width: "100%" }}>
            <thead>
              <tr>
                <th>Quote ID</th>
                <th>Reason</th>
                <th>Premium</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {referredQuotes.map(q => (
                <tr key={q.id}>
                  <td style={{ fontSize: "0.85rem", color: "#666" }}>{q.id.split("-")[0]}...</td>
                  <td><span className="banner-err" style={{ padding: "0.2rem 0.5rem", display: "inline-block", fontSize: "0.8rem", margin: 0 }}>{q.uwDecision}</span></td>
                  <td>{(q.totalMinor / 100).toFixed(2)} {q.currency}</td>
                  <td>
                    <div style={{ display: "flex", gap: "0.5rem" }}>
                      <button className="btn" style={{ background: "#4caf50", color: "#fff", padding: "0.3rem 0.6rem" }} onClick={() => handleApprove(q.id)}>Approve</button>
                      <button className="btn" style={{ background: "#f44336", color: "#fff", padding: "0.3rem 0.6rem" }} onClick={() => handleDecline(q.id)}>Decline</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <h2 style={{ marginTop: "2rem" }}>Claims Workbench</h2>
      <div className="card" style={{ marginTop: "1rem", overflowX: "auto" }}>
        {claims.length === 0 ? (
          <p>No claims in the system.</p>
        ) : (
          <table className="lines" style={{ width: "100%" }}>
            <thead>
              <tr>
                <th>Claim No</th>
                <th>Track</th>
                <th>Status</th>
                <th>Reserves</th>
                <th>Recoveries</th>
                <th>Settlement</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {claims.map(c => (
                <tr key={c.id}>
                  <td>{c.claimNumber}</td>
                  <td>{c.track}</td>
                  <td>
                    <span className={c.status === "SETTLED" ? "banner-ok" : "banner"} style={{ padding: "0.2rem 0.5rem", display: "inline-block", fontSize: "0.8rem", margin: 0 }}>
                      {c.status}
                    </span>
                  </td>
                  <td>{(c.reserveMinor / 100).toFixed(2)}</td>
                  <td>{(c.recoveryMinor / 100).toFixed(2)}</td>
                  <td>{c.settlementMinor ? (c.settlementMinor / 100).toFixed(2) : "-"}</td>
                  <td>
                    <div style={{ display: "flex", gap: "0.5rem" }}>
                      {c.status !== "SETTLED" && (
                        <>
                          <button className="btn" style={{ background: "#ff9800", color: "#fff", padding: "0.3rem 0.6rem", fontSize: "0.8rem" }} onClick={() => handleAdjustReserve(c.id)}>Reserve</button>
                          <button className="btn" style={{ background: "#2196f3", color: "#fff", padding: "0.3rem 0.6rem", fontSize: "0.8rem" }} onClick={() => handleRecordRecovery(c.id)}>Recovery</button>
                          <button className="btn" style={{ background: "#4caf50", color: "#fff", padding: "0.3rem 0.6rem", fontSize: "0.8rem" }} onClick={() => handleSettleClaim(c.id, c.track)}>Settle</button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <h2 style={{ marginTop: "2rem" }}>End of Day (EOD) Operations</h2>
      <div className="card" style={{ marginTop: "1rem" }}>
        <p>Trigger the nightly ERP batch job to reconcile today's Telebirr receipts with the General Ledger.</p>
        <button className="btn" onClick={triggerEOD} style={{ marginTop: "1rem" }}>
          Run EOD Reconciliation
        </button>
        {reconciliationResult && (
          <pre style={{ overflow: "auto", background: "rgba(255,255,255,0.5)", padding: "1rem", fontSize: "0.85rem", marginTop: "1rem" }}>
            {JSON.stringify(reconciliationResult, null, 2)}
          </pre>
        )}
      </div>

      <h2 style={{ marginTop: "2rem" }}>{locale === "am" ? "የአደጋ ዕቅድ (shared-core)" : "Motor risk schema (shared-core)"}</h2>
      <p style={{ color: "#4a5c53" }}>
        {(schema?.note as string) ?? "Loading…"}
      </p>
      <pre style={{ overflow: "auto", background: "rgba(255,255,255,0.5)", padding: "1rem", fontSize: "0.85rem" }}>
        {schema ? JSON.stringify(schema, null, 2) : "…"}
      </pre>

      <h2 style={{ marginTop: "2rem" }}>{locale === "am" ? "የኦዲት ዱካ" : "Audit trail"}</h2>
      <ul className="audit-list">
        {audit.map((e) => (
          <li key={e.id}>
            <strong>{e.entityType}.{e.action}</strong>
            <span className="meta">{e.entityId} · {e.actor} · {new Date(e.at).toLocaleString()} {e.detail ? `· ${e.detail}` : ""}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
