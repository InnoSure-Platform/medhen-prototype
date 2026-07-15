"use client";

import { useSearchParams } from "next/navigation";
import { Suspense, useEffect } from "react";
import { useAuth } from "@/components/AuthProvider";
import { useLocale } from "@/components/Shell";
import { t } from "@/lib/i18n";

function LoginForm() {
  const { login, loading, user } = useAuth();
  const { locale } = useLocale();

  useEffect(() => {
    if (!loading && !user) {
      login();
    }
  }, [loading, user, login]);

  return (
    <div className="section login-section">
      <h1>{t("loginTitle", locale)}</h1>
      <p className="login-sub">Redirecting to EIC Identity Provider...</p>
      <div style={{ marginTop: "2rem" }}>
        <button className="btn btn-primary" onClick={login} disabled={loading}>
          {t("loginBtn", locale)}
        </button>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
