"use client";

import { useLocale, useTranslations } from "next-intl";
import { toast } from "sonner";
import { CreditCard } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useRecents, type RecentPolicy } from "@/lib/recents";
import { useInitiatePayment, useInvoiceByPolicy } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { formatBirr, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/states";
import { Alert } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";

export default function InvoicesPage() {
  const t = useTranslations();
  const policies = useRecents("policy");

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.customer")}</Eyebrow>} title={t("nav.invoices")} subtitle={t("quote.issuedSub")} />

      <Alert tone="info" title="Telebirr">{t("marketing.feature.telebirrBody")}</Alert>

      {policies.length === 0 ? (
        <Card><EmptyState title={t("common.empty")} action={{ label: t("marketing.ctaQuote"), href: "/customer/quote" }} /></Card>
      ) : (
        <div className="space-y-3">
          {policies.map((p) => (
            <InvoiceRow key={p.id} policy={p} />
          ))}
        </div>
      )}
    </div>
  );
}

function InvoiceRow({ policy }: { policy: RecentPolicy }) {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const invoice = useInvoiceByPolicy(policy.id);
  const payInit = useInitiatePayment();

  const paid = invoice.data?.Status === "PAID";
  const amount = invoice.data ? invoice.data.AmountDue - invoice.data.AmountPaid : policy.premium;

  async function pay() {
    if (!invoice.data) return;
    try {
      const intent = await payInit.mutateAsync(invoice.data.ID);
      toast.success(t("servicing.redirecting"));
      window.open(intent.checkout_url, "_blank", "noopener,noreferrer");
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  return (
    <Card>
      <CardContent className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-xl border border-brand-border bg-brand-subtle text-brand-fg">
            <CreditCard className="size-5" />
          </span>
          <div>
            <Link href={`/customer/policies/${policy.id}`} className="font-mono text-sm font-bold text-fg hover:text-brand-fg">
              {policy.policyNumber}
            </Link>
            <p className="text-xs text-fg-muted">{relativeTime(policy.createdAt, locale)}</p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          {invoice.isLoading ? (
            <Skeleton className="h-6 w-24" />
          ) : (
            <>
              <span className="font-mono text-lg font-bold text-fg">{formatBirr(amount, locale)}</span>
              <StatusBadge value={invoice.data?.Status ?? "OPEN"} />
              <Button size="sm" onClick={pay} loading={payInit.isPending} disabled={paid || !invoice.data}>
                {t("servicing.payNow")}
              </Button>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
