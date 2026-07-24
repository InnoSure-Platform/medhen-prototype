"use client";

import { useTranslations } from "next-intl";
import { Package, Users } from "lucide-react";
import { useProducts, useUsers } from "@/lib/api/hooks";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { StatCard } from "@/components/ui/stat-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function AdminDashboard() {
  const t = useTranslations();
  const users = useUsers();
  const products = useProducts();

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("admin.eyebrow")}</Eyebrow>} title={t("admin.title")} subtitle={t("admin.subtitle")} />

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label={t("admin.totalUsers")} value={String(users.data?.length ?? 0)} icon={<Users />} tone="brand" />
        <StatCard label={t("admin.totalProducts")} value={String(products.data?.length ?? 0)} icon={<Package />} tone="info" />
        <StatCard label={t("admin.tenants")} value="1" tone="success" />
      </div>

      <Card>
        <CardHeader><CardTitle>{t("admin.featureFlags")}</CardTitle></CardHeader>
        <CardContent className="flex flex-wrap gap-2 pt-0">
          {["telebirr", "amharic-docs", "stp-underwriting", "audit-trail", "rls"].map((f) => (
            <Badge key={f} tone="success" dot>{f}</Badge>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
