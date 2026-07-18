import { withAuth } from "next-auth/middleware";
import { NextResponse } from "next/server";

// Role requirements per protected route prefix. A route not listed here but
// covered by the matcher only requires an authenticated session (any role).
const roleGuards: { prefix: string; roles: string[] }[] = [
  { prefix: "/admin", roles: ["admin"] },
  { prefix: "/broker", roles: ["broker"] },
  // The staff workbench (underwriting, claims settlement, reserves, EOD) is the
  // most privileged surface and must be gated server-side.
  { prefix: "/staff", roles: ["staff", "claims", "admin"] },
];

export default withAuth(
  function middleware(req) {
    const role = (req.nextauth.token?.role as string) ?? "";
    const pathname = req.nextUrl.pathname;

    for (const guard of roleGuards) {
      if (pathname.startsWith(guard.prefix) && !guard.roles.includes(role)) {
        return NextResponse.redirect(new URL("/login", req.url));
      }
    }

    return NextResponse.next();
  },
  {
    callbacks: {
      // No valid session token → not authorized (redirects to signIn page).
      authorized: ({ token }) => !!token,
    },
  }
);

export const config = {
  matcher: [
    "/customer/:path*",
    "/broker/:path*",
    "/admin/:path*",
    "/staff/:path*",
    "/quote/:path*",
    "/claim/:path*",
  ],
};
