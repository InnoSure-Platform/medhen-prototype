/* eslint-disable @typescript-eslint/no-explicit-any -- NextAuth/Keycloak profile
   and token payloads are loosely typed at the provider boundary; this file is a
   verbatim port of the security-reviewed auth glue. */
import NextAuth, { NextAuthOptions } from "next-auth";
import KeycloakProvider from "next-auth/providers/keycloak";

// Fail closed at RUNTIME (not build time). We never fall back to a hardcoded,
// known value: a known NEXTAUTH_SECRET would let anyone forge a session
// (including role: "admin"). NextAuth itself refuses to run in production when
// NEXTAUTH_SECRET is unset, and an unset KEYCLOAK_SECRET simply fails the OAuth
// exchange — neither is a forgeable constant. These must be configured in the
// deployment environment (see web/.env.local.example). Reading them here (rather
// than throwing at import) keeps `next build` working when env is injected only
// at runtime.
const KEYCLOAK_SECRET = process.env.KEYCLOAK_SECRET ?? "";
const NEXTAUTH_SECRET = process.env.NEXTAUTH_SECRET;

async function refreshAccessToken(token: any) {
  try {
    const url = `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/token`;

    const response = await fetch(url, {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      method: "POST",
      body: new URLSearchParams({
        client_id: process.env.KEYCLOAK_CLIENT || "pc-web",
        client_secret: KEYCLOAK_SECRET,
        grant_type: "refresh_token",
        refresh_token: token.refreshToken,
      }),
    });

    const refreshedTokens = await response.json();

    if (!response.ok) {
      throw refreshedTokens;
    }

    return {
      ...token,
      accessToken: refreshedTokens.access_token,
      accessTokenExpires: Date.now() + refreshedTokens.expires_in * 1000,
      refreshToken: refreshedTokens.refresh_token ?? token.refreshToken,
    };
  } catch (error) {
    console.error("Error refreshing access token", error);
    return {
      ...token,
      error: "RefreshAccessTokenError",
    };
  }
}

const authOptions: NextAuthOptions = {
  providers: [
    KeycloakProvider({
      clientId: process.env.KEYCLOAK_CLIENT || "pc-web",
      clientSecret: KEYCLOAK_SECRET,
      issuer: process.env.KEYCLOAK_ISSUER,
      profile(profile) {
        return {
          id: profile.sub,
          name: profile.name ?? profile.preferred_username,
          email: profile.email,
          image: profile.picture,
          roles: profile.realm_access?.roles || [],
        };
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user, account }) {
      if (account && user) {
        const roles = (user as any).roles as string[] || [];
        let primaryRole = "customer";
        
        if (roles.includes("admin")) primaryRole = "admin";
        else if (roles.includes("agent") || roles.includes("broker")) primaryRole = "broker";
        else if (roles.includes("claims")) primaryRole = "claims";
        else if (roles.includes("staff")) primaryRole = "staff";

        return {
          accessToken: account.access_token,
          accessTokenExpires: Date.now() + (account.expires_in as number) * 1000,
          refreshToken: account.refresh_token,
          idToken: account.id_token,
          role: primaryRole,
          user,
        };
      }

      if (Date.now() < (token.accessTokenExpires as number)) {
        return token;
      }

      const refreshedToken = await refreshAccessToken(token);
      refreshedToken.role = token.role;
      refreshedToken.idToken = token.idToken;
      return refreshedToken;
    },
    async session({ session, token }) {
      if (token) {
        session.user = token.user as any;
        // Access/id tokens are deliberately NOT copied onto the session (C5/H8):
        // the session is serialised to the browser. Server code that needs the
        // token reads it from the encrypted JWT cookie via getToken() instead
        // (see app/api/medhen/[...path]/route.ts and federated-logout). Only the
        // non-secret role/error signals are exposed to the client.
        session.error = token.error as string;
        session.role = token.role as string;
      }
      return session;
    },
  },
  session: { strategy: "jwt" },
  secret: NEXTAUTH_SECRET,
  pages: {
    // Locale-prefixed login route (see app/[locale]/(marketing)/login). The
    // login page itself offers a language toggle.
    signIn: "/en/login",
  },
};

const handler = NextAuth(authOptions);

export { handler as GET, handler as POST };
