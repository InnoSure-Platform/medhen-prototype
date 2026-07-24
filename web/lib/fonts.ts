import { Inter, Plus_Jakarta_Sans, Noto_Sans_Ethiopic } from "next/font/google";

// Body / UI text.
export const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
});

// Display / headings.
export const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-jakarta",
});

// Amharic (Ge'ez script) — the Latin faces above have no Ethiopic glyphs.
export const ethiopic = Noto_Sans_Ethiopic({
  subsets: ["ethiopic"],
  weight: ["400", "500", "600", "700"],
  display: "swap",
  variable: "--font-ethiopic",
});

export const fontVariables = `${inter.variable} ${jakarta.variable} ${ethiopic.variable}`;
