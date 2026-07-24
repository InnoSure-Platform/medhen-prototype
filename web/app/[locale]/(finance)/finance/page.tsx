"use client";

import { useLocale, useTranslations } from "next-intl";
import { Coins, Scale, TrendingUp, Wallet } from "lucide-react";
import { useKpis } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatETB, formatPct } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart } from "@/components/ui/charts";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";

export default function FinanceDashboard() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const kpis = useKpis();

  const gwp = (kpis.data?.premium_written_minor ?? 0) / 100;
  const paid = (kpis.data?.claims_paid_minor ?? 0) / 100;

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.finance")}</Eyebrow>} title={t("finance.title")} subtitle={t("finance.subtitle")} />

      {kpis.isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(kpis.error, tErr)} action={{ label: t("common.retry"), onClick: () => kpis.refetch() }} /></Card>
      ) : kpis.isLoading || !kpis.data ? (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-32 w-full" />)}</div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <StatCard label={t("finance.gwp")} value={formatETB(kpis.data.premium_written_minor, locale, { compact: true })} icon={<Wallet />} tone="success" />
            <StatCard label={t("finance.claimsPaid")} value={formatETB(kpis.data.claims_paid_minor, locale, { compact: true })} icon={<Coins />} tone="warning" />
            <StatCard label={t("finance.netPosition")} value={formatETB(kpis.data.premium_written_minor - kpis.data.claims_paid_minor, locale, { compact: true })} icon={<TrendingUp />} tone="brand" />
            <StatCard label={t("staff.combinedRatio")} value={formatPct(kpis.data.combined_ratio, locale)} icon={<Scale />} tone={kpis.data.combined_ratio > 1 ? "danger" : "info"} />
          </div>

          <Card>
            <CardHeader><CardTitle>{t("finance.premiumVsClaims")}</CardTitle></CardHeader>
            <CardContent>
              <BarChart
                xKey="name"
                data={[
                  { name: t("finance.gwp"), value: Math.round(gwp) },
                  { name: t("finance.claimsPaid"), value: Math.round(paid) },
                ]}
                series={[{ key: "value", name: "ETB" }]}
              />
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
