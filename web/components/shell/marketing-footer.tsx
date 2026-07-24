import { useTranslations } from "next-intl";

export function MarketingFooter() {
  const t = useTranslations("brand");
  return (
    <footer className="border-t border-border bg-surface/60">
      <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-2 px-6 py-6 text-xs text-fg-muted sm:flex-row">
        <span>© {new Date().getFullYear()} {t("org")} · {t("name")}</span>
        <span className="font-mono">Motor · quote → bind → issue → pay → FNOL → settle</span>
      </div>
    </footer>
  );
}
