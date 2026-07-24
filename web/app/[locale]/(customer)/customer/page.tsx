"use client";

import { useLocale, useTranslations } from "next-intl";
import { ArrowRight, ClipboardCheck, FileText, ShieldCheck } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useAuth } from "@/components/providers";
import { useRecents } from "@/lib/recents";
import { formatBirr, relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/states";

export default function CustomerDashboard() {
  const t = useTranslations();
  const locale = useLocale() as Locale;
  const { user } = useAuth();
  const policies = useRecents("policy");
  const claims = useRecents("claim");

  const totalPremium = policies.reduce((s, p) => s + (p.premium ?? 0), 0);

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader
        eyebrow={<Eyebrow>{t("roles.customer")}</Eyebrow>}
        title={user ? `${t("nav.dashboard")} · ${user.name.split(" ")[0]}` : t("nav.dashboard")}
        subtitle={t("brand.tagline")}
        actions={
          <Button asChild>
            <Link href="/customer/quote"><FileText /> {t("marketing.ctaQuote")}</Link>
          </Button>
        }
      />

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label={t("nav.policies")} value={String(policies.length)} icon={<ShieldCheck />} tone="brand" />
        <StatCard label={t("quote.premium")} value={formatBirr(totalPremium, locale, { compact: true })} icon={<FileText />} tone="success" />
        <StatCard label={t("nav.claims")} value={String(claims.length)} icon={<ClipboardCheck />} tone="warning" />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle>{t("nav.policies")}</CardTitle>
            <Button asChild variant="ghost" size="sm">
              <Link href="/customer/policies">{t("common.viewAll")} <ArrowRight /></Link>
            </Button>
          </CardHeader>
          <CardContent className="pt-0">
            {policies.length === 0 ? (
              <EmptyState title={t("common.empty")} description={t("brand.tagline")} action={{ label: t("marketing.ctaQuote"), href: "/customer/quote" }} />
            ) : (
              <ul className="divide-y divide-border-subtle">
                {policies.slice(0, 4).map((p) => (
                  <li key={p.id}>
                    <Link href={`/customer/policies/${p.id}`} className="flex items-center justify-between gap-3 py-3 transition-colors hover:opacity-80">
                      <div className="min-w-0">
                        <p className="truncate font-mono text-sm font-bold text-fg">{p.policyNumber}</p>
                        <p className="text-xs text-fg-muted">{relativeTime(p.createdAt, locale)}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="font-mono text-sm font-medium text-fg">{formatBirr(p.premium, locale)}</span>
                        <StatusBadge value={p.status} />
                      </div>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle>{t("nav.claims")}</CardTitle>
            <Button asChild variant="ghost" size="sm">
              <Link href="/customer/claims">{t("common.viewAll")} <ArrowRight /></Link>
            </Button>
          </CardHeader>
          <CardContent className="pt-0">
            {claims.length === 0 ? (
              <EmptyState title={t("common.empty")} description={t("claim.subtitle")} action={{ label: t("claim.title"), href: "/customer/claims" }} />
            ) : (
              <ul className="divide-y divide-border-subtle">
                {claims.slice(0, 4).map((c) => (
                  <li key={c.id}>
                    <Link href={`/customer/claims/${c.id}`} className="flex items-center justify-between gap-3 py-3 transition-colors hover:opacity-80">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-fg">{c.description || c.id.slice(0, 8)}</p>
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
