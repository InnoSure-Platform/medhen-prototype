"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { captureException } from "@/lib/sentry";
import { Button } from "@/components/ui/button";

export default function LocaleError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  const t = useTranslations("errors");
  useEffect(() => {
    void captureException(error);
  }, [error]);
  return (
    <div className="mesh-bg grid min-h-[60dvh] place-items-center px-6">
      <div className="flex max-w-md flex-col items-center gap-4 text-center">
        <span className="grid size-14 place-items-center rounded-2xl bg-danger-subtle text-danger-fg text-2xl font-bold">!</span>
        <h1 className="text-2xl">{t("boundaryTitle")}</h1>
        <p className="text-fg-muted">{t("boundaryBody")}</p>
        <Button className="mt-2" onClick={reset}>{t("generic")}</Button>
      </div>
    </div>
  );
}
