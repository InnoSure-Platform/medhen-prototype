import { initSentry } from "@/lib/sentry";

// Server + edge runtime error/trace capture (no-op without SENTRY_DSN).
export function register() {
  void initSentry(process.env.SENTRY_DSN);
}

export async function onRequestError(...args: unknown[]) {
  const Sentry = await import("@sentry/nextjs");
  // @ts-expect-error - forward Next's error hook to Sentry when available.
  Sentry.captureRequestError?.(...args);
}
