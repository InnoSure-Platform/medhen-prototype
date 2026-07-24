import { getRequestConfig } from "next-intl/server";
import { routing, type Locale } from "./routing";

// Loads the active locale's messages for Server Components (next-intl).
export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale =
    requested && routing.locales.includes(requested as Locale)
      ? requested
      : routing.defaultLocale;

  return {
    locale,
    messages: (await import(`../../messages/${locale}.json`)).default,
  };
});
