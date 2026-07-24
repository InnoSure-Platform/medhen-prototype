import NextAuth, { DefaultSession } from "next-auth";

declare module "next-auth" {
  // The session is serialised to the browser, so it deliberately carries NO
  // access/id token (C5/H8) — only non-secret signals. Tokens live on the JWT
  // (below) and are read server-side via getToken().
  interface Session {
    error?: string;
    role?: string;
    user: {
      id?: string;
      role?: string;
    } & DefaultSession["user"];
  }

  interface User {
    id?: string;
    role?: string;
    accessToken?: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    accessToken?: string;
    idToken?: string;
    error?: string;
    role?: string;
  }
}
