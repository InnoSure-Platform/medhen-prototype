"use client";

import { signIn, signOut, useSession, SessionProvider } from "next-auth/react";
import { createContext, useContext, useEffect, useMemo, type ReactNode } from "react";

import { usePathname, useRouter } from "next/navigation";

type AuthUser = {
  username: string;
  accessToken: string;
};

type AuthCtx = {
  user: AuthUser | null;
  loading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  getToken: () => string | null;
};

const Ctx = createContext<AuthCtx>({
  user: null,
  loading: true,
  login: async () => undefined,
  logout: async () => undefined,
  getToken: () => null,
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
    if (!session?.accessToken) return null;
    return {
      username: session.user?.name || session.user?.email || "User",
      accessToken: session.accessToken,
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
        const idToken = (session as any)?.idToken;
        await signOut({ redirect: false });
        if (idToken) {
          window.location.href = `/api/auth/federated-logout?idToken=${idToken}`;
        } else {
          window.location.href = "/";
        }
      },
      getToken: () => user?.accessToken ?? null,
    }),
    [user, loading, pathname, session]
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
