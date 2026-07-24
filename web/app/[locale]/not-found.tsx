import { getTranslations } from "next-intl/server";
import { Link } from "@/lib/i18n/navigation";

export default async function NotFound() {
  const t = await getTranslations("errors");
  return (
    <main className="mesh-bg grid min-h-dvh place-items-center px-6">
      <div className="flex max-w-md flex-col items-center gap-4 text-center">
        <span className="font-mono text-6xl font-bold text-brand-fg">404</span>
        <h1 className="text-2xl">{t("notFoundTitle")}</h1>
        <p className="text-fg-muted">{t("notFoundBody")}</p>
        <Link
          href="/"
          className="mt-2 rounded-xl bg-brand px-5 py-2.5 text-sm font-semibold text-fg-onbrand transition hover:bg-brand-hover"
        >
          ← Home
        </Link>
      </div>
    </main>
  );
}
