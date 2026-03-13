import { Page, expect } from "@playwright/test";

export async function login(
  page: Page,
  serverURL: string,
  username: string,
  password: string,
): Promise<void> {
  await page.goto(`${serverURL}/login`);
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Login" }).click();
  await expect(page.getByRole("link", { name: "New Book" })).toBeVisible();
}
