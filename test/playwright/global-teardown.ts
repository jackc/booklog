import { readFileSync, unlinkSync } from "fs";

const MAPPING_FILE = "/tmp/playwright-servers.json";

async function globalTeardown() {
  let mapping: Record<string, { pid: number }>;
  try {
    mapping = JSON.parse(readFileSync(MAPPING_FILE, "utf-8"));
  } catch {
    return;
  }

  for (const info of Object.values(mapping)) {
    try {
      process.kill(info.pid, "SIGTERM");
    } catch {
      // Process may already be gone
    }
  }

  try {
    unlinkSync(MAPPING_FILE);
  } catch {
    // Ignore
  }
}

export default globalTeardown;
