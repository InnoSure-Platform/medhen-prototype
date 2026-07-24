"use client";

import { useTranslations } from "next-intl";
import { useProducts } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge, StatusBadge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorState } from "@/components/ui/states";

export default function AdminProductsPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const { data, isLoading, isError, error, refetch } = useProducts();

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-6 py-8">
      <PageHeader eyebrow={<Eyebrow>{t("admin.eyebrow")}</Eyebrow>} title={t("admin.productsTitle")} subtitle={t("admin.subtitle")} />

      {isLoading && <Skeleton className="h-64 w-full" />}
      {isError && <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(error, tErr)} action={{ label: t("common.retry"), onClick: () => refetch() }} /></Card>}

      <div className="space-y-6">
        {(data ?? []).map((p) => (
          <Card key={p.Code}>
            <CardHeader className="flex-row items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  {p.Name} <span className="font-mono text-sm text-fg-muted">{p.Code}</span>
                </CardTitle>
                <p className="mt-1 font-ethiopic text-sm text-fg-muted">{p.NameAmharic}</p>
              </div>
              <div className="flex items-center gap-2">
                <Badge tone="neutral">{p.LOB}</Badge>
                <StatusBadge value={p.Status} />
                <Badge tone="brand">v{p.RateVersion}</Badge>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("admin.coverages")}</TableHead>
                    <TableHead>Code</TableHead>
                    <TableHead className="text-right">{t("admin.baseRate")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(p.Coverages ?? []).map((c) => (
                    <TableRow key={c.Code}>
                      <TableCell>{c.Name} <span className="font-ethiopic text-fg-muted">· {c.NameAmharic}</span></TableCell>
                      <TableCell className="font-mono text-fg-muted">{c.Code}</TableCell>
                      <TableCell className="text-right font-mono">{c.BaseRate}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
