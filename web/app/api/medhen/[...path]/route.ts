import { getToken } from "next-auth/jwt";
import { NextRequest, NextResponse } from "next/server";

// Server-side proxy to the Medhen modular monolith (C5). The browser never sees
// the access token: it calls this same-origin route, we read the token from the
// encrypted NextAuth session cookie via getToken(), and attach Authorization +
// the tenant server-side before forwarding to the monolith. The upstream base
// URL and tenant are server-only env (NOT NEXT_PUBLIC_*), so neither the token
// nor the API origin leak to the client bundle.
const MEDHEN_API = (process.env.MEDHEN_API ?? "http://localhost:8080").replace(/\/$/, "");
const TENANT = process.env.MEDHEN_TENANT ?? "eic";

// Only these client-supplied headers are forwarded upstream; everything else
// (cookies, host, forged auth/tenant) is dropped and set server-side.
const FORWARD_HEADERS = ["content-type", "accept-language", "idempotency-key"];

async function proxy(req: NextRequest, path: string[]): Promise<NextResponse> {
  const token = await getToken({ req, secret: process.env.NEXTAUTH_SECRET });
  if (!token?.accessToken) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const upstreamPath = "/" + path.map(encodeURIComponent).join("/");
  const search = req.nextUrl.search;
  const url = `${MEDHEN_API}${upstreamPath}${search}`;

  const headers = new Headers();
  for (const name of FORWARD_HEADERS) {
    const v = req.headers.get(name);
    if (v) headers.set(name, v);
  }
  headers.set("Authorization", `Bearer ${token.accessToken}`);
  headers.set("X-Tenant-ID", TENANT);

  const method = req.method.toUpperCase();
  const hasBody = method !== "GET" && method !== "HEAD";
  const body = hasBody ? await req.text() : undefined;

  let upstream: Response;
  try {
    upstream = await fetch(url, { method, headers, body, cache: "no-store" });
  } catch {
    return NextResponse.json({ error: "upstream unreachable" }, { status: 502 });
  }

  // Stream the upstream body back verbatim; the client maps status → a
  // user-safe message and never renders raw error bodies (M8).
  const respBody = await upstream.text();
  return new NextResponse(respBody, {
    status: upstream.status,
    headers: {
      "Content-Type": upstream.headers.get("content-type") ?? "application/json",
    },
  });
}

type Ctx = { params: Promise<{ path: string[] }> };

async function handle(req: NextRequest, ctx: Ctx) {
  const { path } = await ctx.params;
  return proxy(req, path);
}

export const GET = handle;
export const POST = handle;
export const PUT = handle;
export const PATCH = handle;
export const DELETE = handle;
