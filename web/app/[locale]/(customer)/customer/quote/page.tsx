"use client";

import * as React from "react";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { ArrowRight, CheckCircle2, FileText } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useBindQuote, useCreateQuote, useRegisterParty } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { addRecentPolicy } from "@/lib/recents";
import { formatBirr } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { useLocale } from "next-intl";
import type { Quote } from "@/lib/api/types";

import { Stepper } from "@/components/ui/stepper";
import { Card } from "@/components/ui/card";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { MoneyInput } from "@/components/ui/money-input";
import { PhoneInput } from "@/components/ui/phone-input";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { StatusBadge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const schema = z.object({
  fullName: z.string().min(2, "Enter the full name"),
  fullNameAm: z.string().min(1, "Enter the Amharic name"),
  phone: z.string().regex(/^\+2519\d{8}$/, "Enter a valid +2519XXXXXXXX number"),
  nationalId: z.string().min(4, "Enter the Fayda / national ID"),
  plate: z.string().min(2, "Enter the plate number"),
  make: z.string().min(1, "Required"),
  model: z.string().min(1, "Required"),
  year: z.coerce.number().int().min(1980).max(2026),
  ageBand: z.enum(["young", "adult", "senior"]),
  cover: z.enum(["comprehensive", "tpl"]),
  sumInsured: z.number().min(1000, "At least 1,000 ETB"),
});
type FormValues = z.infer<typeof schema>;

type Step = 0 | 1 | 2 | 3;

export default function QuoteWizard() {
  const t = useTranslations("quote");
  const tc = useTranslations("common");
  const tNav = useTranslations("nav");
  const tErr = useTranslations("errors");
  const locale = useLocale() as Locale;

  const [step, setStep] = React.useState<Step>(0);
  const [partyId, setPartyId] = React.useState("");
  const [quote, setQuote] = React.useState<Quote | null>(null);
  const [policyNumber, setPolicyNumber] = React.useState("");
  const [policyMeta, setPolicyMeta] = React.useState<{ from: string; to: string } | null>(null);

  const registerParty = useRegisterParty();
  const createQuote = useCreateQuote();
  const bindQuote = useBindQuote();
  const busy = registerParty.isPending || createQuote.isPending || bindQuote.isPending;

  const {
    register,
    control,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    mode: "onTouched",
    defaultValues: {
      fullName: "Abebe Kebede",
      fullNameAm: "አበበ ከበደ",
      phone: "+251911234567",
      nationalId: "1234567890",
      plate: "AA-3-12345",
      make: "Toyota",
      model: "Corolla",
      year: 2021,
      ageBand: "adult",
      cover: "comprehensive",
      sumInsured: 1500000,
    },
  });

  const steps = [
    { id: "identity", label: t("step.identity") },
    { id: "vehicle", label: t("step.asset") },
    { id: "quote", label: t("step.quotation") },
    { id: "done", label: t("step.complete") },
  ];

  async function submitIdentity() {
    if (!(await trigger(["fullName", "fullNameAm", "phone", "nationalId"]))) return;
    const v = getValues();
    try {
      const party = await registerParty.mutateAsync({
        full_name: v.fullName,
        full_name_amharic: v.fullNameAm,
        phone_e164: v.phone,
        national_id: v.nationalId,
        address: { region: "Addis Ababa", zone: "Kirkos", woreda: "08", kebele: "15", city: "Addis Ababa", line1: "Bole Road" },
      });
      setPartyId(party.id);
      setStep(1);
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  async function submitVehicle() {
    if (!(await trigger(["plate", "make", "model", "year", "ageBand", "cover", "sumInsured"]))) return;
    const v = getValues();
    try {
      const q = await createQuote.mutateAsync({
        party_id: partyId,
        product_code: "MOT",
        coverages: v.cover === "comprehensive" ? ["OD", "TPL"] : ["TPL"],
        risk_dimensions: {
          age_band: v.ageBand,
          plate_number: v.plate,
          make: v.make,
          model: v.model,
          year: String(v.year),
          usage: "private",
          sum_insured: String(v.sumInsured),
        },
      });
      setQuote(q);
      setStep(2);
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  async function bind() {
    if (!quote) return;
    try {
      const policy = await bindQuote.mutateAsync(quote.ID);
      setPolicyNumber(policy.PolicyNumber);
      setPolicyMeta({ from: policy.EffectiveFrom, to: policy.EffectiveTo });
      addRecentPolicy({
        id: policy.ID,
        policyNumber: policy.PolicyNumber,
        premium: policy.GrossPremium,
        status: policy.Status,
        productCode: policy.ProductCode ?? "MOT",
        createdAt: new Date().toISOString(),
      });
      toast.success(t("issuedTitle"), { description: policy.PolicyNumber });
      setStep(3);
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  return (
    <div className="mx-auto max-w-3xl px-6 py-10">
      <span className="text-xs font-bold uppercase tracking-widest text-brand-fg">{t("eyebrow")}</span>
      <h1 className="mb-10 mt-2 text-3xl">{t("title")}</h1>

      <div className="mb-14">
        <Stepper steps={steps} current={step} />
      </div>

      <Card className="relative overflow-hidden p-6 sm:p-8">
        {busy && (
          <div className="absolute inset-0 z-20 grid place-items-center bg-surface/70 backdrop-blur-sm">
            <div className="flex flex-col items-center gap-3">
              <span className="size-9 animate-spin rounded-full border-4 border-brand border-t-transparent" />
              <p className="text-sm font-semibold text-brand-fg">{tc("processing")}</p>
            </div>
          </div>
        )}

        {step === 0 && (
          <div className="animate-rise space-y-6">
            <div>
              <h2 className="text-xl font-bold text-fg">{t("identityTitle")}</h2>
              <p className="mt-1 text-sm text-fg-muted">{t("identitySub")}</p>
            </div>
            <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
              <Field label={t("fullName")} htmlFor="fullName" required error={errors.fullName?.message}>
                <Input id="fullName" aria-invalid={!!errors.fullName} {...register("fullName")} />
              </Field>
              <Field label={t("fullNameAm")} htmlFor="fullNameAm" required error={errors.fullNameAm?.message}>
                <Input id="fullNameAm" className="font-ethiopic" aria-invalid={!!errors.fullNameAm} {...register("fullNameAm")} />
              </Field>
              <Field label={t("phone")} htmlFor="phone" required error={errors.phone?.message}>
                <Controller
                  control={control}
                  name="phone"
                  render={({ field }) => (
                    <PhoneInput id="phone" value={field.value} onChange={field.onChange} aria-invalid={!!errors.phone} />
                  )}
                />
              </Field>
              <Field label={t("nationalId")} htmlFor="nationalId" required error={errors.nationalId?.message}>
                <Input id="nationalId" className="font-mono tracking-widest" maxLength={12} aria-invalid={!!errors.nationalId} {...register("nationalId")} />
              </Field>
            </div>
            <div className="flex justify-end border-t border-border-subtle pt-5">
              <Button onClick={submitIdentity}>
                {tc("continue")} <ArrowRight />
              </Button>
            </div>
          </div>
        )}

        {step === 1 && (
          <div className="animate-rise space-y-6">
            <div>
              <h2 className="text-xl font-bold text-fg">{t("assetTitle")}</h2>
              <p className="mt-1 text-sm text-fg-muted">{t("assetSub")}</p>
            </div>
            <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
              <Field label={t("plate")} htmlFor="plate" required error={errors.plate?.message}>
                <Input id="plate" className="font-mono uppercase" aria-invalid={!!errors.plate} {...register("plate")} />
              </Field>
              <Field label={t("make")} htmlFor="make" required error={errors.make?.message}>
                <Input id="make" {...register("make")} />
              </Field>
              <Field label={t("model")} htmlFor="model" required error={errors.model?.message}>
                <Input id="model" {...register("model")} />
              </Field>
              <Field label={t("year")} htmlFor="year" required error={errors.year?.message}>
                <Input id="year" type="number" {...register("year")} />
              </Field>
              <Field label={t("ageBand")} htmlFor="ageBand">
                <Controller
                  control={control}
                  name="ageBand"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id="ageBand"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="young">{t("ageYoung")}</SelectItem>
                        <SelectItem value="adult">{t("ageAdult")}</SelectItem>
                        <SelectItem value="senior">{t("ageSenior")}</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
              </Field>
              <Field label={t("cover")} htmlFor="cover">
                <Controller
                  control={control}
                  name="cover"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id="cover"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="comprehensive">{t("coverComprehensive")}</SelectItem>
                        <SelectItem value="tpl">{t("coverThirdParty")}</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                />
              </Field>
              <div className="md:col-span-2">
                <Field label={t("sumInsured")} htmlFor="sumInsured" error={errors.sumInsured?.message}>
                  <Controller
                    control={control}
                    name="sumInsured"
                    render={({ field }) => (
                      <MoneyInput id="sumInsured" value={field.value} onChange={(v) => field.onChange(v === "" ? 0 : v)} min={1000} aria-invalid={!!errors.sumInsured} />
                    )}
                  />
                </Field>
              </div>
            </div>
            <div className="flex justify-between border-t border-border-subtle pt-5">
              <Button variant="ghost" onClick={() => setStep(0)}>{tc("back")}</Button>
              <Button onClick={submitVehicle}>
                {t("calculate")} <ArrowRight />
              </Button>
            </div>
          </div>
        )}

        {step === 2 && quote && (
          <div className="animate-rise space-y-6">
            <Alert tone="success" title={t("stp")}>
              <span className="inline-flex items-center gap-2">Status: <StatusBadge value={quote.Status} /></span>
            </Alert>
            <div className="overflow-hidden rounded-xl border border-border">
              <dl className="divide-y divide-border-subtle text-sm">
                <Row label={`${t("netPremium")} (${quote.Coverages.join(", ")})`} value={formatBirr(quote.NetPremium, locale)} />
                <Row label={t("taxes")} value={formatBirr(quote.TotalTaxes, locale)} />
                <div className="flex items-center justify-between bg-subtle/60 px-5 py-4">
                  <dt className="font-bold uppercase tracking-wide text-fg">{t("gross")}</dt>
                  <dd className="font-mono text-lg font-bold text-brand-fg">{formatBirr(quote.GrossPremium, locale)}</dd>
                </div>
              </dl>
            </div>
            <div className="flex justify-between border-t border-border-subtle pt-5">
              <Button variant="ghost" onClick={() => setStep(1)}>{tc("back")}</Button>
              <Button onClick={bind} disabled={quote.Status !== "QUOTED"}>
                {t("bind")} <ArrowRight />
              </Button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="animate-rise flex flex-col items-center gap-6 py-6 text-center">
            <span className="grid size-20 place-items-center rounded-full bg-success-subtle text-success ring-8 ring-success-subtle/40">
              <CheckCircle2 className="size-10" />
            </span>
            <div>
              <h2 className="text-2xl">{t("issuedTitle")}</h2>
              <p className="mx-auto mt-1 max-w-md text-fg-muted">{t("issuedSub")}</p>
            </div>
            <div className="grid w-full grid-cols-1 gap-3 sm:grid-cols-2">
              <Meta label={t("policyNumber")} value={policyNumber} mono />
              <Meta label={t("premium")} value={quote ? formatBirr(quote.GrossPremium, locale) : "—"} mono />
              {policyMeta && (
                <div className="sm:col-span-2">
                  <Meta
                    label={t("coverPeriod")}
                    value={`${new Date(policyMeta.from).toLocaleDateString()} — ${new Date(policyMeta.to).toLocaleDateString()}`}
                  />
                </div>
              )}
            </div>
            <div className="flex flex-wrap justify-center gap-3">
              <Button asChild variant="secondary">
                <Link href="/customer/policies"><FileText /> {tNav("policies")}</Link>
              </Button>
              <Button asChild>
                <Link href="/customer">{t("toDashboard")}</Link>
              </Button>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between px-5 py-4">
      <dt className="text-fg-muted">{label}</dt>
      <dd className="font-mono font-medium text-fg">{value}</dd>
    </div>
  );
}

function Meta({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-xl border border-border bg-subtle/50 p-4 text-left">
      <div className="text-xs font-bold uppercase tracking-wider text-fg-subtle">{label}</div>
      <div className={`mt-1 text-lg font-bold text-fg ${mono ? "font-mono" : ""}`}>{value}</div>
    </div>
  );
}
