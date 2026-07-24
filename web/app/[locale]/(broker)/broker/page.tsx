"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useLocale, useTranslations } from "next-intl";
import { BadgePercent, FileText, ShieldCheck } from "lucide-react";
import { Link, useRouter } from "@/lib/i18n/navigation";
import { useRecents, type RecentPolicy } from "@/lib/recents";
import { formatBirr, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

const COMMISSION_RATE = 0.125;

export default function BrokerBook() {
  const t = useTranslations();
  const locale = useLocale() as Locale;
  const router = useRouter();
  const policies = useRecents("policy");

  const gwp = policies.reduce((s, p) => s + (p.premium ?? 0), 0);
  const commission = gwp * COMMISSION_RATE;

  const columns: ColumnDef<RecentPolicy>[] = React.useMemo(
    () => [
      { accessorKey: "policyNumber", header: t("quote.policyNumber"), cell: ({ row }) => <span className="font-mono font-semibold text-fg">{row.original.policyNumber}</span> },
      { accessorKey: "premium", header: t("quote.premium"), cell: ({ row }) => <span className="font-mono">{formatBirr(row.original.premium, locale)}</span> },
      {
        id: "commission",
        header: t("nav.commissions"),
        cell: ({ row }) => <span className="font-mono text-success-fg">{formatBirr(row.original.premium * COMMISSION_RATE, locale)}</span>,
      },
      { accessorKey: "status", header: t("common.status"), cell: ({ row }) => <StatusBadge value={row.original.status} /> },
      { accessorKey: "createdAt", header: t("staff.when"), cell: ({ row }) => <span className="text-fg-muted">{relativeTime(row.original.createdAt, locale)}</span> },
    ],
    [locale, t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader
        eyebrow={<Eyebrow>{t("roles.broker")}</Eyebrow>}
        title={t("nav.book")}
        actions={<Button asChild><Link href="/broker/new-business"><FileText /> {t("nav.newBusiness")}</Link></Button>}
      />
      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label={t("nav.policies")} value={String(policies.length)} icon={<ShieldCheck />} tone="brand" />
        <StatCard label={t("staff.gwp")} value={formatBirr(gwp, locale, { compact: true })} icon={<FileText />} tone="info" />
        <StatCard label={t("nav.commissions")} value={formatBirr(commission, locale, { compact: true })} icon={<BadgePercent />} tone="success" hint={`${(COMMISSION_RATE * 100).toFixed(1)}%`} />
      </div>
      <DataTable columns={columns} data={policies} filterColumn="policyNumber" filterPlaceholder={t("common.search")} onRowClick={(p) => router.push(`/customer/policies/${p.id}`)} emptyTitle={t("common.empty")} emptyDescription={t("brand.tagline")} />
    </div>
  );
}
