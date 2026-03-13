import { Client } from "pg";

export async function connectDB(dbName: string): Promise<Client> {
  const client = new Client({ database: dbName });
  await client.connect();
  return client;
}

export async function resetDB(client: Client): Promise<void> {
  await client.query("SELECT pgundolog.undo()");
}
