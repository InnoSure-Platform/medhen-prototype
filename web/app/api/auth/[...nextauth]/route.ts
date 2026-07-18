import NextAuth, { NextAuthOptions } from "next-auth";
import KeycloakProvider from "next-auth/providers/keycloak";

async function refreshAccessToken(token: any) {
  try {
    const url = `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/token`;

    const response = await fetch(url, {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      method: "POST",
      body: new URLSearchParams({
        client_id: process.env.KEYCLOAK_CLIENT || "pc-web",
        client_secret: process.env.KEYCLOAK_SECRET || "",
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
      clientSecret: process.env.KEYCLOAK_SECRET || "",
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
        session.accessToken = token.accessToken as string;
        session.idToken = token.idToken as string;
        session.error = token.error as string;
        session.role = token.role as string;
      }
      return session;
    },
  },
  session: { strategy: "jwt" },
  secret: process.env.NEXTAUTH_SECRET || "supersecret123",
  pages: {
    signIn: "/login",
  },
};

const handler = NextAuth(authOptions);

export { handler as GET, handler as POST };
