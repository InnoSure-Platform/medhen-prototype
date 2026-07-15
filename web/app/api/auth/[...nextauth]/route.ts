import NextAuth, { NextAuthOptions } from "next-auth";
import KeycloakProvider from "next-auth/providers/keycloak";

const authOptions: NextAuthOptions = {
  providers: [
    KeycloakProvider({
      clientId: process.env.KEYCLOAK_CLIENT || process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT || "pc-web",
      clientSecret: process.env.KEYCLOAK_SECRET || process.env.NEXT_PUBLIC_KEYCLOAK_SECRET || "dummy",
      issuer: (process.env.KEYCLOAK_URL || process.env.NEXT_PUBLIC_KEYCLOAK_URL || "http://localhost:8081") + "/realms/" + (process.env.KEYCLOAK_REALM || process.env.NEXT_PUBLIC_KEYCLOAK_REALM || "medhen"),
    }),
  ],
  callbacks: {
    async jwt({ token, account, user }) {
      if (account) {
        token.accessToken = account.access_token;
        token.idToken = account.id_token;
      }
      return token;
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken as string;
      return session;
    },
  },
  session: {
    strategy: "jwt",
  },
  secret: process.env.NEXTAUTH_SECRET || "supersecret123",
};

const handler = NextAuth(authOptions);

export { handler as GET, handler as POST };
