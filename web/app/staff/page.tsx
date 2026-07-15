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

  useEffect(() => {
    (async () => {
      try {
        const [k, a, s] = await Promise.all([api.kpis(locale), api.audit(locale), api.riskSchema(locale)]);
        setKpis(k);
        setAudit(a);
        setSchema(s);
      } catch (e) {
        setErr(String(e));
      }
    })();
  }, [locale]);

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
