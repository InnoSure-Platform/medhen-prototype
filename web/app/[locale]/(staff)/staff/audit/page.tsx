"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useLocale, useTranslations } from "next-intl";
import { useAudit } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import type { Audit } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";

export default function AuditPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const { data, isLoading, isError, error, refetch } = useAudit(100);

  const columns: ColumnDef<Audit>[] = React.useMemo(
    () => [
      { accessorKey: "topic", header: t("staff.event"), cell: ({ row }) => <StatusBadge value={row.original.topic} /> },
      {
        id: "payload",
        header: t("staff.payload"),
        cell: ({ row }) => (
          <span className="block max-w-md truncate font-mono text-xs text-fg-muted">{JSON.stringify(row.original.payload)}</span>
        ),
      },
      { accessorKey: "id", header: "ID", cell: ({ row }) => <span className="font-mono text-xs text-fg-subtle">{row.original.id.slice(0, 10)}…</span> },
      {
        accessorKey: "recorded_at",
        header: t("staff.when"),
        cell: ({ row }) => <span className="whitespace-nowrap text-xs text-fg-muted">{relativeTime(row.original.recorded_at, locale)}</span>,
      },
    ],
    [locale, t],
  );

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={t("staff.auditTitle")} subtitle={t("staff.auditSub")} />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : (
        <DataTable columns={columns} data={data ?? []} loading={isLoading} filterColumn="topic" filterPlaceholder={t("common.search")} pageSize={15} emptyTitle={t("common.empty")} />
      )}
    </div>
  );
}
