"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "@/lib/i18n/navigation";
import { useRecents, type RecentPolicy } from "@/lib/recents";
import { formatBirr, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Link } from "@/lib/i18n/navigation";

export default function PoliciesPage() {
  const t = useTranslations();
  const locale = useLocale() as Locale;
  const router = useRouter();
  const policies = useRecents("policy");

  const columns: ColumnDef<RecentPolicy>[] = React.useMemo(
    () => [
      {
        accessorKey: "policyNumber",
        header: t("quote.policyNumber"),
        cell: ({ row }) => <span className="font-mono font-semibold text-fg">{row.original.policyNumber}</span>,
      },
      { accessorKey: "productCode", header: "Product" },
      {
        accessorKey: "premium",
        header: t("quote.premium"),
        cell: ({ row }) => <span className="font-mono">{formatBirr(row.original.premium, locale)}</span>,
      },
      {
        accessorKey: "status",
        header: t("common.status"),
        cell: ({ row }) => <StatusBadge value={row.original.status} />,
      },
      {
        accessorKey: "createdAt",
        header: t("staff.when"),
        cell: ({ row }) => <span className="text-fg-muted">{relativeTime(row.original.createdAt, locale)}</span>,
      },
    ],
    [locale, t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader
        eyebrow={<Eyebrow>{t("roles.customer")}</Eyebrow>}
        title={t("nav.policies")}
        actions={
          <Button asChild>
            <Link href="/customer/quote">{t("marketing.ctaQuote")}</Link>
          </Button>
        }
      />
      <DataTable
        columns={columns}
        data={policies}
        filterColumn="policyNumber"
        filterPlaceholder={t("common.search")}
        onRowClick={(p) => router.push(`/customer/policies/${p.id}`)}
        emptyTitle={t("common.empty")}
        emptyDescription={t("brand.tagline")}
      />
    </div>
  );
}
