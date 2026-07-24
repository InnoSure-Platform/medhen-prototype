"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/lib/i18n/navigation";
import { useAuth } from "@/components/providers";
import { Logo } from "./logo";
import { ThemeToggle } from "./theme-toggle";
import { LocaleToggle } from "./locale-toggle";
import { UserMenu } from "./user-menu";
import { Button } from "@/components/ui/button";

export function MarketingNav() {
  const t = useTranslations();
  const { user } = useAuth();

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-surface/75 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2.5">
          <Logo size={34} />
          <span className="leading-none">
            <span className="font-display text-[17px] font-extrabold tracking-tight text-fg">
              {t("brand.name")} <span className="text-brand-fg">·</span> EIC
            </span>
            <span className="mt-1 block text-[11px] text-fg-muted">{t("brand.org")}</span>
          </span>
        </Link>

        <div className="flex items-center gap-1 sm:gap-2">
          <LocaleToggle />
          <ThemeToggle />
          <div className="mx-1 h-6 w-px bg-border" />
          {user ? (
            <UserMenu />
          ) : (
            <Button asChild size="sm">
              <Link href="/login">{t("common.signIn")}</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
