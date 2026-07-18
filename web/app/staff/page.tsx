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
    <div className="max-w-7xl mx-auto px-6 py-12 space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
      
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-slate-200">
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 mb-4 rounded-full bg-slate-100 text-slate-700 text-xs font-bold tracking-wider uppercase border border-slate-200 shadow-sm">
            <svg className="w-4 h-4 text-brand-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
            System Governance
          </div>
          <h1 className="text-4xl font-display font-bold text-slate-900 tracking-tight">{locale === "am" ? "የሰራተኛ ዳሽቦርድ" : "Staff Administration"}</h1>
          <p className="mt-2 text-lg text-slate-500 max-w-2xl">{locale === "am" ? "የጋራ መሠረት ማረጋገጫ፣ ኦዲት እና KPI።" : "Shared-core operations, underwriting workbench, immutable audit trail, and real-time KPIs."}</p>
        </div>
        {lastPolicy && (
          <a href={`/policy/${lastPolicy.policyId}`} className="group flex items-center gap-3 bg-white border border-slate-200 hover:border-brand-blue-300 p-4 rounded-xl shadow-sm hover:shadow-md transition-all">
            <div className="w-10 h-10 rounded-full bg-emerald-50 flex items-center justify-center shrink-0">
              <svg className="w-5 h-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
            </div>
            <div>
              <div className="text-xs font-semibold uppercase text-slate-500 tracking-wider">Recent Issue</div>
              <div className="font-mono font-bold text-slate-900 group-hover:text-brand-blue-600 transition-colors">{lastPolicy.policyNumber}</div>
            </div>
          </a>
        )}
      </div>

      {err && (
        <div className="p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 flex items-start gap-3">
          <svg className="w-5 h-5 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <div className="text-sm font-medium">{err}</div>
        </div>
      )}

      {kpis && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <MetricCard title="Policies In Force" value={kpis.policiesInForce.toString()} icon="shield" />
          <MetricCard title="Gross Written Premium" value={formatETB(kpis.gwpMinor, locale)} icon="chart" />
          <MetricCard title="Claims Settled" value={kpis.claimsSettled.toString()} icon="check" />
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-8">
        
        {/* UNDERWRITING WORKBENCH */}
        <section className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-slate-100 bg-slate-50/50 flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-slate-900 font-display">Underwriting Referral</h2>
              <p className="text-sm text-slate-500 mt-1">Quotes awaiting manual underwriter approval</p>
            </div>
            <div className="px-3 py-1 bg-amber-100 text-amber-700 rounded-full text-xs font-bold">
              {referredQuotes.length} Pending
            </div>
          </div>
          <div className="p-0 flex-1 overflow-x-auto">
            {referredQuotes.length === 0 ? (
              <div className="p-12 text-center flex flex-col items-center justify-center text-slate-500 h-full">
                <svg className="w-12 h-12 text-slate-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"></path></svg>
                <p>No quotes pending review.</p>
              </div>
            ) : (
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-slate-500 uppercase bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4">Quote ID</th>
                    <th className="px-6 py-4">Flag</th>
                    <th className="px-6 py-4">Premium</th>
                    <th className="px-6 py-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {referredQuotes.map(q => (
                    <tr key={q.id} className="hover:bg-slate-50/80 transition-colors">
                      <td className="px-6 py-4 font-mono text-xs text-slate-600">{q.id.split("-")[0]}</td>
                      <td className="px-6 py-4"><span className="px-2.5 py-1 bg-red-50 text-red-700 rounded-md text-xs font-semibold border border-red-100">{q.uwDecision}</span></td>
                      <td className="px-6 py-4 font-mono font-medium">{(q.totalMinor / 100).toLocaleString()} {q.currency}</td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex justify-end gap-2">
                          <button className="bg-white hover:bg-red-50 text-red-600 border border-slate-200 hover:border-red-200 px-3 py-1.5 rounded-lg text-xs font-bold transition-colors" onClick={() => handleDecline(q.id)}>Decline</button>
                          <button className="bg-brand-blue-600 hover:bg-brand-blue-700 text-white px-3 py-1.5 rounded-lg text-xs font-bold transition-colors shadow-sm" onClick={() => handleApprove(q.id)}>Approve</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </section>

        {/* CLAIMS WORKBENCH */}
        <section className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-slate-100 bg-slate-50/50 flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-slate-900 font-display">Claims Workbench</h2>
              <p className="text-sm text-slate-500 mt-1">Manage reserves, recoveries, and settlements</p>
            </div>
            <div className="px-3 py-1 bg-brand-blue-50 text-brand-blue-700 rounded-full text-xs font-bold border border-blue-100">
              {claims.length} Active
            </div>
          </div>
          <div className="p-0 flex-1 overflow-x-auto">
            {claims.length === 0 ? (
              <div className="p-12 text-center flex flex-col items-center justify-center text-slate-500 h-full">
                <svg className="w-12 h-12 text-slate-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
                <p>No active claims in the system.</p>
              </div>
            ) : (
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-slate-500 uppercase bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4">Claim No</th>
                    <th className="px-6 py-4">Status / Track</th>
                    <th className="px-6 py-4">Financials (ETB)</th>
                    <th className="px-6 py-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {claims.map(c => (
                    <tr key={c.id} className="hover:bg-slate-50/80 transition-colors">
                      <td className="px-6 py-4 font-mono font-bold text-slate-900">{c.claimNumber}</td>
                      <td className="px-6 py-4">
                        <div className="flex flex-col gap-1 items-start">
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border ${c.status === "SETTLED" ? "bg-emerald-50 text-emerald-700 border-emerald-200" : "bg-amber-50 text-amber-700 border-amber-200"}`}>
                            {c.status}
                          </span>
                          <span className="text-xs text-slate-500 font-semibold">{c.track}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-xs">
                        <div className="grid grid-cols-2 gap-x-2 gap-y-1">
                          <span className="text-slate-500">Reserve:</span> <span className="font-mono font-medium text-slate-900">{(c.reserveMinor / 100).toLocaleString()}</span>
                          <span className="text-slate-500">Recovery:</span> <span className="font-mono font-medium text-emerald-600">{(c.recoveryMinor / 100).toLocaleString()}</span>
                          <span className="text-slate-500 font-bold">Settled:</span> <span className="font-mono font-bold text-slate-900">{c.settlementMinor ? (c.settlementMinor / 100).toLocaleString() : "-"}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-right">
                        {c.status !== "SETTLED" ? (
                          <div className="flex flex-col items-end gap-2">
                            <div className="flex gap-2">
                              <button className="bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 px-2 py-1 rounded text-xs font-semibold transition-colors" onClick={() => handleAdjustReserve(c.id)}>Reserve</button>
                              <button className="bg-white hover:bg-emerald-50 text-emerald-700 border border-slate-200 hover:border-emerald-200 px-2 py-1 rounded text-xs font-semibold transition-colors" onClick={() => handleRecordRecovery(c.id)}>Recovery</button>
                            </div>
                            <button className="bg-brand-blue-600 hover:bg-brand-blue-700 text-white px-4 py-1.5 rounded-lg text-xs font-bold transition-colors w-full sm:w-auto" onClick={() => handleSettleClaim(c.id, c.track)}>Settle Claim</button>
                          </div>
                        ) : (
                          <span className="text-slate-400 text-xs font-semibold italic">Closed</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </section>

      </div>

      {/* OPERATIONS ROW */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* EOD Reconciliation */}
        <div className="lg:col-span-1 space-y-6">
          <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-8">
            <div className="w-12 h-12 bg-indigo-50 text-indigo-600 rounded-xl flex items-center justify-center mb-6 border border-indigo-100">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
            </div>
            <h2 className="text-xl font-bold text-slate-900 font-display">End of Day (EOD)</h2>
            <p className="text-slate-500 mt-2 text-sm leading-relaxed mb-6">Trigger the nightly ERP batch job to reconcile today's Telebirr receipts with the General Ledger. This process ensures immutable financial consistency.</p>
            <button className="w-full btn btn-primary py-3" onClick={triggerEOD}>
              Run EOD Reconciliation
            </button>
            {reconciliationResult && (
              <div className="mt-6 p-4 bg-slate-900 rounded-xl overflow-hidden relative group">
                <div className="absolute top-2 right-2 flex space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <div className="w-2.5 h-2.5 rounded-full bg-red-500"></div>
                  <div className="w-2.5 h-2.5 rounded-full bg-yellow-500"></div>
                  <div className="w-2.5 h-2.5 rounded-full bg-green-500"></div>
                </div>
                <pre className="text-xs text-green-400 font-mono overflow-x-auto pt-2">
                  {JSON.stringify(reconciliationResult, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>

        {/* Audit Trail */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden flex flex-col h-full">
            <div className="p-6 border-b border-slate-100 bg-slate-50/50">
              <h2 className="text-xl font-bold text-slate-900 font-display">{locale === "am" ? "የኦዲት ዱካ" : "Immutable Audit Trail"}</h2>
            </div>
            <div className="p-0 flex-1 overflow-y-auto max-h-[400px]">
              <ul className="divide-y divide-slate-100">
                {audit.map((e) => (
                  <li key={e.id} className="p-4 hover:bg-slate-50/80 transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div>
                      <div className="font-mono text-sm font-bold text-slate-900">{e.entityType}.<span className="text-brand-blue-600">{e.action}</span></div>
                      <div className="text-xs text-slate-500 mt-1 truncate max-w-md">{e.detail || "System action recorded"}</div>
                    </div>
                    <div className="flex flex-col sm:items-end text-xs font-medium text-slate-400">
                      <span>{e.actor} · {e.entityId.substring(0, 8)}...</span>
                      <span>{new Date(e.at).toLocaleString()}</span>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

      </div>

      {/* Schema Reference */}
      <div className="bg-slate-900 rounded-2xl p-6 sm:p-8 relative overflow-hidden group border border-slate-800">
        <div className="absolute top-4 right-4 flex space-x-2 opacity-50 group-hover:opacity-100 transition-opacity">
          <div className="w-3 h-3 rounded-full bg-red-500"></div>
          <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
          <div className="w-3 h-3 rounded-full bg-green-500"></div>
        </div>
        <h2 className="text-lg font-bold text-slate-100 font-mono mb-2">{locale === "am" ? "የአደጋ ዕቅድ (shared-core)" : "system_schema://shared_core/risk_model"}</h2>
        <p className="text-slate-400 text-sm mb-6 font-mono">
          // {(schema?.note as string) ?? "Loading schema definition…"}
        </p>
        <pre className="text-xs text-indigo-300 font-mono overflow-x-auto p-4 bg-slate-950/50 rounded-xl border border-slate-800">
          {schema ? JSON.stringify(schema, null, 2) : "..."}
        </pre>
      </div>

    </div>
  );
}

function MetricCard({ title, value, icon }: { title: string; value: string; icon: string }) {
  const IconMap = {
    shield: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />,
    chart: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />,
    check: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
  };

  return (
    <div className="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
      <div className="absolute -right-4 -top-4 w-24 h-24 bg-slate-50 rounded-full group-hover:scale-150 transition-transform duration-700 ease-out z-0" />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-slate-500 uppercase tracking-wider">{title}</h3>
          <div className="p-2 bg-slate-50 rounded-lg text-slate-400 group-hover:text-brand-blue-600 transition-colors border border-slate-100">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {IconMap[icon as keyof typeof IconMap]}
            </svg>
          </div>
        </div>
        <div className="text-3xl font-bold text-slate-900 font-display mb-1">{value}</div>
      </div>
    </div>
  );
}
