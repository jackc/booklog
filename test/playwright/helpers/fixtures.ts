import { test as base, expect } from "@playwright/test";
import { Client } from "pg";
import { readFileSync } from "fs";
import { connectDB, resetDB } from "./db";

const MAPPING_FILE = "/tmp/playwright-servers.json";

interface ServerInfo {
  port: number;
  dbName: string;
  pid: number;
}

function getServerInfo(workerIndex: number): ServerInfo {
  const mapping: Record<string, ServerInfo> = JSON.parse(
    readFileSync(MAPPING_FILE, "utf-8"),
  );
  return mapping[workerIndex];
}

type TestFixtures = {
  resetDB: void;
};

type WorkerFixtures = {
  serverURL: string;
  db: Client;
};

export const test = base.extend<TestFixtures, WorkerFixtures>({
  serverURL: [
    async ({}, use, workerInfo) => {
      const info = getServerInfo(workerInfo.workerIndex);
      await use(`http://127.0.0.1:${info.port}`);
    },
    { scope: "worker" },
  ],

  db: [
    async ({}, use, workerInfo) => {
      const info = getServerInfo(workerInfo.workerIndex);
      const client = await connectDB(info.dbName);
      await use(client);
      await client.end();
    },
    { scope: "worker" },
  ],

  resetDB: [
    async ({ db }, use) => {
      await resetDB(db);
      await use();
    },
    { auto: true },
  ],
});

export { expect };
