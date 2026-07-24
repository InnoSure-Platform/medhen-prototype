"use client";

import { useTranslations } from "next-intl";
import { Building2 } from "lucide-react";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

const FLAGS = [
  { key: "telebirr", label: "Telebirr payments", on: true },
  { key: "amharic-docs", label: "Bilingual documents", on: true },
  { key: "stp-underwriting", label: "Straight-through underwriting", on: true },
  { key: "rls", label: "Runtime row-level security", on: false },
];

export default function AdminSettingsPage() {
  const t = useTranslations();

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("admin.eyebrow")}</Eyebrow>} title={t("admin.settingsTitle")} subtitle={t("admin.subtitle")} />

      <Card>
        <CardHeader><CardTitle>{t("admin.tenants")}</CardTitle></CardHeader>
        <CardContent className="pt-0">
          <div className="flex items-center justify-between rounded-xl border border-border bg-subtle/50 p-4">
            <div className="flex items-center gap-3">
              <span className="grid size-10 place-items-center rounded-xl border border-brand-border bg-brand-subtle text-brand-fg">
                <Building2 className="size-5" />
              </span>
              <div>
                <p className="font-semibold text-fg">Ethiopian Insurance Corporation</p>
                <p className="font-mono text-xs text-fg-muted">tenant · eic</p>
              </div>
            </div>
            <Badge tone="success" dot>active</Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>{t("admin.featureFlags")}</CardTitle></CardHeader>
        <CardContent className="space-y-3 pt-0">
          {FLAGS.map((f) => (
            <div key={f.key} className="flex items-center justify-between border-b border-border-subtle py-3 last:border-0">
              <Label htmlFor={`flag-${f.key}`} className="flex flex-col gap-0.5">
                <span>{f.label}</span>
                <span className="font-mono text-xs text-fg-subtle">{f.key}</span>
              </Label>
              <Switch id={`flag-${f.key}`} defaultChecked={f.on} />
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
