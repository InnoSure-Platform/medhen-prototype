"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useLocale } from "@/components/Shell";
import { StatusBadge } from "@/components/ui/Badge";
import { PageHeader, Skeleton } from "@/components/ui/PageHeader";
import { StatCard } from "@/components/ui/StatCard";
import { api, errorMessage, type Audit, type KPI } from "@/lib/api";
import { formatETB } from "@/lib/i18n";

export default function StaffPage() {
  const { locale } = useLocale();
  const am = locale === "am";
  const [kpis, setKpis] = useState<KPI | null>(null);
  const [audit, setAudit] = useState<Audit[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [lastPolicy, setLastPolicy] = useState<{ policyId: string; policyNumber: string } | null>(null);

  useEffect(() => {
    setLastPolicy(api.loadLastPolicy());
    let alive = true;
    setLoading(true);
    (async () => {
      try {
        const [k, a] = await Promise.all([api.kpis(locale), api.audit(locale)]);
        if (!alive) return;
        setKpis(k);
        setAudit(a);
        setErr("");
      } catch (e) {
        if (alive) setErr(errorMessage(e));
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, [locale]);

  const pct = (v: number) => `${(v * 100).toFixed(1)}%`;
  const lossTone = kpis && kpis.loss_ratio > 1 ? "danger" : kpis && kpis.loss_ratio > 0.7 ? "warning" : "success";

  return (
    <div className="animate-rise mx-auto max-w-7xl space-y-10 px-6 py-12">
      <PageHeader
        eyebrow={
          <span className="eyebrow">
            <ShieldIcon /> {am ? "የስርዓት አስተዳደር" : "System governance"}
          </span>
        }
        title={am ? "የሰራተኛ ዳሽቦርድ" : "Staff dashboard"}
        subtitle={am ? "ቀጥታ የምርት KPI እና የማይለወጥ ኦዲት ዱካ።" : "Real-time production KPIs and the immutable audit trail."}
        actions={
          lastPolicy && (
            <Link
              href={`/staff/policies/${lastPolicy.policyId}`}
              className="card-interactive flex items-center gap-3 px-4 py-3"
            >
              <span className="grid h-9 w-9 place-items-center rounded-lg bg-success-50 text-success-600">
                <DocIcon />
              </span>
              <span className="text-left">
                <span className="block text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                  {am ? "የቅርብ ጊዜ ፖሊሲ" : "Latest policy"}
                </span>
                <span className="font-mono text-sm font-bold text-slate-900">{lastPolicy.policyNumber}</span>
              </span>
            </Link>
          )
        }
      />

      {err && (
        <div className="badge badge-danger w-full justify-start rounded-xl px-4 py-3 text-sm" role="alert">
          <WarnIcon /> {err}
        </div>
      )}

      {/* KPI grid */}
      <div className="grid grid-cols-2 gap-4 sm:gap-6 lg:grid-cols-5">
        {loading || !kpis ? (
          Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-32 w-full" />)
        ) : (
          <>
            <StatCard label={am ? "ፖሊሲዎች" : "Policies in force"} value={kpis.policy_count.toLocaleString()} icon={<ShieldIcon />} />
            <StatCard label={am ? "ጠቅላላ ፕሪሚየም" : "Gross written premium"} value={formatETB(kpis.premium_written_minor, locale)} icon={<ChartIcon />} tone="success" />
            <StatCard label={am ? "የተከፈለ ይገባኛል" : "Claims paid"} value={formatETB(kpis.claims_paid_minor, locale)} icon={<CashIcon />} tone="warning" />
            <StatCard label={am ? "የኪሳራ ሬሾ" : "Loss ratio"} value={pct(kpis.loss_ratio)} icon={<TrendIcon />} tone={lossTone} hint={`${kpis.claim_count} ${am ? "ይገባኛል" : "claims"}`} />
            <StatCard label={am ? "ጥምር ሬሾ" : "Combined ratio"} value={pct(kpis.combined_ratio)} icon={<LayersIcon />} tone={kpis.combined_ratio > 1 ? "danger" : "brand"} hint={`+${pct(kpis.assumed_expense_ratio)} ${am ? "ወጪ" : "expense"}`} />
          </>
        )}
      </div>

      {/* Audit trail */}
      <section className="card overflow-hidden">
        <div className="flex items-center justify-between gap-4 border-b border-slate-100 bg-slate-50/60 px-6 py-4">
          <div>
            <h2 className="text-lg font-bold text-slate-900">{am ? "የማይለወጥ ኦዲት ዱካ" : "Immutable audit trail"}</h2>
            <p className="mt-0.5 text-sm text-slate-500">{am ? "በሁሉም ሞጁሎች የተመዘገቡ የቅርብ ክስተቶች።" : "Most recent domain events across every module."}</p>
          </div>
          <span className="badge badge-neutral hidden sm:inline-flex">{audit.length} {am ? "ክስተቶች" : "events"}</span>
        </div>

        {loading ? (
          <div className="space-y-3 p-6">
            {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-10 w-full" />)}
          </div>
        ) : audit.length === 0 ? (
          <div className="flex flex-col items-center gap-3 p-16 text-center">
            <span className="grid h-12 w-12 place-items-center rounded-full bg-slate-100 text-slate-400"><DocIcon /></span>
            <p className="text-sm text-slate-500">{am ? "እስካሁን ምንም ክስተት አልተመዘገበም።" : "No audit events recorded yet."}</p>
            <Link href="/quote" className="btn btn-primary btn-sm mt-1">{am ? "ቅናሽ ጀምር" : "Create a quote"}</Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="data-table">
              <thead>
                <tr>
                  <th>{am ? "ክስተት" : "Event"}</th>
                  <th className="hidden md:table-cell">{am ? "ዝርዝር" : "Payload"}</th>
                  <th>ID</th>
                  <th className="text-right">{am ? "ጊዜ" : "When"}</th>
                </tr>
              </thead>
              <tbody>
                {audit.map((e) => (
                  <tr key={e.id}>
                    <td><StatusBadge value={e.topic} /></td>
                    <td className="hidden max-w-md truncate font-mono text-xs text-slate-500 md:table-cell">{JSON.stringify(e.payload)}</td>
                    <td className="font-mono text-xs text-slate-400">{e.id.slice(0, 10)}…</td>
                    <td className="whitespace-nowrap text-right text-xs font-medium text-slate-500">{relTime(e.recorded_at, am)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}

function relTime(iso: string, am: boolean): string {
  const d = new Date(iso).getTime();
  if (Number.isNaN(d)) return "";
  const s = Math.max(0, Math.round((Date.now() - d) / 1000));
  if (s < 60) return am ? "አሁን" : "just now";
  const m = Math.round(s / 60);
  if (m < 60) return `${m}${am ? " ደቂቃ" : "m ago"}`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h}${am ? " ሰዓት" : "h ago"}`;
  return new Date(iso).toLocaleDateString();
}

/* icons */
const ShieldIcon = () => <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>;
const ChartIcon = () => <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6m4 6V5m4 14v-9M3 21h18" /></svg>;
const CashIcon = () => <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" /></svg>;
const TrendIcon = () => <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>;
const LayersIcon = () => <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 2 2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" /></svg>;
const DocIcon = () => <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>;
const WarnIcon = () => <svg className="h-4 w-4 shrink-0" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
