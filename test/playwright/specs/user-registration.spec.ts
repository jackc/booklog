import { test, expect } from "../helpers/fixtures";
import { createUser } from "../helpers/factories";

test("user registration", async ({ page, serverURL, db }) => {
  await createUser(db, { username: "john", password: "mysecret" });

  await page.goto(`${serverURL}/user_registration/new`);
  await page.getByLabel("Username").fill("test");
  await page.getByLabel("Password").fill("secret phrase");
  await page.getByRole("button", { name: "Sign up" }).click();
  await expect(page.getByRole("link", { name: "New Book" })).toBeVisible();
});
