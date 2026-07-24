"use client";

import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useLocale, useTranslations } from "next-intl";
import { toast } from "sonner";
import { MapPin, Send } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useSubmitFNOL } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { addRecentClaim, useRecents } from "@/lib/recents";
import { relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { MoneyInput } from "@/components/ui/money-input";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/states";

const schema = z.object({
  policyId: z.string().min(6, "Enter the policy ID"),
  description: z.string().min(4, "Describe what happened"),
  latitude: z.coerce.number().min(-90).max(90),
  longitude: z.coerce.number().min(-180).max(180),
  reserve: z.number().min(1, "Enter an estimated loss"),
});
type FormValues = z.infer<typeof schema>;

export default function ClaimsPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;
  const claims = useRecents("claim");
  const submitFNOL = useSubmitFNOL();

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { policyId: "", description: "", latitude: 9.0192, longitude: 38.7525, reserve: 40000 },
  });

  const onSubmit = handleSubmit(async (v) => {
    try {
      const claim = await submitFNOL.mutateAsync({
        policy_id: v.policyId,
        description: v.description,
        latitude: v.latitude,
        longitude: v.longitude,
        reserve_minor: Math.round(v.reserve * 100),
      });
      addRecentClaim({
        id: claim.ID,
        policyId: claim.PolicyID,
        status: claim.Status,
        description: claim.Description,
        reserve: claim.Reserve,
        createdAt: new Date().toISOString(),
      });
      toast.success(t("claim.filedTitle"), { description: claim.ID.slice(0, 12) });
      reset({ ...v, description: "" });
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  });

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("claim.eyebrow")}</Eyebrow>} title={t("claim.title")} subtitle={t("claim.subtitle")} />

      <div className="grid gap-6 lg:grid-cols-5">
        <Card className="lg:col-span-3">
          <CardHeader><CardTitle>{t("claim.title")}</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} className="space-y-5">
              <Field label={t("claim.policyId")} htmlFor="policyId" required error={errors.policyId?.message}>
                <Input id="policyId" className="font-mono" placeholder="policy UUID" aria-invalid={!!errors.policyId} {...register("policyId")} />
              </Field>
              <Field label={t("claim.description")} htmlFor="description" required error={errors.description?.message}>
                <Textarea id="description" aria-invalid={!!errors.description} {...register("description")} />
              </Field>
              <div className="grid grid-cols-2 gap-4">
                <Field label={t("claim.latitude")} htmlFor="latitude" error={errors.latitude?.message}>
                  <Input id="latitude" type="number" step="0.0001" className="font-mono" {...register("latitude")} />
                </Field>
                <Field label={t("claim.longitude")} htmlFor="longitude" error={errors.longitude?.message}>
                  <Input id="longitude" type="number" step="0.0001" className="font-mono" {...register("longitude")} />
                </Field>
              </div>
              <Field label={t("claim.reserve")} htmlFor="reserve" error={errors.reserve?.message}>
                <Controller
                  control={control}
                  name="reserve"
                  render={({ field }) => (
                    <MoneyInput id="reserve" value={field.value} onChange={(v) => field.onChange(v === "" ? 0 : v)} min={1} aria-invalid={!!errors.reserve} />
                  )}
                />
              </Field>
              <div className="flex items-center gap-2 text-xs text-fg-muted">
                <MapPin className="size-3.5" /> Addis Ababa (default GPS) — adjust as needed.
              </div>
              <Button type="submit" loading={submitFNOL.isPending} className="w-full">
                <Send /> {t("claim.submit")}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader><CardTitle>{t("nav.claims")}</CardTitle></CardHeader>
          <CardContent className="pt-0">
            {claims.length === 0 ? (
              <EmptyState title={t("common.empty")} />
            ) : (
              <ul className="divide-y divide-border-subtle">
                {claims.map((c) => (
                  <li key={c.id}>
                    <Link href={`/customer/claims/${c.id}`} className="flex items-center justify-between gap-3 py-3 transition-colors hover:opacity-80">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-fg">{c.description || c.id.slice(0, 10)}</p>
                        <p className="text-xs text-fg-muted">{relativeTime(c.createdAt, locale)}</p>
                      </div>
                      <StatusBadge value={c.status} />
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
