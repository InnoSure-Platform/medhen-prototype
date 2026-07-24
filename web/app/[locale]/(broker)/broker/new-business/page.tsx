"use client";

import { useTranslations } from "next-intl";
import { ArrowRight, FileText, Gauge, ShieldCheck } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export default function NewBusinessPage() {
  const t = useTranslations();
  const steps = [
    { icon: ShieldCheck, title: t("quote.step.identity"), body: t("quote.identitySub") },
    { icon: FileText, title: t("quote.step.asset"), body: t("quote.assetSub") },
    { icon: Gauge, title: t("quote.step.quotation"), body: t("marketing.feature.stpBody") },
  ];

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.broker")}</Eyebrow>} title={t("nav.newBusiness")} subtitle={t("brand.tagline")} />
      <div className="grid gap-4 sm:grid-cols-3">
        {steps.map((s, i) => (
          <Card key={i}>
            <CardContent className="space-y-2 p-5">
              <span className="grid size-10 place-items-center rounded-xl border border-brand-border bg-brand-subtle text-brand-fg">
                <s.icon className="size-5" />
              </span>
              <h3 className="font-semibold text-fg">{s.title}</h3>
              <p className="text-sm text-fg-muted">{s.body}</p>
            </CardContent>
          </Card>
        ))}
      </div>
      <Card className="bg-gradient-to-tr from-brand-subtle to-surface">
        <CardContent className="flex flex-col items-start gap-4 p-8">
          <h2 className="text-xl font-bold text-fg">{t("marketing.ctaQuote")}</h2>
          <p className="max-w-lg text-fg-muted">{t("quote.identitySub")}</p>
          <Button asChild size="lg">
            <Link href="/customer/quote">{t("marketing.ctaQuote")} <ArrowRight /></Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
