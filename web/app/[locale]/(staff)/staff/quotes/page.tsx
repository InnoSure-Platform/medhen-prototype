"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useLocale, useTranslations } from "next-intl";
import { useQuotes } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import type { Quote } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";

export default function StaffQuotesPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const { data, isLoading, isError, error, refetch } = useQuotes();

  const columns: ColumnDef<Quote>[] = React.useMemo(
    () => [
      { accessorKey: "ID", header: "ID", cell: ({ row }) => <span className="font-mono text-xs text-fg-muted">{row.original.ID.slice(0, 12)}…</span> },
      { accessorKey: "ProductCode", header: "Product" },
      { accessorKey: "GrossPremium", header: t("quote.gross"), cell: ({ row }) => <span className="font-mono">{formatBirr(row.original.GrossPremium, locale)}</span> },
      { accessorKey: "Status", header: t("common.status"), cell: ({ row }) => <StatusBadge value={row.original.Status} /> },
    ],
    [locale, t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={t("nav.quotes")} subtitle="All quotes for this tenant." />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : (
        <DataTable columns={columns} data={data ?? []} loading={isLoading} filterColumn="ProductCode" filterPlaceholder={t("common.search")} pageSize={15} emptyTitle={t("common.empty")} />
      )}
    </div>
  );
}
