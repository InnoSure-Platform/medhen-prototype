import createIntlMiddleware from "next-intl/middleware";
import { getToken } from "next-auth/jwt";
import { NextResponse, type NextRequest } from "next/server";
import { routing } from "@/lib/i18n/routing";

const intlMiddleware = createIntlMiddleware(routing);

// Role requirements per (locale-stripped) route prefix. A protected prefix not
// listed here only requires an authenticated session (any role).
const roleGuards: { prefix: string; roles: string[] }[] = [
  { prefix: "/admin", roles: ["admin"] },
  { prefix: "/broker", roles: ["broker", "admin"] },
  { prefix: "/finance", roles: ["finance", "staff", "admin"] },
  // Staff workbench (underwriting, settlement, audit) is the most privileged surface.
  { prefix: "/staff", roles: ["staff", "claims", "admin"] },
];

// Any-authenticated prefixes (plus everything guarded above).
const protectedPrefixes = ["/customer", "/broker", "/staff", "/admin", "/finance"];

const matches = (rest: string, prefix: string) => rest === prefix || rest.startsWith(prefix + "/");

export default async function middleware(req: NextRequest) {
  // 1) Locale negotiation / rewrite (next-intl).
  const res = intlMiddleware(req);

  // 2) Strip the locale segment to evaluate route protection.
  const segments = req.nextUrl.pathname.split("/");
  const locale = routing.locales.includes(segments[1] as never) ? segments[1] : routing.defaultLocale;
  const rest = "/" + segments.slice(routing.locales.includes(segments[1] as never) ? 2 : 1).join("/");

  const needsAuth = protectedPrefixes.some((p) => matches(rest, p));
  if (!needsAuth) return res;

  const token = await getToken({ req, secret: process.env.NEXTAUTH_SECRET });

  // Unauthenticated → sign in, preserving the intended destination.
  if (!token) {
    const loginUrl = req.nextUrl.clone();
    loginUrl.pathname = `/${locale}/login`;
    loginUrl.searchParams.set("next", req.nextUrl.pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Authenticated but lacking the role → 403 (not a login loop).
  const role = (token.role as string) ?? "";
  for (const guard of roleGuards) {
    if (matches(rest, guard.prefix) && !guard.roles.includes(role)) {
      const forbidden = req.nextUrl.clone();
      forbidden.pathname = `/${locale}/forbidden`;
      forbidden.search = "";
      return NextResponse.redirect(forbidden);
    }
  }

  return res;
}

export const config = {
  // Run on everything except API routes, Next internals, and static files.
  matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
