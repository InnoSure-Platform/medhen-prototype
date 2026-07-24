import { describe, expect, it } from "vitest";
import { hasPermission, homeForRole } from "./permissions";

describe("permissions", () => {
  it("grants everything to admin", () => {
    expect(hasPermission("admin", "users:manage")).toBe(true);
    expect(hasPermission("admin", "claim:settle")).toBe(true);
  });

  it("gives claims officers settlement authority but not customers", () => {
    expect(hasPermission("claims", "claim:settle")).toBe(true);
    expect(hasPermission("staff", "claim:settle")).toBe(false);
    expect(hasPermission("customer", "claim:settle")).toBe(false);
  });

  it("lets customers create quotes and pay invoices", () => {
    expect(hasPermission("customer", "quote:create")).toBe(true);
    expect(hasPermission("customer", "invoice:pay")).toBe(true);
    expect(hasPermission("customer", "users:manage")).toBe(false);
  });

  it("denies everything to an empty role", () => {
    expect(hasPermission("", "policy:read")).toBe(false);
  });

  it("routes each role to its portal home", () => {
    expect(homeForRole("admin")).toBe("/admin");
    expect(homeForRole("staff")).toBe("/staff");
    expect(homeForRole("claims")).toBe("/staff");
    expect(homeForRole("broker")).toBe("/broker");
    expect(homeForRole("customer")).toBe("/customer");
    expect(homeForRole("")).toBe("/customer");
  });
});
