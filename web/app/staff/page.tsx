"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/Shell";
import { api, errorMessage, type Audit, type KPI } from "@/lib/api";
import { formatETB } from "@/lib/i18n";

export default function StaffPage() {
  const { locale } = useLocale();
  const [kpis, setKpis] = useState<KPI | null>(null);
  const [audit, setAudit] = useState<Audit[]>([]);
  const [err, setErr] = useState("");
  const [lastPolicy, setLastPolicy] = useState<{ policyId: string; policyNumber: string } | null>(null);

  useEffect(() => {
    setLastPolicy(api.loadLastPolicy());
    (async () => {
      try {
        const [k, a] = await Promise.all([api.kpis(locale), api.audit(locale)]);
        setKpis(k);
        setAudit(a);
      } catch (e) {
        setErr(errorMessage(e));
      }
    })();
  }, [locale]);

  const pct = (v: number) => `${(v * 100).toFixed(1)}%`;

  return (
    <div className="max-w-7xl mx-auto px-6 py-12 space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-700">

      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-slate-200">
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 mb-4 rounded-full bg-slate-100 text-slate-700 text-xs font-bold tracking-wider uppercase border border-slate-200 shadow-sm">
            <svg className="w-4 h-4 text-brand-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
            System Governance
          </div>
          <h1 className="text-4xl font-display font-bold text-slate-900 tracking-tight">{locale === "am" ? "የሰራተኛ ዳሽቦርድ" : "Staff Administration"}</h1>
          <p className="mt-2 text-lg text-slate-500 max-w-2xl">{locale === "am" ? "የጋራ መሠረት KPI እና የማይለወጥ ኦዲት።" : "Real-time production KPIs and the immutable audit trail."}</p>
        </div>
        {lastPolicy && (
          <a href={`/staff/policies/${lastPolicy.policyId}`} className="group flex items-center gap-3 bg-white border border-slate-200 hover:border-brand-blue-300 p-4 rounded-xl shadow-sm hover:shadow-md transition-all">
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
        <div className="p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 flex items-start gap-3" role="alert">
          <svg className="w-5 h-5 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <div className="text-sm font-medium">{err}</div>
        </div>
      )}

      {kpis && (
        <div className="grid grid-cols-2 lg:grid-cols-5 gap-6">
          <MetricCard title="Policies In Force" value={kpis.policy_count.toString()} icon="shield" />
          <MetricCard title="Gross Written Premium" value={formatETB(kpis.premium_written_minor, locale)} icon="chart" />
          <MetricCard title="Claims Paid" value={formatETB(kpis.claims_paid_minor, locale)} icon="check" />
          <MetricCard title="Loss Ratio" value={pct(kpis.loss_ratio)} icon="chart" />
          <MetricCard title="Combined Ratio" value={pct(kpis.combined_ratio)} icon="shield" />
        </div>
      )}

      {/* Audit Trail */}
      <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden flex flex-col">
        <div className="p-6 border-b border-slate-100 bg-slate-50/50">
          <h2 className="text-xl font-bold text-slate-900 font-display">{locale === "am" ? "የኦዲት ዱካ" : "Immutable Audit Trail"}</h2>
          <p className="text-sm text-slate-500 mt-1">Most recent domain events recorded across all bounded contexts.</p>
        </div>
        <div className="p-0 flex-1 overflow-y-auto max-h-[520px]">
          {audit.length === 0 ? (
            <div className="p-12 text-center text-slate-500">No audit events recorded yet.</div>
          ) : (
            <ul className="divide-y divide-slate-100">
              {audit.map((e) => (
                <li key={e.id} className="p-4 hover:bg-slate-50/80 transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="min-w-0">
                    <div className="font-mono text-sm font-bold text-brand-blue-600">{e.topic}</div>
                    <div className="text-xs text-slate-500 mt-1 truncate max-w-xl font-mono">{JSON.stringify(e.payload)}</div>
                  </div>
                  <div className="flex flex-col sm:items-end text-xs font-medium text-slate-400 shrink-0">
                    <span className="font-mono">{e.id.substring(0, 8)}…</span>
                    <span>{new Date(e.recorded_at).toLocaleString()}</span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
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
        <div className="text-2xl font-bold text-slate-900 font-display mb-1">{value}</div>
      </div>
    </div>
  );
}
