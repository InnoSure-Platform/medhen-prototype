"use client";

import { useRouter, usePathname } from "next/navigation";
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { clearAuth, loadAuth, login as doLogin, logout as doLogout, type AuthUser } from "@/lib/auth";

type AuthCtx = {
  user: AuthUser | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
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

const PUBLIC_PATHS = new Set(["/", "/login"]);

function isPublic(path: string | null) {
  return path != null && PUBLIC_PATHS.has(path);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();
  const authRequired = process.env.NEXT_PUBLIC_AUTH_REQUIRED !== "false";

  useEffect(() => {
    setUser(loadAuth());
    setLoading(false);
  }, []);

  useEffect(() => {
    if (!authRequired || loading) return;
    if (!user && pathname && !isPublic(pathname)) {
      router.replace("/login?next=" + encodeURIComponent(pathname));
    }
  }, [user, loading, pathname, router, authRequired]);

  const login = useCallback(async (username: string, password: string) => {
    const u = await doLogin(username, password);
    setUser(u);
  }, []);

  const logout = useCallback(async () => {
    await doLogout();
    setUser(null);
    router.push("/login");
  }, [router]);

  const value = useMemo(
    () => ({
      user,
      loading,
      login,
      logout,
      getToken: () => user?.accessToken ?? null,
    }),
    [user, loading, login, logout],
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}
