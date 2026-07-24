import { describe, expect, it } from "vitest";
import { formatBirr, formatETB, formatPct } from "./format";

describe("money & number formatting", () => {
  it("formats major-unit Birr with 2 decimals + ETB suffix", () => {
    expect(formatBirr(2680, "en")).toBe("2,680.00 ETB");
  });

  it("converts minor units (santim) to Birr", () => {
    // 268000 santim = 2,680.00 Birr
    expect(formatETB(268000, "en")).toBe("2,680.00 ETB");
  });

  it("formats ratios as percentages", () => {
    expect(formatPct(0.72, "en")).toBe("72.0%");
    expect(formatPct(1.05, "en", 1)).toBe("105.0%");
  });

  it("handles nullish amounts safely", () => {
    expect(formatBirr(undefined as unknown as number, "en")).toBe("0.00 ETB");
  });
});
