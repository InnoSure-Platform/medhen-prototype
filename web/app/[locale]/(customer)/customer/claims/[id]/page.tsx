"use client";

import * as React from "react";
import { useLocale, useTranslations } from "next-intl";
import { useClaim } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { Link } from "@/lib/i18n/navigation";
import { Breadcrumb } from "@/components/ui/breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { Timeline } from "@/components/ui/timeline";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";

export default function ClaimDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = React.use(params);
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const { data: claim, isLoading, isError, error, refetch } = useClaim(id);

  const settled = claim?.Status === "SETTLED";

  return (
    <div className="mx-auto max-w-3xl space-y-6 px-6 py-8">
      <Breadcrumb LinkComponent={Link} items={[{ label: t("nav.claims"), href: "/customer/claims" }, { label: id.slice(0, 10) + "…" }]} />

      {isLoading && <Skeleton className="h-72 w-full" />}
      {isError && (
        <Card>
          <ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} />
        </Card>
      )}

      {claim && (
        <>
          <Card>
            <CardContent className="flex items-center justify-between gap-4 p-6">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("nav.claims")}</p>
                <p className="mt-1 text-lg font-medium text-fg">{claim.Description || "—"}</p>
              </div>
              <StatusBadge value={claim.Status} className="text-sm" />
            </CardContent>
          </Card>

          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader><CardTitle>Details</CardTitle></CardHeader>
              <CardContent className="space-y-3 pt-0 text-sm">
                <Row label="Reserve" value={formatBirr(claim.Reserve, locale)} mono />
                <Row label={t("claim.settled")} value={settled ? formatBirr(claim.SettledAmount, locale) : "—"} mono />
                <Row label="Location" value={`${claim.Latitude.toFixed(4)}, ${claim.Longitude.toFixed(4)}`} mono />
              </CardContent>
            </Card>
            <Card>
              <CardHeader><CardTitle>Progress</CardTitle></CardHeader>
              <CardContent className="pt-0">
                <Timeline
                  items={[
                    { title: t("claim.filedTitle"), description: t("claim.filedSub"), state: "done" },
                    { title: "Assessment", state: settled ? "done" : "current" },
                    { title: t("claim.settled"), description: settled ? formatBirr(claim.SettledAmount, locale) : undefined, state: settled ? "done" : "upcoming" },
                  ]}
                />
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-fg-muted">{label}</span>
      <span className={mono ? "font-mono font-medium text-fg" : "font-medium text-fg"}>{value}</span>
    </div>
  );
}
