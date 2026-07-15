import type { Metadata } from "next";
import { DM_Sans, Fraunces, Noto_Sans_Ethiopic } from "next/font/google";
import { AuthProvider } from "@/components/AuthProvider";
import { Shell } from "@/components/Shell";
import "./globals.css";

const display = Fraunces({ subsets: ["latin"], variable: "--font-display", weight: ["500", "600", "700"] });
const body = DM_Sans({ subsets: ["latin"], variable: "--font-body", weight: ["400", "500", "600", "700"] });
const ethiopic = Noto_Sans_Ethiopic({ subsets: ["ethiopic"], variable: "--font-ethiopic", weight: ["400", "600", "700"] });

export const metadata: Metadata = {
  title: "Medhen · Ethiopian Insurance Corporation",
  description: "Phase 0 Motor insurance automation prototype for EIC",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${display.variable} ${body.variable} ${ethiopic.variable}`}>
        <AuthProvider>
          <Shell>{children}</Shell>
        </AuthProvider>
      </body>
    </html>
  );
}
