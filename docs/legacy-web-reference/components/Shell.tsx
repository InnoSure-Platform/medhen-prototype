"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { createContext, useContext, useMemo, useState, type ReactNode } from "react";
import { useAuth } from "@/components/AuthProvider";
import { Logo } from "@/components/ui/Logo";
import { cn } from "@/lib/utils";
import { type Locale, t } from "@/lib/i18n";

const LocaleCtx = createContext<{ locale: Locale; setLocale: (l: Locale) => void }>({
  locale: "en",
  setLocale: () => undefined,
});

export function useLocale() {
  return useContext(LocaleCtx);
}

const NAV = [
  { href: "/", key: "product", exact: true },
  { href: "/quote", key: "navQuote" },
  { href: "/claim", key: "navClaim" },
  { href: "/staff", key: "navStaff" },
] as const;

export function Shell({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>("en");
  const pathname = usePathname();
  const { user, login, logout } = useAuth();
  const value = useMemo(() => ({ locale, setLocale }), [locale]);

  const isActive = (href: string, exact?: boolean) =>
    exact ? pathname === href : pathname?.startsWith(href);

  return (
    <LocaleCtx.Provider value={value}>
      <div className="app-shell">
        <header className="topbar">
          {/* Brand */}
          <Link href="/" className="brand-lockup group">
            <Logo size={34} className="transition-transform group-hover:scale-105" />
            <span className="flex flex-col">
              <span className="brand-name">
                {t("product", locale)} <span className="text-brand-600">·</span> EIC
              </span>
              <span className="brand-sub">{t("brand", locale)}</span>
            </span>
          </Link>

          {/* Primary nav */}
          <nav className="nav" aria-label="Primary">
            {NAV.map((n) => (
              <Link
                key={n.href}
                href={n.href}
                className={cn("nav-link", isActive(n.href, (n as { exact?: boolean }).exact) && "active")}
              >
                {n.key === "product" ? (locale === "am" ? "መነሻ" : "Overview") : t(n.key, locale)}
              </Link>
            ))}
          </nav>

          {/* Actions */}
          <div className="flex items-center gap-2 sm:gap-3">
            <span className="chip hidden lg:inline-flex">
              <span className="h-1.5 w-1.5 rounded-full bg-success-500" />
              tenant&nbsp;·&nbsp;eic
            </span>
            <button
              type="button"
              className="locale-btn"
              onClick={() => setLocale(locale === "en" ? "am" : "en")}
              aria-label="Toggle language"
            >
              {locale === "en" ? "አማርኛ" : "EN"}
            </button>
            {user ? (
              <div className="flex items-center gap-2">
                <span className="hidden items-center gap-2 rounded-full border border-slate-200 bg-white py-1 pl-1 pr-3 shadow-[var(--shadow-soft)] sm:flex">
                  <span className="grid h-7 w-7 place-items-center rounded-full bg-brand-600 text-xs font-bold text-white">
                    {user.username.slice(0, 1).toUpperCase()}
                  </span>
                  <span className="max-w-[10rem] truncate text-sm font-medium text-slate-700">{user.username}</span>
                </span>
                <button type="button" className="icon-btn" onClick={() => logout()}>
                  {t("logout", locale)}
                </button>
              </div>
            ) : (
              <button type="button" className="btn btn-primary btn-sm" onClick={() => login()}>
                {t("loginNav", locale)}
              </button>
            )}
          </div>
        </header>

        <main>{children}</main>

        <footer className="mt-auto border-t border-slate-200/70 bg-white/60 py-6">
          <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-2 px-6 text-xs text-slate-400 sm:flex-row">
            <span>© {new Date().getFullYear()} Ethiopian Insurance Corporation · Medhen</span>
            <span className="font-mono">Motor · quote → bind → issue → pay → FNOL → settle</span>
          </div>
        </footer>
      </div>
    </LocaleCtx.Provider>
  );
}
