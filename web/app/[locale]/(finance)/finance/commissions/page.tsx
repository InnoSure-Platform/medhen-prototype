"use client";

import { useLocale, useTranslations } from "next-intl";
import { BadgePercent } from "lucide-react";
import { usePolicies } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart } from "@/components/ui/charts";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";

const RATE = 0.125;

export default function FinanceCommissionsPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const { data, isLoading, isError, error, refetch } = usePolicies();

  const policies = data ?? [];
  const gwp = policies.reduce((s, p) => s + (p.GrossPremium ?? 0), 0);
  const commission = gwp * RATE;
  const chart = policies.slice(0, 10).reverse().map((p) => ({
    name: p.PolicyNumber.split("/").pop() ?? p.ID.slice(0, 4),
    value: Math.round((p.GrossPremium ?? 0) * RATE),
  }));

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.finance")}</Eyebrow>} title={t("finance.commissionsTitle")} subtitle={`${(RATE * 100).toFixed(1)}% of gross written premium.`} />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-3">
            <StatCard label={t("nav.commissions")} value={formatBirr(commission, locale, { compact: true })} icon={<BadgePercent />} tone="success" />
            <StatCard label={t("finance.gwp")} value={formatBirr(gwp, locale, { compact: true })} tone="brand" />
            <StatCard label={t("nav.policies")} value={String(policies.length)} tone="info" />
          </div>
          <Card>
            <CardHeader><CardTitle>{t("finance.commissionsTitle")}</CardTitle></CardHeader>
            <CardContent>
              <BarChart data={chart} xKey="name" series={[{ key: "value", name: t("nav.commissions") }]} />
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
