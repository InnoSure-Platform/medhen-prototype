"use client";

import Link from "next/link";
import { useLocale } from "@/components/Shell";
import { t } from "@/lib/i18n";

export default function HomePage() {
  const { locale } = useLocale();
  return (
    <section className="hero">
      <div className="hero-copy">
        <div className="eyebrow">Ethiopian Insurance Corporation</div>
        <h1>{t("product", locale)}</h1>
        <p>{t("tagline", locale)}</p>
        <div className="cta-row">
          <Link className="btn btn-primary" href="/quote">{t("ctaQuote", locale)}</Link>
          <Link className="btn btn-ghost" href="/claim">{t("ctaClaim", locale)}</Link>
        </div>
      </div>
      <div className="hero-visual">
        <div className="panel">
          <h2>{locale === "am" ? "የጋራ መድረክ" : "Shared core"}</h2>
          <p>{t("heroSide", locale)}</p>
        </div>
      </div>
    </section>
  );
}
