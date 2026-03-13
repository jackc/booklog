import { Client } from "pg";
import bcrypt from "bcryptjs";

export async function createUser(
  client: Client,
  attrs: Record<string, unknown>,
): Promise<Record<string, unknown>> {
  const password = (attrs.password as string) || "password";
  const fields = { ...attrs };
  delete fields.password;

  const passwordDigest = bcrypt.hashSync(password, 4);
  fields.password_digest = passwordDigest;

  if (!fields.username) {
    fields.username = "test";
  }

  const columns = Object.keys(fields);
  const values = Object.values(fields);
  const placeholders = columns.map((_, i) => `$${i + 1}`);

  const sql = `INSERT INTO users (${columns.join(", ")}) VALUES (${placeholders.join(", ")}) RETURNING *`;
  const result = await client.query(sql, values);
  return result.rows[0];
}

export async function createBook(
  client: Client,
  attrs: Record<string, unknown>,
): Promise<Record<string, unknown>> {
  const columns = Object.keys(attrs);
  const values = Object.values(attrs);
  const placeholders = columns.map((_, i) => `$${i + 1}`);

  const sql = `INSERT INTO books (${columns.join(", ")}) VALUES (${placeholders.join(", ")}) RETURNING *`;
  const result = await client.query(sql, values);
  return result.rows[0];
}
