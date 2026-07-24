import { defineRouting } from "next-intl/routing";

export const locales = ["en", "am"] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = "en";

export const routing = defineRouting({
  locales,
  defaultLocale,
  // Always prefix so the active locale is explicit in the URL (/en, /am).
  localePrefix: "always",
});
