import { NextRequest, NextResponse } from "next/server";
import { getToken } from "next-auth/jwt";

// Federated logout. The OIDC id_token is read from the server-side NextAuth
// session (never from the URL — H8: a token in the query string is logged,
// cached, and leaked via Referer). If no session is present we just go home.
export async function GET(req: NextRequest) {
  const baseUrl = new URL("/", req.url).toString();

  const token = await getToken({ req, secret: process.env.NEXTAUTH_SECRET });
  const idToken = token?.idToken as string | undefined;
  if (!idToken) {
    return NextResponse.redirect(baseUrl);
  }

  const endSessionURL =
    `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/logout` +
    `?id_token_hint=${encodeURIComponent(idToken)}` +
    `&post_logout_redirect_uri=${encodeURIComponent(baseUrl)}`;

  return NextResponse.redirect(endSessionURL);
}
