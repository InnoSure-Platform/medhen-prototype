import { setRequestLocale, getTranslations } from "next-intl/server";
import {
  Activity,
  BadgeCheck,
  BarChart3,
  Check,
  FileCheck2,
  Languages,
  Lock,
  Wallet,
} from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/badge";
import { Eyebrow } from "@/components/patterns/page-header";

export default async function LandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("marketing");
  const tb = await getTranslations("brand");

  const steps = ["quote", "underwrite", "bind", "pay", "fnol", "settle"] as const;
  const features = [
    { icon: Activity, key: "stp" },
    { icon: Wallet, key: "money" },
    { icon: Lock, key: "audit" },
    { icon: BadgeCheck, key: "telebirr" },
    { icon: Languages, key: "bilingual" },
    { icon: BarChart3, key: "kpi" },
  ] as const;

  return (
    <div className="mesh-bg overflow-hidden">
      {/* Hero */}
      <section className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-12 px-6 pb-10 pt-16 lg:grid-cols-2 lg:pt-24">
        <div className="animate-rise flex flex-col items-start gap-6">
          <Eyebrow>
            <span className="size-1.5 rounded-full bg-accent" />
            {t("eyebrow")}
          </Eyebrow>
          <h1 className="text-5xl leading-[1.08] lg:text-6xl">
            {t("heroLead")}
            <br />
            <span className="bg-gradient-to-r from-brand-600 to-brand-400 bg-clip-text text-transparent">
              {t("heroAccent")}
            </span>
          </h1>
          <p className="max-w-lg text-lg leading-relaxed text-fg-muted">{tb("tagline")}</p>
          <div className="mt-2 flex flex-wrap items-center gap-3">
            <Button asChild size="lg">
              <Link href="/customer/quote">
                {t("ctaQuote")}
                <FileCheck2 />
              </Link>
            </Button>
            <Button asChild size="lg" variant="secondary">
              <Link href="/customer/claims">{t("ctaClaim")}</Link>
            </Button>
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-fg-muted">
            {[t("trust1"), t("trust2"), t("trust3")].map((label) => (
              <span key={label} className="inline-flex items-center gap-2">
                <Check className="size-4 text-success" strokeWidth={3} /> {label}
              </span>
            ))}
          </div>
        </div>

        {/* Floating premium quote card */}
        <div className="animate-rise relative [animation-delay:0.15s]">
          <div className="absolute -inset-4 -z-10 rounded-[2.5rem] bg-gradient-to-tr from-brand-subtle via-surface to-accent-subtle blur-xl" />
          <Card className="animate-float p-7">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs font-semibold uppercase tracking-wider text-fg-subtle">Motor quote</div>
                <div className="mt-1 font-mono text-sm font-bold text-brand-fg">EIC/MOT/2026/000001</div>
              </div>
              <StatusBadge value="ISSUED" />
            </div>
            <div className="my-5 h-px bg-border-subtle" />
            <dl className="space-y-3 text-sm">
              {[
                ["Own damage (OD)", "1,500.00"],
                ["Third-party (TPL)", "800.00"],
                ["VAT 15%", "345.00"],
                ["Stamp duty", "35.00"],
              ].map(([k, v]) => (
                <div key={k} className="flex items-center justify-between">
                  <dt className="text-fg-muted">{k}</dt>
                  <dd className="font-mono font-medium text-fg">{v}</dd>
                </div>
              ))}
              <div className="flex items-center justify-between border-t border-border-subtle pt-3">
                <dt className="font-semibold text-fg">Gross premium (ETB)</dt>
                <dd className="font-mono text-lg font-bold text-brand-fg">2,680.00</dd>
              </div>
            </dl>
          </Card>
        </div>
      </section>

      {/* Lifecycle */}
      <section className="mx-auto max-w-7xl px-6 py-14">
        <Card className="p-7">
          <Eyebrow className="mb-6">{t("lifecycleTitle")}</Eyebrow>
          <ol className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
            {steps.map((step, i) => (
              <li key={step} className="flex items-center gap-3">
                <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-brand text-sm font-bold text-fg-onbrand shadow-[var(--shadow-lift)]">
                  {i + 1}
                </span>
                <span className="text-sm font-semibold text-fg">{t(`step.${step}`)}</span>
              </li>
            ))}
          </ol>
        </Card>
      </section>

      {/* Features */}
      <section className="mx-auto max-w-7xl px-6 pb-20">
        <h2 className="mb-6 font-display text-2xl font-bold tracking-tight text-fg">{t("featuresTitle")}</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {features.map(({ icon: Icon, key }) => (
            <Card key={key} interactive className="p-7">
              <span className="mb-4 grid size-11 place-items-center rounded-xl border border-brand-border bg-brand-subtle text-brand-fg">
                <Icon className="size-5" />
              </span>
              <h3 className="text-lg font-bold text-fg">{t(`feature.${key}Title`)}</h3>
              <p className="mt-2 text-sm leading-relaxed text-fg-muted">{t(`feature.${key}Body`)}</p>
            </Card>
          ))}
        </div>
      </section>
    </div>
  );
}
