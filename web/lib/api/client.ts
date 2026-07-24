import createClient, { type Middleware } from "openapi-fetch";
import type { paths } from "./schema";

// All calls go same-origin to the Next.js proxy (app/api/medhen/[...path]), which
// attaches the access token + tenant server-side. The browser holds no token and
// never learns the upstream origin. Types come from the OpenAPI contract.
const PROXY_BASE = "/api/medhen";

/**
 * ApiError carries only the HTTP status and an i18n message *key* (into the
 * `errors.*` namespace). Raw backend bodies are never surfaced to users (M8);
 * components translate `error.code` with next-intl.
 */
export class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number) {
    const code = errorCode(status);
    super(code);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

const KNOWN = new Set([400, 401, 403, 404, 409, 422]);

/** Map an HTTP status to an i18n message key in the `errors.*` namespace. */
export function errorCode(status: number): string {
  if (KNOWN.has(status)) return String(status);
  if (status >= 500) return "500";
  return "generic";
}

// Adds a per-request idempotency key to unsafe methods (backend dedupes writes).
const idempotency: Middleware = {
  async onRequest({ request }) {
    if (!["GET", "HEAD"].includes(request.method) && !request.headers.get("Idempotency-Key")) {
      request.headers.set("Idempotency-Key", crypto.randomUUID());
    }
    return request;
  },
};

export const rawClient = createClient<paths>({ baseUrl: PROXY_BASE });
rawClient.use(idempotency);

/** Throw a typed ApiError from an openapi-fetch result, or return its data. */
export function unwrap<T>(result: { data?: T; error?: unknown; response: Response }): T {
  if (result.error !== undefined || !result.response.ok) {
    throw new ApiError(result.response.status);
  }
  return result.data as T;
}

/** Resolve a user-safe message from any thrown error using a translator. */
export function errorMessage(e: unknown, t: (key: string) => string): string {
  if (e instanceof ApiError) return t(e.code);
  return t("generic");
}

export interface RequestOptions {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  locale?: string;
}

/**
 * Same-origin request to the proxy with typed response. Adds an Idempotency-Key
 * to writes and Accept-Language when a locale is given. Never surfaces raw
 * bodies — non-2xx throws a typed ApiError (M8).
 */
export async function apiRequest<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const method = opts.method ?? "GET";
  const headers: Record<string, string> = {};
  if (opts.locale) headers["Accept-Language"] = opts.locale;
  if (method !== "GET") {
    headers["Content-Type"] = "application/json";
    headers["Idempotency-Key"] = crypto.randomUUID();
  }

  let res: Response;
  try {
    res = await fetch(`${PROXY_BASE}${path}`, {
      method,
      headers,
      body: opts.body != null ? JSON.stringify(opts.body) : undefined,
      cache: "no-store",
    });
  } catch {
    throw new ApiError(502);
  }

  if (!res.ok) {
    await res.text().catch(() => "");
    // Signal an expired/invalid session so the app can re-authenticate while
    // preserving the current location (handled in components/providers).
    if (res.status === 401 && typeof window !== "undefined") {
      window.dispatchEvent(new Event("medhen:unauthorized"));
    }
    throw new ApiError(res.status);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}
