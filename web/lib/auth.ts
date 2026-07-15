export type AuthUser = {
  username: string;
  accessToken: string;
  expiresAt: number;
};

const STORAGE_KEY = "medhen_auth";

export function loadAuth(): AuthUser | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const u = JSON.parse(raw) as AuthUser;
    if (u.expiresAt < Date.now()) {
      localStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return u;
  } catch {
    return null;
  }
}

export function saveAuth(user: AuthUser) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
}

export function clearAuth() {
  localStorage.removeItem(STORAGE_KEY);
}

export async function login(username: string, password: string): Promise<AuthUser> {
  const res = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) {
    const t = await res.text();
    throw new Error(t || "Login failed");
  }
  const data = (await res.json()) as { access_token: string; expires_in: number; preferred_username?: string };
  const user: AuthUser = {
    username: data.preferred_username ?? username,
    accessToken: data.access_token,
    expiresAt: Date.now() + data.expires_in * 1000 - 30_000,
  };
  saveAuth(user);
  return user;
}

export async function logout() {
  clearAuth();
}
