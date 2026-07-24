"use client";

import * as React from "react";
import { Suspense } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "@/lib/i18n/navigation";
import { toast } from "sonner";
import { Search, ShieldCheck } from "lucide-react";
import { useAuth } from "@/components/providers";
import { useClaim, useClaims, useSettleClaim } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import type { Claim } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { MoneyInput } from "@/components/ui/money-input";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { Alert } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";

export default function StaffClaimsPage() {
  return (
    <Suspense>
      <StaffClaimsInner />
    </Suspense>
  );
}

function StaffClaimsInner() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const params = useSearchParams();
  const router = useRouter();

  const { can, stepUp } = useAuth();
  const canSettle = can("claim:settle");

  // Claim id + step-up flag live in the URL so they survive the step-up redirect.
  const activeId = params.get("claim") ?? "";
  const steppedUp = params.get("stepup") === "1";

  const [idInput, setIdInput] = React.useState(activeId);
  const [amount, setAmount] = React.useState<number | "">("");

  const claim = useClaim(activeId || undefined);
  const settle = useSettleClaim();
  const claimsList = useClaims();

  function lookup(id: string) {
    router.push(`/staff/claims?claim=${encodeURIComponent(id)}`);
  }

  const columns: ColumnDef<Claim>[] = React.useMemo(
    () => [
      { accessorKey: "Description", header: t("claim.description"), cell: ({ row }) => <span className="text-fg">{row.original.Description || row.original.ID.slice(0, 10)}</span> },
      { accessorKey: "Reserve", header: "Reserve", cell: ({ row }) => <span className="font-mono">{formatBirr(row.original.Reserve, locale)}</span> },
      { accessorKey: "Status", header: t("common.status"), cell: ({ row }) => <StatusBadge value={row.original.Status} /> },
      { accessorKey: "CreatedAt", header: t("staff.when"), cell: ({ row }) => <span className="text-fg-muted">{relativeTime(row.original.CreatedAt, locale)}</span> },
    ],
    [locale, t],
  );

  async function onSettle() {
    if (!activeId || amount === "") return;
    try {
      const updated = await settle.mutateAsync({ claimId: activeId, amountMinor: Math.round((amount as number) * 100) });
      toast.success(t("claim.settled"), { description: formatBirr(updated.SettledAmount, locale) });
    } catch (e) {
      // 409 = above authority → referred for manual review.
      toast.error(errorMessage(e, tErr));
    }
  }

  return (
    <div className="mx-auto max-w-3xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={`${t("nav.claims")} · ${t("claim.settle")}`} subtitle={t("claim.subtitle")} />

      <DataTable
        columns={columns}
        data={claimsList.data ?? []}
        loading={claimsList.isLoading}
        filterColumn="Description"
        filterPlaceholder={t("common.search")}
        pageSize={10}
        onRowClick={(c) => lookup(c.ID)}
        emptyTitle={t("common.empty")}
      />

      <Card>
        <CardHeader><CardTitle>{t("common.search")}</CardTitle></CardHeader>
        <CardContent>
          <form className="flex items-end gap-3" onSubmit={(e) => { e.preventDefault(); lookup(idInput.trim()); }}>
            <Field label={t("claim.policyId")} htmlFor="claimId" className="flex-1">
              <Input id="claimId" className="font-mono" placeholder="claim UUID" value={idInput} onChange={(e) => setIdInput(e.target.value)} />
            </Field>
            <Button type="submit"><Search /> {t("common.search")}</Button>
          </form>
        </CardContent>
      </Card>

      {activeId && claim.isLoading && <Skeleton className="h-48 w-full" />}
      {activeId && claim.isError && <Alert tone="danger" title={t("errors.notFoundTitle")}>{errorMessage(claim.error, tErr)}</Alert>}

      {claim.data && (
        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle>{claim.data.Description || claim.data.ID.slice(0, 10)}</CardTitle>
            <StatusBadge value={claim.data.Status} />
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <Row label="Reserve" value={formatBirr(claim.data.Reserve, locale)} />
              <Row label={t("claim.settled")} value={claim.data.Status === "SETTLED" ? formatBirr(claim.data.SettledAmount, locale) : "—"} />
            </div>

            {claim.data.Status === "SETTLED" ? (
              <Alert tone="success" title={t("claim.settled")}>{formatBirr(claim.data.SettledAmount, locale)}</Alert>
            ) : !canSettle ? (
              <Alert tone="warning" title={t("errors.forbiddenTitle")}>{t("errors.forbiddenBody")}</Alert>
            ) : !steppedUp ? (
              // Sensitive action → require a fresh authentication (step-up).
              <Alert tone="info" title={t("auth.stepUpTitle")}>
                <p className="mb-3">{t("auth.stepUpBody")}</p>
                <Button size="sm" onClick={() => stepUp(`/${locale}/staff/claims?claim=${encodeURIComponent(activeId)}&stepup=1`)}>
                  <ShieldCheck /> {t("auth.stepUp")}
                </Button>
              </Alert>
            ) : (
              <>
                <Field label={t("claim.settle")} htmlFor="settleAmount" hint="Above your authority limit routes to manual review (409).">
                  <MoneyInput id="settleAmount" value={amount} onChange={setAmount} min={1} />
                </Field>
                <Button onClick={onSettle} loading={settle.isPending} disabled={amount === ""} className="w-full">
                  {t("claim.settle")}
                </Button>
              </>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-subtle/50 p-4">
      <div className="text-xs font-bold uppercase tracking-wider text-fg-subtle">{label}</div>
      <div className="mt-1 font-mono text-lg font-bold text-fg">{value}</div>
    </div>
  );
}
