"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowRight, ShieldCheck, UserPlus } from "lucide-react";
import { useAuth } from "@/components/providers";
import { Logo } from "@/components/shell/logo";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Alert } from "@/components/ui/alert";

export default function LoginPage() {
  return (
    <Suspense>
      <LoginInner />
    </Suspense>
  );
}

function LoginInner() {
  const t = useTranslations("auth");
  const locale = useLocale();
  const params = useSearchParams();
  const router = useRouter();
  const { login, signUp, loading, user, home } = useAuth();
  const next = params.get("next");

  // Already authenticated → route to the intended destination, or the portal
  // home for this user's role.
  useEffect(() => {
    if (user) router.replace(next ?? `/${locale}${home}`);
  }, [user, next, home, locale, router]);

  // Deep-link preserved through Keycloak; otherwise return here to route by role.
  const dest = next ?? `/${locale}/login`;

  return (
    <div className="mesh-bg grid min-h-[calc(100dvh-4rem)] place-items-center px-6 py-16">
      <Card className="w-full max-w-md p-8">
        <div className="flex flex-col items-center gap-4 text-center">
          <Logo size={48} />
          <div>
            <h1 className="text-2xl">{t("title")}</h1>
            <p className="mt-1 text-sm text-fg-muted">{t("subtitle")}</p>
          </div>
        </div>

        <div className="mt-8 space-y-4">
          <Button className="w-full" size="lg" loading={loading} onClick={() => login(dest)}>
            <ShieldCheck />
            {t("cta")}
            <ArrowRight />
          </Button>

          <button
            type="button"
            onClick={() => login(dest)}
            className="w-full text-center text-sm font-medium text-brand-fg hover:underline"
          >
            {t("forgot")}
          </button>

          <div className="flex items-center gap-3 text-xs text-fg-subtle">
            <span className="h-px flex-1 bg-border" />
            {t("noAccount")}
            <span className="h-px flex-1 bg-border" />
          </div>

          <Button variant="secondary" className="w-full" size="lg" onClick={() => signUp(next ?? "/customer")}>
            <UserPlus />
            {t("createAccount")}
          </Button>

          <Alert tone="info">{t("notice")}</Alert>
          <p className="text-center font-mono text-xs text-fg-subtle">{t("demoHint")}</p>
        </div>
      </Card>
    </div>
  );
}
