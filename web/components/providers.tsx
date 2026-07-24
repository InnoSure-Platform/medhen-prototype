"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { SessionProvider, signIn, useSession } from "next-auth/react";
import { ThemeProvider } from "next-themes";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster, toast } from "sonner";
import { useTranslations } from "next-intl";
import { makeQueryClient } from "@/lib/query";
import { usePathname } from "@/lib/i18n/navigation";
import { hasPermission, homeForRole, type Permission, type Role } from "@/lib/auth/permissions";
import { IdleTimeout } from "@/components/auth/idle-timeout";

// Keycloak issuer exposed to the client only for the account-console deep link
// (no secrets). e.g. http://localhost:8081/realms/medhen
const KEYCLOAK_ISSUER = process.env.NEXT_PUBLIC_KEYCLOAK_ISSUER ?? "";

// ---- Auth context ----------------------------------------------------------
type AuthUser = { name: string; role: Role };
type AuthCtx = {
  user: AuthUser | null;
  role: Role;
  home: string;
  loading: boolean;
  can: (perm: Permission) => boolean;
  login: (next?: string) => void;
  logout: () => void;
  /** Start Keycloak self-registration (OIDC prompt=create). */
  signUp: (next?: string) => void;
  /** Enroll a TOTP authenticator via Keycloak required action. */
  enrollMfa: (next?: string) => void;
  /** Force a fresh authentication (step-up) before a sensitive action. */
  stepUp: (next?: string) => void;
  /** Keycloak account console URL (manage password / MFA devices), or "". */
  accountUrl: string;
};

const noop = () => undefined;
const Ctx = createContext<AuthCtx>({
  user: null,
  role: "",
  home: "/customer",
  loading: true,
  can: () => false,
  login: noop,
  logout: noop,
  signUp: noop,
  enrollMfa: noop,
  stepUp: noop,
  accountUrl: "",
});

export const useAuth = () => useContext(Ctx);

function AuthBridge({ children }: { children: ReactNode }) {
  const { data: session, status } = useSession();
  const pathname = usePathname();
  const t = useTranslations("auth");
  const reauthing = useRef(false);

  const role = ((session?.role as string) ?? "") as Role;

  const login = useCallback((next?: string) => {
    void signIn("keycloak", { callbackUrl: next ?? pathname ?? "/" });
  }, [pathname]);

  // NextAuth's 3rd signIn arg becomes extra params on the Keycloak authorize URL.
  const signUp = useCallback((next?: string) => {
    void signIn("keycloak", { callbackUrl: next ?? "/customer" }, { prompt: "create" });
  }, []);
  const enrollMfa = useCallback((next?: string) => {
    void signIn("keycloak", { callbackUrl: next ?? pathname ?? "/" }, { kc_action: "CONFIGURE_TOTP" });
  }, [pathname]);
  const stepUp = useCallback((next?: string) => {
    // max_age=0 forces a fresh credential prompt (re-authentication).
    void signIn("keycloak", { callbackUrl: next ?? pathname ?? "/" }, { prompt: "login", max_age: "0" });
  }, [pathname]);

  const logout = useCallback(() => {
    // Server route reads the id token from the encrypted cookie (never the
    // browser), clears the session cookie, and drives Keycloak end-session (H8).
    window.location.href = "/api/auth/federated-logout";
  }, []);

  // Re-authenticate on a 401 from the API, preserving the current location.
  useEffect(() => {
    const onUnauthorized = () => {
      if (reauthing.current) return;
      reauthing.current = true;
      toast.error(t("expired"));
      login(pathname ?? "/");
    };
    window.addEventListener("medhen:unauthorized", onUnauthorized);
    return () => window.removeEventListener("medhen:unauthorized", onUnauthorized);
  }, [login, pathname, t]);

  // Refresh-token rotation failed → session is dead; force re-login.
  useEffect(() => {
    if (session?.error === "RefreshAccessTokenError" && !reauthing.current) {
      reauthing.current = true;
      toast.error(t("expired"));
      login(pathname ?? "/");
    }
  }, [session?.error, login, pathname, t]);

  const value = useMemo<AuthCtx>(() => {
    const user = session?.user ? { name: session.user.name || session.user.email || "User", role } : null;
    return {
      user,
      role,
      home: homeForRole(role),
      loading: status === "loading",
      can: (perm: Permission) => hasPermission(role, perm),
      login,
      logout,
      signUp,
      enrollMfa,
      stepUp,
      accountUrl: KEYCLOAK_ISSUER ? `${KEYCLOAK_ISSUER.replace(/\/$/, "")}/account` : "",
    };
  }, [session, status, role, login, logout, signUp, enrollMfa, stepUp]);

  return (
    <Ctx.Provider value={value}>
      {children}
      {value.user && <IdleTimeout onTimeout={logout} />}
    </Ctx.Provider>
  );
}

// ---- Root client providers -------------------------------------------------
export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(makeQueryClient);

  return (
    <SessionProvider>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
        <QueryClientProvider client={queryClient}>
          <AuthBridge>{children}</AuthBridge>
          <Toaster position="top-right" richColors closeButton toastOptions={{ classNames: { toast: "font-sans" } }} />
        </QueryClientProvider>
      </ThemeProvider>
    </SessionProvider>
  );
}
