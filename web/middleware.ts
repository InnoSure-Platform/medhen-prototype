import { withAuth } from "next-auth/middleware";
import { NextResponse } from "next/server";

export default withAuth(
  function middleware(req) {
    const token = req.nextauth.token;
    const pathname = req.nextUrl.pathname;

    // Simple role-based routing (example)
    if (pathname.startsWith("/admin") && token?.role !== "admin") {
      // You could redirect to a 403 or the portal they belong to
      return NextResponse.redirect(new URL("/login", req.url));
    }
    
    if (pathname.startsWith("/broker") && token?.role !== "broker") {
      return NextResponse.redirect(new URL("/login", req.url));
    }

    return NextResponse.next();
  },
  {
    callbacks: {
      authorized: ({ token }) => !!token,
    },
  }
);

export const config = {
  matcher: [
    "/customer/:path*",
    "/broker/:path*",
    "/admin/:path*",
  ],
};
