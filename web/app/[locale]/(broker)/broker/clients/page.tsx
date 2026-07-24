"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { useParties } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import type { Party } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";

export default function ClientsPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const { data, isLoading, isError, error, refetch } = useParties();

  const columns: ColumnDef<Party>[] = React.useMemo(
    () => [
      { accessorKey: "FullName", header: t("quote.fullName"), cell: ({ row }) => <span className="font-medium text-fg">{row.original.FullName}</span> },
      { id: "am", header: t("quote.fullNameAm"), cell: ({ row }) => <span className="font-ethiopic text-fg-muted">{row.original.FullNameAmharic || "—"}</span> },
      { accessorKey: "PhoneE164", header: t("quote.phone"), cell: ({ row }) => <span className="font-mono text-fg-muted">{row.original.PhoneE164}</span> },
      { accessorKey: "Status", header: t("common.status"), cell: ({ row }) => <StatusBadge value={row.original.Status} /> },
    ],
    [t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.broker")}</Eyebrow>} title={t("nav.clients")} subtitle="Parties registered on this tenant." />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : (
        <DataTable columns={columns} data={data ?? []} loading={isLoading} filterColumn="FullName" filterPlaceholder={t("common.search")} pageSize={15} emptyTitle={t("common.empty")} />
      )}
    </div>
  );
}
