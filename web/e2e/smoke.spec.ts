import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test("root redirects to a locale", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/(en|am)(\/)?$/);
});

test("marketing landing renders the hero and CTAs", async ({ page }) => {
  await page.goto("/en");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  await expect(page.getByRole("link", { name: /start a motor quote/i })).toBeVisible();
});

test("landing has no serious accessibility violations", async ({ page }) => {
  await page.goto("/en");
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
  const serious = results.violations.filter((v) => v.impact === "serious" || v.impact === "critical");
  expect(serious).toEqual([]);
});

test("login page offers sign in and self-registration", async ({ page }) => {
  await page.goto("/en/login");
  await expect(page.getByRole("button", { name: /continue with keycloak/i })).toBeVisible();
  await expect(page.getByRole("button", { name: /create an account/i })).toBeVisible();
});

test("switching to Amharic renders Geʼez content", async ({ page }) => {
  await page.goto("/am");
  await expect(page.locator("html")).toHaveAttribute("lang", "am");
});
