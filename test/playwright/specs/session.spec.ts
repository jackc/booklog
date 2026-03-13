import { test, expect } from "../helpers/fixtures";
import { createUser } from "../helpers/factories";
import { login } from "../helpers/login";

test("user logs in and logs out", async ({ page, serverURL, db }) => {
  await createUser(db, { username: "john", password: "mysecret" });
  await login(page, serverURL, "john", "mysecret");

  await expect(page.locator("body")).toContainText("Per Year");

  await page.getByRole("button", { name: "Logout" }).click();

  await expect(page.getByText("Username")).toBeVisible();
  await expect(page.getByText("Password")).toBeVisible();
});

test("user with invalid session is logged out", async ({ page, serverURL, db }) => {
  await createUser(db, { username: "john", password: "mysecret" });
  await login(page, serverURL, "john", "mysecret");

  await expect(page.locator("body")).toContainText("Per Year");

  await db.query("DELETE FROM user_sessions");

  await page.getByRole("link", { name: "New Book" }).click();

  await expect(page.getByText("Username")).toBeVisible();
  await expect(page.getByText("Password")).toBeVisible();
});
