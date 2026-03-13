import { test, expect } from "../helpers/fixtures";
import { createUser } from "../helpers/factories";
import { login } from "../helpers/login";

test("book CRUD cycle", async ({ page, serverURL, db }) => {
  await createUser(db, { username: "john", password: "mysecret" });
  await login(page, serverURL, "john", "mysecret");

  await expect(page.getByRole("link", { name: "New Book" })).toBeVisible();

  await page.getByRole("link", { name: "New Book" }).click();
  await page.getByLabel("Title").fill("Paradise Lost");
  await page.getByLabel("Author").fill("Paradise Lost");
  await page.getByLabel("Finish Date").fill("2019-01-01");
  await page.getByLabel("Format").selectOption("audio");
  await page.getByRole("button", { name: "Save" }).click();

  await expect(page.locator("dd").first()).toContainText("Paradise Lost");

  // Verify in database
  const result = await db.query("SELECT * FROM books");
  expect(result.rows).toHaveLength(1);
  expect(result.rows[0].title).toBe("Paradise Lost");

  await page.getByRole("link", { name: "Edit" }).click();
  await page.getByLabel("Title").fill("Paradise Regained");
  await page.getByRole("button", { name: "Save" }).click();

  await expect(page.locator("dd").first()).toContainText("Paradise Regained");

  await page.getByRole("link", { name: "Delete" }).click();
  await page.getByRole("button", { name: "Delete" }).click();

  await expect(page.getByRole("link", { name: "New Book" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Paradise Regained" })).not.toBeVisible();
});
