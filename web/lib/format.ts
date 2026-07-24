import type { Locale } from "@/lib/i18n/routing";

const intlLocale = (l: Locale) => (l === "am" ? "am-ET" : "en-ET");

/**
 * Format an amount already in MAJOR units (Birr). The monolith's domain money
 * marshals as a major-unit JSON number (e.g. 2160.00 for reads).
 */
export function formatBirr(birr: number, locale: Locale, opts?: { compact?: boolean }): string {
  const value = (birr ?? 0).toLocaleString(intlLocale(locale), {
    minimumFractionDigits: opts?.compact ? 0 : 2,
    maximumFractionDigits: opts?.compact ? 1 : 2,
    notation: opts?.compact ? "compact" : "standard",
  });
  return `${value} ETB`;
}

/** Format an amount given in MINOR units (santim), e.g. KPI figures. */
export function formatETB(minor: number, locale: Locale, opts?: { compact?: boolean }): string {
  return formatBirr((minor ?? 0) / 100, locale, opts);
}

/** Percentage from a ratio (0.72 → "72.0%"). */
export function formatPct(ratio: number, locale: Locale, digits = 1): string {
  return `${((ratio ?? 0) * 100).toLocaleString(intlLocale(locale), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })}%`;
}

/** Absolute, locale-aware date. */
export function formatDate(iso: string, locale: Locale): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString(intlLocale(locale), { year: "numeric", month: "short", day: "numeric" });
}

/** Compact relative time ("just now", "3m", "2h", or a date). */
export function relativeTime(iso: string, locale: Locale): string {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return "";
  const s = Math.max(0, Math.round((Date.now() - t) / 1000));
  const am = locale === "am";
  if (s < 60) return am ? "አሁን" : "just now";
  const m = Math.round(s / 60);
  if (m < 60) return am ? `ከ${m} ደቂቃ` : `${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24) return am ? `ከ${h} ሰዓት` : `${h}h ago`;
  return formatDate(iso, locale);
}
