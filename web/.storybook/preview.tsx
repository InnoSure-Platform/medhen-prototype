import * as React from "react";
import type { Preview } from "@storybook/react";
import { NextIntlClientProvider } from "next-intl";
import { withThemeByClassName } from "@storybook/addon-themes";
import en from "../messages/en.json";
import am from "../messages/am.json";
import "../app/globals.css";

const messages: Record<string, typeof en> = { en, am: am as typeof en };

const preview: Preview = {
  parameters: {
    layout: "centered",
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
    a11y: { test: "error" },
    options: {
      storySort: {
        order: ["Foundations", ["Introduction", "Colors", "Typography", "Spacing", "Elevation", "Motion"], "Primitives", "Data Display", "Overlays", "Feedback", "Charts", "Shell", "Patterns"],
      },
    },
  },
  globalTypes: {
    locale: {
      description: "Locale",
      toolbar: {
        icon: "globe",
        items: [
          { value: "en", title: "English" },
          { value: "am", title: "አማርኛ" },
        ],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: { locale: "en" },
  decorators: [
    (Story, ctx) => {
      const locale = (ctx.globals.locale as "en" | "am") ?? "en";
      return (
        <div
          lang={locale}
          className="font-sans bg-canvas text-fg"
          style={{ padding: "2rem", minWidth: "min(100%, 24rem)" }}
        >
          <NextIntlClientProvider locale={locale} messages={messages[locale]}>
            <Story />
          </NextIntlClientProvider>
        </div>
      );
    },
    withThemeByClassName({
      themes: { light: "", dark: "dark" },
      defaultTheme: "light",
    }),
  ],
};

export default preview;
