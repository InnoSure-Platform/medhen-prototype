import { NextResponse } from "next/server";

const KC = process.env.KEYCLOAK_URL ?? process.env.NEXT_PUBLIC_KEYCLOAK_URL ?? "http://localhost:8081";
const REALM = process.env.KEYCLOAK_REALM ?? process.env.NEXT_PUBLIC_KEYCLOAK_REALM ?? "medhen";
const CLIENT = process.env.KEYCLOAK_CLIENT ?? process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT ?? "pc-web";

export async function POST(req: Request) {
  const { username, password } = (await req.json()) as { username?: string; password?: string };
  if (!username || !password) {
    return NextResponse.json({ message: "username and password required" }, { status: 400 });
  }
  const body = new URLSearchParams({
    grant_type: "password",
    client_id: CLIENT,
    username,
    password,
  });
  const tokenURL = `${KC.replace(/\/$/, "")}/realms/${REALM}/protocol/openid-connect/token`;
  const res = await fetch(tokenURL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
    cache: "no-store",
  });
  const text = await res.text();
  if (!res.ok) {
    return new NextResponse(text, { status: res.status });
  }
  return new NextResponse(text, { status: 200, headers: { "Content-Type": "application/json" } });
}
