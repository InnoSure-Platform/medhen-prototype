"use client";

import { useTranslations } from "next-intl";
import { ShieldX } from "lucide-react";
import { Link } from "@/lib/i18n/navigation";
import { useAuth } from "@/components/providers";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

export default function ForbiddenPage() {
  const t = useTranslations("errors");
  const { home, user } = useAuth();

  return (
    <div className="mesh-bg grid min-h-[calc(100dvh-4rem)] place-items-center px-6 py-16">
      <Card className="w-full max-w-md p-8 text-center">
        <div className="flex flex-col items-center gap-4">
          <span className="grid size-14 place-items-center rounded-2xl bg-danger-subtle text-danger-fg">
            <ShieldX className="size-7" />
          </span>
          <span className="font-mono text-sm font-bold text-fg-subtle">403</span>
          <h1 className="text-2xl">{t("forbiddenTitle")}</h1>
          <p className="text-fg-muted">{t("forbiddenBody")}</p>
          <Button asChild className="mt-2">
            <Link href={user ? home : "/"}>{t("goHome")}</Link>
          </Button>
        </div>
      </Card>
    </div>
  );
}
