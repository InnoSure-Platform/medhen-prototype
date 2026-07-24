"use client";

import { useLocale } from "next-intl";
import { useTransition } from "react";
import { Languages } from "lucide-react";
import { usePathname, useRouter } from "@/lib/i18n/navigation";
import type { Locale } from "@/lib/i18n/routing";
import { Button } from "@/components/ui/button";

export function LocaleToggle() {
  const locale = useLocale() as Locale;
  const pathname = usePathname();
  const router = useRouter();
  const [pending, startTransition] = useTransition();

  const next: Locale = locale === "en" ? "am" : "en";

  return (
    <Button
      variant="ghost"
      size="sm"
      disabled={pending}
      aria-label="Toggle language"
      onClick={() => startTransition(() => router.replace(pathname, { locale: next }))}
    >
      <Languages className="size-4" />
      <span className={next === "am" ? "font-ethiopic" : undefined}>{next === "am" ? "አማርኛ" : "EN"}</span>
    </Button>
  );
}
