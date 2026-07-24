import { getToken } from "next-auth/jwt";
import { NextRequest, NextResponse } from "next/server";

// Server-side federated logout (H8). The id token is read from the encrypted
// NextAuth cookie via getToken() — never passed through the browser or the URL.
// We clear the NextAuth session cookie and redirect to Keycloak's end-session
// endpoint (falling back to the home page when no id token is present).
const SESSION_COOKIES = [
  "next-auth.session-token",
  "__Secure-next-auth.session-token",
];

export async function GET(req: NextRequest) {
  const baseUrl = new URL("/", req.url).toString();
  const token = await getToken({ req, secret: process.env.NEXTAUTH_SECRET });
  const idToken = token?.idToken as string | undefined;

  const target =
    idToken && process.env.KEYCLOAK_ISSUER
      ? `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/logout` +
        `?id_token_hint=${encodeURIComponent(idToken)}` +
        `&post_logout_redirect_uri=${encodeURIComponent(baseUrl)}`
      : baseUrl;

  const res = NextResponse.redirect(target);
  for (const name of SESSION_COOKIES) {
    res.cookies.set(name, "", { maxAge: 0, path: "/" });
  }
  return res;
}
