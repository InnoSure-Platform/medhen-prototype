import NextAuth, { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

const authOptions: NextAuthOptions = {
  providers: [
    CredentialsProvider({
      name: "Custom Backend",
      credentials: {
        email: { label: "Email", type: "email", placeholder: "you@example.com" },
        password: { label: "Password", type: "password" },
        portal: { label: "Portal", type: "text" }, // To distinguish customer, broker, admin logins
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          return null;
        }

        try {
          // Replace with actual API endpoint to Go backend
          const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
          const res = await fetch(`${apiUrl}/api/v1/auth/login`, {
            method: 'POST',
            body: JSON.stringify(credentials),
            headers: { "Content-Type": "application/json" }
          });
          
          const data = await res.json();
          
          // Return user object if successful, else null
          if (res.ok && data.token) {
            return {
              id: data.user.id || data.user.email,
              name: data.user.name,
              email: data.user.email,
              role: data.user.role,
              accessToken: data.token,
            };
          }
          return null;
        } catch (e) {
          console.error("Auth error:", e);
          return null;
        }
      }
    }),
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.accessToken = user.accessToken;
        token.role = user.role;
      }
      return token;
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken as string;
      session.role = token.role as string;
      return session;
    },
  },
  session: {
    strategy: "jwt",
  },
  pages: {
    signIn: "/login",
  },
  secret: process.env.NEXTAUTH_SECRET || "supersecret123",
};

const handler = NextAuth(authOptions);

export { handler as GET, handler as POST };

