import { spawnSync } from "node:child_process";
import path from "node:path";
import fs from "node:fs";
import os from "node:os";
import crypto from "node:crypto";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PROJECT_ROOT = path.resolve(__dirname, "..", "..", "..");
const BIN_PATH = path.join(PROJECT_ROOT, "bin", "jsoon");

export interface JSoonOpts {
  upcomingDays?: number;
  limit?: number;
  template?: string;
  offsetMarkers?: Record<string, string>;
}

export function buildBinary() {
  console.log("[JSOON Wrapper] Building binary...");
  spawnSync("make", ["build"], { cwd: PROJECT_ROOT });
}

export function runJSoon(input: string | string[], options: JSoonOpts = {}) {
  if (!fs.existsSync(BIN_PATH)) {
    buildBinary();
  }

  // Basic server-side validation
  const upcomingDays = Math.max(0, options.upcomingDays || 7);
  const limit = Math.max(0, options.limit || 0);
  const template = (options.template || "").trim();
  const offsetMarkers = options.offsetMarkers || {};

  const args: string[] = ["-f", "stdout"];
  args.push("-u", upcomingDays.toString());
  args.push("-l", limit.toString());
  if (template) args.push("-t", template);

  const configPath = path.join(
    os.tmpdir(),
    `jsoon-config-${crypto.randomUUID()}.json`,
  );

  try {
    if (Array.isArray(input)) {
      // URL Mode: We must ignore stdin to prevent the binary from entering Stdin Mode
      const config = {
        calendars: input,
        upcoming_days: upcomingDays,
        events_limit: limit,
        date_template: template,
        offset_markers: offsetMarkers,
      };

      fs.writeFileSync(configPath, JSON.stringify(config));

      const result = spawnSync(BIN_PATH, [...args, "-c", configPath], {
        cwd: PROJECT_ROOT,
        stdio: ["ignore", "pipe", "pipe"], // 'ignore' sets stdin to /dev/null
      });

      return handleResult(result);
    } else {
      // Stdin Mode: We provide the input string to stdin
      // If we have offsetMarkers, we STILL need to provide a config file
      // because there's no CLI flag for markers yet.
      if (Object.keys(offsetMarkers).length > 0) {
        const config = {
          upcoming_days: upcomingDays,
          events_limit: limit,
          date_template: template,
          offset_markers: offsetMarkers,
        };
        fs.writeFileSync(configPath, JSON.stringify(config));
        args.push("-c", configPath);
      }

      const result = spawnSync(BIN_PATH, args, {
        cwd: PROJECT_ROOT,
        input: input,
        stdio: ["pipe", "pipe", "pipe"],
      });

      return handleResult(result);
    }
  } finally {
    if (fs.existsSync(configPath)) fs.unlinkSync(configPath);
  }
}

function handleResult(result: any) {
  if (result.status !== 0) {
    const err =
      result.stderr?.toString() ||
      result.error?.message ||
      "Binary execution failed";
    console.error(`[JSOON Wrapper] Error: ${err}`);
    throw new Error(err);
  }

  const output = result.stdout.toString().trim();
  if (!output || output === "null") return [];

  try {
    // Extract JSON array in case of surrounding logs
    const start = output.indexOf("[");
    const end = output.lastIndexOf("]");
    if (start !== -1 && end !== -1 && end > start) {
      return JSON.parse(output.substring(start, end + 1));
    }
    return JSON.parse(output);
  } catch (e: any) {
    console.error(`[JSOON Wrapper] Parse Error. Output: ${output}`);
    throw new Error(`Failed to parse binary output: ${e.message}`);
  }
}
