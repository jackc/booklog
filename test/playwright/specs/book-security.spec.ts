import { test, expect } from "../helpers/fixtures";
import { createUser, createBook } from "../helpers/factories";
import { login } from "../helpers/login";

test("anonymous user is redirected to login", async ({ page, serverURL, db }) => {
  const user = await createUser(db, { username: "test", password: "secret phrase" });
  await createBook(db, {
    user_id: user.id,
    title: "Foo",
    author: "Bar",
    finish_date: "2019-01-01",
    format: "text",
  });

  await page.goto(`${serverURL}/users/test/books`);
  await expect(page.locator("form")).toContainText("Login");
  expect(page.url()).toBe(`${serverURL}/login`);
});

test("other user sees forbidden", async ({ page, serverURL, db }) => {
  const user = await createUser(db, { username: "test", password: "secret phrase" });
  await createBook(db, {
    user_id: user.id,
    title: "Foo",
    author: "Bar",
    finish_date: "2019-01-01",
    format: "text",
  });

  await createUser(db, { username: "other", password: "secret phrase" });
  await login(page, serverURL, "other", "secret phrase");
  await page.goto(`${serverURL}/users/test/books`);
  await expect(page.locator("body")).toContainText("Forbidden");
});
