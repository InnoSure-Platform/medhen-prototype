// Lazy Sentry wrapper. @sentry/nextjs is only imported at runtime when a DSN is
// configured, so it never bloats the browser bundle in the common (no-DSN) case.

let inited = false;

export async function initSentry(dsn: string | undefined) {
  if (!dsn || inited) return;
  inited = true;
  const Sentry = await import("@sentry/nextjs");
  Sentry.init({
    dsn,
    environment: process.env.NEXT_PUBLIC_MEDHEN_ENV ?? process.env.NODE_ENV,
    tracesSampleRate: 0.1,
    sendDefaultPii: false, // the app never sends tokens to the browser anyway
  });
}

/** Report an exception if Sentry is configured (no-op otherwise). */
export async function captureException(error: unknown) {
  const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN ?? process.env.SENTRY_DSN;
  if (!dsn) return;
  const Sentry = await import("@sentry/nextjs");
  Sentry.captureException(error);
}
