"use client";

import { signIn, useSession, SessionProvider } from "next-auth/react";
import { createContext, useContext, useEffect, useMemo, type ReactNode } from "react";

import { usePathname, useRouter } from "next/navigation";

type AuthUser = {
  username: string;
};

type AuthCtx = {
  user: AuthUser | null;
  loading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
};

const Ctx = createContext<AuthCtx>({
  user: null,
  loading: true,
  login: async () => undefined,
  logout: async () => undefined,
});

export function useAuth() {
  return useContext(Ctx);
}

function InnerProvider({ children }: { children: ReactNode }) {
  const { data: session, status } = useSession();
  const loading = status === "loading";

  const router = useRouter();
  const pathname = usePathname();
  const authRequired = process.env.NEXT_PUBLIC_AUTH_REQUIRED !== "false";

  const user = useMemo(() => {
    // The access token is no longer on the session (C5); presence of an
    // authenticated user is signalled by session.user.
    if (!session?.user) return null;
    return {
      username: session.user?.name || session.user?.email || "User",
    };
  }, [session]);

  const isPublic = (p: string | null) => p && (p === "/" || p === "/login");

  useEffect(() => {
    if (!authRequired || loading) return;
    if (!user && pathname && !isPublic(pathname)) {
      router.replace("/login?next=" + encodeURIComponent(pathname));
    }
  }, [user, loading, pathname, router, authRequired]);

  const value = useMemo(
    () => ({
      user,
      loading,
      login: async () => { await signIn("keycloak", { callbackUrl: pathname !== "/login" ? pathname : "/quote" }); },
      logout: async () => {
        // The server-side route reads the id token from the encrypted cookie,
        // clears the session cookie, and drives the Keycloak logout. The id
        // token is never placed in the browser or the URL (H8).
        window.location.href = "/api/auth/federated-logout";
      },
    }),
    [user, loading, pathname]
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <InnerProvider>{children}</InnerProvider>
    </SessionProvider>
  );
}
