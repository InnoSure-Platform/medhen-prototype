"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { useAuth } from "@/components/AuthProvider";
import { useLocale } from "@/components/Shell";
import { t } from "@/lib/i18n";

function LoginForm() {
  const { login } = useAuth();
  const { locale } = useLocale();
  const router = useRouter();
  const params = useSearchParams();
  const [username, setUsername] = useState("demo-agent");
  const [password, setPassword] = useState("medhen-demo");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr("");
    try {
      await login(username, password);
      router.push(params.get("next") || "/quote");
    } catch (ex) {
      setErr(String(ex));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="section login-section">
      <h1>{t("loginTitle", locale)}</h1>
      <p className="login-sub">{t("loginSub", locale)}</p>
      {err && <div className="banner-err">{err}</div>}
      <form className="stack login-form" onSubmit={submit}>
        <div className="field">
          <label>{t("loginUser", locale)}</label>
          <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" />
        </div>
        <div className="field">
          <label>{t("loginPass", locale)}</label>
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" />
        </div>
        <button className="btn btn-primary" type="submit" disabled={busy}>
          {busy ? "…" : t("loginBtn", locale)}
        </button>
      </form>
      <p className="login-hint">{t("loginHint", locale)}</p>
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
