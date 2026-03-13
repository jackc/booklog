import { spawn, ChildProcess } from "child_process";
import { writeFileSync } from "fs";
import { fileURLToPath } from "url";
import http from "http";
import path from "path";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const MAPPING_FILE = "/tmp/playwright-servers.json";
const BASE_PORT = 13081;
const STARTUP_TIMEOUT = 10_000;
const POLL_INTERVAL = 100;

interface ServerInfo {
  port: number;
  dbName: string;
  pid: number;
}

function waitForReady(port: number): Promise<void> {
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + STARTUP_TIMEOUT;

    function poll() {
      if (Date.now() > deadline) {
        reject(new Error(`Server on port ${port} not ready within ${STARTUP_TIMEOUT}ms`));
        return;
      }

      const req = http.get(`http://127.0.0.1:${port}/login`, (res) => {
        res.resume();
        resolve();
      });
      req.on("error", () => {
        setTimeout(poll, POLL_INTERVAL);
      });
      req.end();
    }

    poll();
  });
}

async function globalSetup() {
  const workers = parseInt(process.env.PLAYWRIGHT_WORKERS || "4", 10);
  const projectRoot = path.resolve(__dirname, "../..");

  const mapping: Record<number, ServerInfo> = {};
  const processes: ChildProcess[] = [];

  for (let i = 0; i < workers; i++) {
    const port = BASE_PORT + i;
    const dbName = `booklog_test_${i + 1}`;

    const child = spawn(
      path.join(projectRoot, "build", "booklog"),
      [
        "serve",
        "--port", String(port),
        "--frontend-path", path.join(projectRoot, "build", "frontend"),
        "--html-template-path", path.join(projectRoot, "html"),
      ],
      {
        env: {
          ...process.env,
          DATABASE_URL: `database=${dbName}`,
          CSRF_KEY: "test-csrf-key-that-is-at-least-32-characters-long",
          COOKIE_HASH_KEY: "test-cookie-hash-key-at-least-32-characters-long",
          COOKIE_BLOCK_KEY: "test-cookie-block-key-at-least-32-chars-long",
        },
        stdio: "pipe",
      },
    );

    child.stderr?.on("data", (data: Buffer) => {
      process.stderr.write(`[server:${port}] ${data}`);
    });

    processes.push(child);

    mapping[i] = {
      port,
      dbName,
      pid: child.pid!,
    };
  }

  // Wait for all servers to be ready
  await Promise.all(
    Object.values(mapping).map((info) => waitForReady(info.port)),
  );

  writeFileSync(MAPPING_FILE, JSON.stringify(mapping));
}

export default globalSetup;
