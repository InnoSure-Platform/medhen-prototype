import { initSentry } from "@/lib/sentry";

// Browser error/trace capture. initSentry lazy-imports the SDK only when a DSN
// is set, so nothing ships to the browser bundle in the no-DSN case.
void initSentry(process.env.NEXT_PUBLIC_SENTRY_DSN);
