"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "@/lib/i18n/navigation";
import { usePolicies } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr, formatDate } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import type { Policy } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";

export default function StaffPoliciesPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const router = useRouter();
  const { data, isLoading, isError, error, refetch } = usePolicies();

  const columns: ColumnDef<Policy>[] = React.useMemo(
    () => [
      { accessorKey: "PolicyNumber", header: t("quote.policyNumber"), cell: ({ row }) => <span className="font-mono font-semibold text-fg">{row.original.PolicyNumber}</span> },
      { accessorKey: "ProductCode", header: "Product" },
      { accessorKey: "GrossPremium", header: t("quote.premium"), cell: ({ row }) => <span className="font-mono">{formatBirr(row.original.GrossPremium, locale)}</span> },
      { accessorKey: "Status", header: t("common.status"), cell: ({ row }) => <StatusBadge value={row.original.Status} /> },
      { accessorKey: "EffectiveFrom", header: t("staff.when"), cell: ({ row }) => <span className="text-fg-muted">{formatDate(row.original.EffectiveFrom, locale)}</span> },
    ],
    [locale, t],
  );

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={t("nav.policies")} subtitle="All issued policies for this tenant." />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : (
        <DataTable
          columns={columns}
          data={data ?? []}
          loading={isLoading}
          filterColumn="PolicyNumber"
          filterPlaceholder={t("common.search")}
          pageSize={15}
          onRowClick={(p) => router.push(`/customer/policies/${p.ID}`)}
          emptyTitle={t("common.empty")}
        />
      )}
    </div>
  );
}
