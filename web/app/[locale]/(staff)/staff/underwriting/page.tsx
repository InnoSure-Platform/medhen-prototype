"use client";

import { useTranslations } from "next-intl";
import { Ban, Gauge, ShieldAlert, Zap } from "lucide-react";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function UnderwritingPage() {
  const t = useTranslations();

  const rules = [
    { icon: Zap, title: "Straight-through processing", body: "Clean risks auto-accept in milliseconds; no manual touch.", tone: "success" as const, tag: "AUTO-ACCEPT" },
    { icon: Gauge, title: "Refer above threshold", body: "Premiums or sums insured above the configured limit route to a referral queue.", tone: "warning" as const, tag: "REFER" },
    { icon: ShieldAlert, title: "Max prior claims", body: "Applicants exceeding the prior-claims ceiling are referred for manual assessment.", tone: "warning" as const, tag: "REFER" },
    { icon: Ban, title: "Blacklist", body: "Blacklisted parties are declined automatically.", tone: "danger" as const, tag: "DECLINE" },
  ];

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("staff.eyebrow")}</Eyebrow>} title={t("nav.underwriting")} subtitle={t("marketing.feature.stpBody")} />
      <div className="grid gap-4 sm:grid-cols-2">
        {rules.map((r) => (
          <Card key={r.title}>
            <CardContent className="flex gap-4 p-5">
              <span className="grid size-11 shrink-0 place-items-center rounded-xl border border-border bg-subtle text-fg">
                <r.icon className="size-5" />
              </span>
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-fg">{r.title}</h3>
                  <Badge tone={r.tone}>{r.tag}</Badge>
                </div>
                <p className="text-sm text-fg-muted">{r.body}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
