"use client";

import { useLocale, useTranslations } from "next-intl";
import { BadgePercent } from "lucide-react";
import { useRecents } from "@/lib/recents";
import { formatBirr } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart } from "@/components/ui/charts";
import { EmptyState } from "@/components/ui/states";

const RATE = 0.125;

export default function CommissionsPage() {
  const t = useTranslations();
  const locale = useLocale() as Locale;
  const policies = useRecents("policy");

  const total = policies.reduce((s, p) => s + p.premium * RATE, 0);
  const chartData = policies
    .slice(0, 8)
    .reverse()
    .map((p) => ({ name: p.policyNumber.split("/").pop() ?? p.id.slice(0, 4), value: Math.round(p.premium * RATE) }));

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.broker")}</Eyebrow>} title={t("nav.commissions")} subtitle={`Earned at ${(RATE * 100).toFixed(1)}% of gross written premium.`} />
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label={t("nav.commissions")} value={formatBirr(total, locale)} icon={<BadgePercent />} tone="success" />
        <StatCard label={t("nav.policies")} value={String(policies.length)} tone="brand" />
        <StatCard label="Rate" value={`${(RATE * 100).toFixed(1)}%`} tone="info" />
      </div>
      <Card>
        <CardHeader><CardTitle>{t("nav.commissions")}</CardTitle></CardHeader>
        <CardContent>
          {chartData.length === 0 ? (
            <EmptyState title={t("common.empty")} action={{ label: t("nav.newBusiness"), href: "/broker/new-business" }} />
          ) : (
            <BarChart data={chartData} xKey="name" series={[{ key: "value", name: t("nav.commissions") }]} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
