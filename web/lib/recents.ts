"use client";

import * as React from "react";

// The monolith exposes get-by-id endpoints but no per-tenant list endpoints.
// To power dashboards and list views we keep a lightweight client-side record of
// what the user has created this session (localStorage). Detail pages always
// fetch the authoritative record live by id.

export type RecentKind = "policy" | "claim" | "quote";

export interface RecentPolicy {
  id: string;
  policyNumber: string;
  premium: number;
  status: string;
  productCode: string;
  createdAt: string;
}
export interface RecentClaim {
  id: string;
  policyId: string;
  status: string;
  description: string;
  reserve: number;
  createdAt: string;
}

type Shape = { policy: RecentPolicy[]; claim: RecentClaim[]; quote: unknown[] };
const EMPTY: Shape = { policy: [], claim: [], quote: [] };
const KEY = "medhen.recents.v1";
const EVENT = "medhen:recents";

// Cached snapshot so useSyncExternalStore gets a stable reference between writes.
let cache: Shape | null = null;

function load(): Shape {
  if (cache) return cache;
  if (typeof window === "undefined") return EMPTY;
  let next: Shape;
  try {
    next = { ...EMPTY, ...JSON.parse(localStorage.getItem(KEY) || "{}") };
  } catch {
    next = EMPTY;
  }
  cache = next;
  return next;
}

function write(next: Shape) {
  cache = next;
  localStorage.setItem(KEY, JSON.stringify(next));
  window.dispatchEvent(new Event(EVENT));
}

if (typeof window !== "undefined") {
  // Invalidate cache when another tab mutates the store.
  window.addEventListener("storage", (e) => {
    if (e.key === KEY) cache = null;
  });
}

export function addRecentPolicy(p: RecentPolicy) {
  const s = load();
  write({ ...s, policy: [p, ...s.policy.filter((x) => x.id !== p.id)].slice(0, 50) });
}
export function addRecentClaim(c: RecentClaim) {
  const s = load();
  write({ ...s, claim: [c, ...s.claim.filter((x) => x.id !== c.id)].slice(0, 50) });
}

/** Subscribe to a slice of the recents store (SSR-safe, reference-stable). */
export function useRecents<K extends RecentKind>(kind: K): Shape[K] {
  const subscribe = React.useCallback((cb: () => void) => {
    const handler = () => {
      cache = null; // ensure next snapshot reflects latest storage
      cb();
    };
    window.addEventListener(EVENT, cb);
    window.addEventListener("storage", handler);
    return () => {
      window.removeEventListener(EVENT, cb);
      window.removeEventListener("storage", handler);
    };
  }, []);
  const getSnapshot = React.useCallback(() => load()[kind], [kind]);
  const getServerSnapshot = React.useCallback(() => EMPTY[kind], [kind]);
  return React.useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
