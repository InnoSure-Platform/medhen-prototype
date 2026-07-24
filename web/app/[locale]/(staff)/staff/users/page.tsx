"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { useUsers } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import type { IamUser } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";

export default function UsersPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const { data, isLoading, isError, error, refetch } = useUsers();

  const columns: ColumnDef<IamUser>[] = React.useMemo(
    () => [
      { accessorKey: "full_name", header: t("quote.fullName"), cell: ({ row }) => <span className="font-medium text-fg">{row.original.full_name || row.original.subject}</span> },
      { accessorKey: "email", header: "Email", cell: ({ row }) => <span className="text-fg-muted">{row.original.email || "—"}</span> },
      {
        id: "roles",
        header: t("roles.admin"),
        cell: ({ row }) => (
          <div className="flex flex-wrap gap-1">
            {(row.original.roles ?? []).map((r) => (
              <Badge key={r} tone="brand">{r}</Badge>
            ))}
          </div>
        ),
      },
    ],
    [t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={t("nav.users")} subtitle="Application users and their roles (IAM)." />
      {isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>
      ) : (
        <DataTable columns={columns} data={data ?? []} loading={isLoading} filterColumn="full_name" filterPlaceholder={t("common.search")} emptyTitle={t("common.empty")} />
      )}
    </div>
  );
}
