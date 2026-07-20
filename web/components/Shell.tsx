"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { createContext, useContext, useMemo, useState, type ReactNode } from "react";
import { useAuth } from "@/components/AuthProvider";
import { type Locale, t } from "@/lib/i18n";

const LocaleCtx = createContext<{ locale: Locale; setLocale: (l: Locale) => void }>({
  locale: "en",
  setLocale: () => undefined,
});

export function useLocale() {
  return useContext(LocaleCtx);
}

export function Shell({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>("en");
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const value = useMemo(() => ({ locale, setLocale }), [locale]);

  return (
    <LocaleCtx.Provider value={value}>
      <div className="app-shell">
        <header className="topbar">
          <div className="brand">
            <strong>{t("product", locale)} · {t("brand", locale)}</strong>
            <span>{t("tenantTag", locale)}</span>
          </div>
          <nav className="nav">
            <Link className={pathname === "/" ? "active" : ""} href="/">{t("product", locale)}</Link>
            <Link className={pathname?.startsWith("/quote") ? "active" : ""} href="/quote">{t("navQuote", locale)}</Link>
            <Link className={pathname?.startsWith("/claim") ? "active" : ""} href="/claim">{t("navClaim", locale)}</Link>
            <Link className={pathname?.startsWith("/staff") ? "active" : ""} href="/staff">{t("navStaff", locale)}</Link>
            <button type="button" className="locale-btn" onClick={() => setLocale(locale === "en" ? "am" : "en")}>
              {locale === "en" ? "አማርኛ" : "English"}
            </button>
            {user ? (
              <button type="button" className="locale-btn" onClick={() => logout()}>
                {user.username} · {t("logout", locale)}
              </button>
            ) : (
              <Link className={pathname === "/login" ? "active" : ""} href="/login">{t("loginNav", locale)}</Link>
            )}
          </nav>
        </header>
        <main>{children}</main>
      </div>
    </LocaleCtx.Provider>
  );
}
