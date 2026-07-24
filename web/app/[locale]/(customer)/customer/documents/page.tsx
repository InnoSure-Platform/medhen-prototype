"use client";

import { useTranslations } from "next-intl";
import { Download, FileText } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useRecents } from "@/lib/recents";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/states";

export default function DocumentsPage() {
  const t = useTranslations();
  const policies = useRecents("policy");

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("roles.customer")}</Eyebrow>} title={t("nav.documents")} subtitle={t("marketing.feature.bilingualBody")} />

      {policies.length === 0 ? (
        <Card><EmptyState title={t("common.empty")} action={{ label: t("marketing.ctaQuote"), href: "/customer/quote" }} /></Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {policies.map((p) => (
            <Card key={p.id} interactive>
              <CardContent className="flex items-center justify-between gap-3 p-5">
                <div className="flex items-center gap-3">
                  <span className="grid size-10 place-items-center rounded-xl border border-success/20 bg-success-subtle text-success-fg">
                    <FileText className="size-5" />
                  </span>
                  <div>
                    <p className="text-sm font-semibold text-fg">Certificate of Insurance</p>
                    <Link href={`/customer/policies/${p.id}`} className="font-mono text-xs text-fg-muted hover:text-brand-fg">
                      {p.policyNumber}
                    </Link>
                  </div>
                </div>
                <Button variant="ghost" size="icon-sm" aria-label={t("common.download")}>
                  <Download />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
