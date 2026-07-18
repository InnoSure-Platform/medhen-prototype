import { NextResponse } from "next/server";

export async function GET(req: Request) {
  const url = new URL(req.url);
  const idToken = url.searchParams.get("idToken");
  
  const baseUrl = new URL("/", req.url).toString();

  if (!idToken) {
    return NextResponse.redirect(baseUrl);
  }

  // Construct Keycloak end session endpoint URL
  const endSessionURL = `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/logout?id_token_hint=${idToken}&post_logout_redirect_uri=${encodeURIComponent(baseUrl)}`;
  
  return NextResponse.redirect(endSessionURL);
}
