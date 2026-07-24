"use client";

import { useLocale, useTranslations } from "next-intl";
import { BarChart3, Coins, Layers, ShieldCheck, TrendingUp } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useAudit, useKpis } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatETB, formatPct, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart, DonutChart } from "@/components/ui/charts";
import { StatusBadge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";
import { Button } from "@/components/ui/button";

export default function StaffDashboard() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const kpis = useKpis();
  const audit = useAudit(8);

  const lossTone = kpis.data && kpis.data.loss_ratio > 1 ? "danger" : kpis.data && kpis.data.loss_ratio > 0.7 ? "warning" : "success";

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow><ShieldCheck /> {t("staff.eyebrow")}</Eyebrow>} title={t("staff.title")} subtitle={t("staff.subtitle")} />

      {kpis.isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(kpis.error, tErr)} action={{ label: t("common.retry"), onClick: () => kpis.refetch() }} /></Card>
      ) : (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
          {kpis.isLoading || !kpis.data ? (
            Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-32 w-full" />)
          ) : (
            <>
              <StatCard label={t("staff.policiesInForce")} value={kpis.data.policy_count.toLocaleString()} icon={<ShieldCheck />} tone="brand" />
              <StatCard label={t("staff.gwp")} value={formatETB(kpis.data.premium_written_minor, locale, { compact: true })} icon={<BarChart3 />} tone="success" />
              <StatCard label={t("staff.claimsPaid")} value={formatETB(kpis.data.claims_paid_minor, locale, { compact: true })} icon={<Coins />} tone="warning" />
              <StatCard label={t("staff.lossRatio")} value={formatPct(kpis.data.loss_ratio, locale)} icon={<TrendingUp />} tone={lossTone} hint={`${kpis.data.claim_count} ${t("nav.claims")}`} />
              <StatCard label={t("staff.combinedRatio")} value={formatPct(kpis.data.combined_ratio, locale)} icon={<Layers />} tone={kpis.data.combined_ratio > 1 ? "danger" : "info"} hint={`+${formatPct(kpis.data.assumed_expense_ratio, locale)}`} />
            </>
          )}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader><CardTitle>{t("staff.gwp")} vs {t("staff.claimsPaid")}</CardTitle></CardHeader>
          <CardContent>
            {kpis.data && (
              <BarChart
                xKey="name"
                data={[
                  { name: t("staff.gwp"), value: kpis.data.premium_written_minor / 100 },
                  { name: t("staff.claimsPaid"), value: kpis.data.claims_paid_minor / 100 },
                ]}
                series={[{ key: "value", name: "ETB" }]}
              />
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>Portfolio</CardTitle></CardHeader>
          <CardContent>
            {kpis.data && (
              <DonutChart data={[{ name: t("nav.policies"), value: kpis.data.policy_count }, { name: t("nav.claims"), value: kpis.data.claim_count || 0.001 }]} />
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <CardTitle>{t("staff.auditTitle")}</CardTitle>
          <Button asChild variant="ghost" size="sm"><Link href="/staff/audit">{t("common.viewAll")}</Link></Button>
        </CardHeader>
        <CardContent className="pt-0">
          {audit.isLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-9 w-full" />)}</div>
          ) : (
            <ul className="divide-y divide-border-subtle">
              {(audit.data ?? []).map((e) => (
                <li key={e.id} className="flex items-center justify-between gap-3 py-2.5">
                  <StatusBadge value={e.topic} />
                  <span className="font-mono text-xs text-fg-subtle">{relativeTime(e.recorded_at, locale)}</span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
