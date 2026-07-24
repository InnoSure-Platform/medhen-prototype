"use client";

import * as React from "react";
import { useLocale, useTranslations } from "next-intl";
import { toast } from "sonner";
import { CreditCard, Download, FileText, RefreshCw, ShieldOff, SquarePen } from "lucide-react";
import { useCancelPolicy, useEndorsePolicy, usePolicy, useRenewPolicy } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { useAuth } from "@/components/providers";
import { formatBirr, formatDate } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { Link } from "@/lib/i18n/navigation";
import { Breadcrumb } from "@/components/ui/breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { Timeline } from "@/components/ui/timeline";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { MoneyInput } from "@/components/ui/money-input";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

export default function PolicyDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = React.use(params);
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const { can } = useAuth();
  const { data: policy, isLoading, isError, error, refetch } = usePolicy(id);

  const canService = can("policy:service");
  const inForce = policy?.Status === "ISSUED";

  return (
    <div className="mx-auto max-w-4xl space-y-6 px-6 py-8">
      <Breadcrumb
        LinkComponent={Link}
        items={[{ label: t("nav.policies"), href: "/customer/policies" }, { label: policy?.PolicyNumber ?? "…" }]}
      />

      {isLoading && (
        <div className="space-y-4">
          <Skeleton className="h-28 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      )}
      {isError && (
        <Card>
          <ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} />
        </Card>
      )}

      {policy && (
        <>
          <Card>
            <CardContent className="flex flex-col gap-4 p-6 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">{t("quote.policyNumber")}</p>
                <p className="mt-1 font-mono text-2xl font-bold text-fg">{policy.PolicyNumber}</p>
              </div>
              <StatusBadge value={policy.Status} className="text-sm" />
            </CardContent>
          </Card>

          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader><CardTitle>{t("quote.step.quotation")}</CardTitle></CardHeader>
              <CardContent className="space-y-3 pt-0 text-sm">
                <Row label={t("quote.premium")} value={formatBirr(policy.GrossPremium, locale)} mono />
                <Row label="Product" value={policy.ProductCode ?? "MOT"} />
                <Row label={t("quote.coverPeriod")} value={`${formatDate(policy.EffectiveFrom, locale)} — ${formatDate(policy.EffectiveTo, locale)}`} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader><CardTitle>Lifecycle</CardTitle></CardHeader>
              <CardContent className="pt-0">
                <Timeline
                  items={[
                    { title: t("marketing.step.quote"), state: "done" },
                    { title: t("marketing.step.underwrite"), description: "STP", state: "done" },
                    { title: t("marketing.step.bind"), description: policy.PolicyNumber, state: "done" },
                    { title: t("marketing.step.pay"), description: t("quote.issuedSub"), state: policy.Status === "ISSUED" ? "current" : "done" },
                  ]}
                />
              </CardContent>
            </Card>
          </div>

          {/* Servicing — staff/broker/admin only */}
          {canService && (
            <Card>
              <CardHeader><CardTitle>{t("servicing.title")}</CardTitle></CardHeader>
              <CardContent className="flex flex-wrap gap-3 pt-0">
                <EndorseDialog policyId={policy.ID} disabled={!inForce} />
                <RenewButton policyId={policy.ID} disabled={!inForce} />
                <CancelDialog policyId={policy.ID} disabled={!inForce} locale={locale} />
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader><CardTitle>{t("nav.documents")}</CardTitle></CardHeader>
            <CardContent className="flex flex-wrap gap-3 pt-0">
              <Button variant="secondary"><FileText /> Certificate of Insurance <Download className="opacity-60" /></Button>
              <Button asChild variant="secondary"><Link href="/customer/invoices"><CreditCard /> {t("nav.invoices")}</Link></Button>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}

function EndorseDialog({ policyId, disabled }: { policyId: string; disabled: boolean }) {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const endorse = useEndorsePolicy();
  const [open, setOpen] = React.useState(false);
  const [delta, setDelta] = React.useState<number | "">("");
  const [reason, setReason] = React.useState("");

  async function submit() {
    if (delta === "") return;
    try {
      await endorse.mutateAsync({ policyId, deltaMinor: Math.round((delta as number) * 100), reason });
      toast.success(t("servicing.endorsed"));
      setOpen(false);
      setDelta("");
      setReason("");
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="secondary" disabled={disabled}><SquarePen /> {t("servicing.endorse")}</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader><DialogTitle>{t("servicing.endorse")}</DialogTitle></DialogHeader>
        <div className="space-y-4">
          <Field label={t("servicing.premiumDelta")} htmlFor="delta">
            <MoneyInput id="delta" value={delta} onChange={setDelta} />
          </Field>
          <Field label={t("servicing.reason")} htmlFor="reason">
            <Input id="reason" value={reason} onChange={(e) => setReason(e.target.value)} />
          </Field>
        </div>
        <DialogFooter>
          <Button variant="secondary" onClick={() => setOpen(false)}>{t("common.cancel")}</Button>
          <Button onClick={submit} loading={endorse.isPending} disabled={delta === ""}>{t("servicing.endorse")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RenewButton({ policyId, disabled }: { policyId: string; disabled: boolean }) {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const renew = useRenewPolicy();
  return (
    <Button
      variant="secondary"
      disabled={disabled}
      loading={renew.isPending}
      onClick={async () => {
        try {
          const p = await renew.mutateAsync(policyId);
          toast.success(t("servicing.renewed"), { description: p.PolicyNumber });
        } catch (e) {
          toast.error(errorMessage(e, tErr));
        }
      }}
    >
      <RefreshCw /> {t("servicing.renew")}
    </Button>
  );
}

function CancelDialog({ policyId, disabled, locale }: { policyId: string; disabled: boolean; locale: Locale }) {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const cancel = useCancelPolicy();
  const [open, setOpen] = React.useState(false);
  const [reason, setReason] = React.useState("");

  async function submit() {
    try {
      const res = await cancel.mutateAsync({ policyId, reason });
      toast.success(t("servicing.cancelled"), {
        description: `${t("servicing.refund")}: ${formatBirr((res.refund_minor ?? 0) / 100, locale)}`,
      });
      setOpen(false);
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="danger" disabled={disabled}><ShieldOff /> {t("servicing.cancel")}</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("servicing.cancel")}</DialogTitle>
          <DialogDescription>{t("servicing.cancelWarning")}</DialogDescription>
        </DialogHeader>
        <Field label={t("servicing.reason")} htmlFor="cancel-reason">
          <Input id="cancel-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
        </Field>
        <DialogFooter>
          <Button variant="secondary" onClick={() => setOpen(false)}>{t("common.cancel")}</Button>
          <Button variant="danger" onClick={submit} loading={cancel.isPending}>{t("servicing.confirmCancel")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
