#! /usr/bin/env node

const { execSync } = require("child_process");
const path = require("path");

const BINARY = path.join(__dirname, "..", "bin", "jfi");
const CONFIG = path.join(__dirname, "test_calendars.conf");

function parseICalDate(str) {
  if (!str) return null;
  const year = parseInt(str.substring(0, 4));
  const month = parseInt(str.substring(4, 6)) - 1;
  const day = parseInt(str.substring(6, 8));

  let hour = 0,
    min = 0,
    sec = 0;
  if (str.includes("T")) {
    hour = parseInt(str.substring(9, 11));
    min = parseInt(str.substring(11, 13));
    sec = parseInt(str.substring(13, 15));
  }

  // Treat as local time to match program logic
  return new Date(year, month, day, hour, min, sec);
}

function getTodayBoundaries() {
  const now = new Date();
  const start = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate(),
    0,
    0,
    0,
    0,
  );
  const end = new Date(start);
  end.setHours(23, 59, 59, 999);
  return { start, end };
}

function runTest(upcomingDays) {
  console.log(`\n--- Testing Lookahead: ${upcomingDays} days ---`);

  const cmd = `${BINARY} -u ${upcomingDays} -c ${CONFIG}`;
  let output;
  try {
    output = execSync(cmd, { encoding: "utf-8", cwd: path.join(__dirname, "..") });
  } catch (e) {
    console.error("Execution failed:", e.stderr);
    process.exit(1);
  }

  const events = JSON.parse(output) || [];
  const { start: windowStart } = getTodayBoundaries();
  const windowEnd = new Date(windowStart);
  windowEnd.setDate(windowEnd.getDate() + upcomingDays);
  windowEnd.setHours(23, 59, 59, 999);

  console.log(
    `Window: ${windowStart.toISOString()} to ${windowEnd.toISOString()}`,
  );
  console.log(`Found ${events.length} events.`);

  events.forEach((event) => {
    const eStart = parseICalDate(event.Start);
    const eEnd = parseICalDate(event.ActualEnd);

    // 1. Temporal Integrity
    if (eEnd < windowStart) {
      throw new Error(
        `Event ${event.UID} ends before window start! End: ${eEnd.toISOString()}, WindowStart: ${windowStart.toISOString()}`,
      );
    }
    if (eStart > windowEnd) {
      throw new Error(
        `Event ${event.UID} starts after window end! Start: ${eStart.toISOString()}, WindowEnd: ${windowEnd.toISOString()}`,
      );
    }

    // 2. Ongoing Logic
    const isActuallyOngoing = eStart < windowStart;
    if (event.Ongoing !== isActuallyOngoing) {
      throw new Error(
        `Event ${event.UID} Ongoing flag mismatch! Expected: ${isActuallyOngoing}, Found: ${event.Ongoing}`,
      );
    }
    if (isActuallyOngoing && !event.Description.startsWith("Ongoing")) {
      throw new Error(
        `Event ${event.UID} is ongoing but description missing prefix!`,
      );
    }

    // 3. Calculation Check
    const diffHours = (eEnd - eStart) / (1000 * 60 * 60);
    // Precision check due to potential float issues
    if (Math.abs(event.Hours - diffHours) > 0.001) {
      throw new Error(
        `Event ${event.UID} Hours mismatch! Expected: ${diffHours}, Found: ${event.Hours}`,
      );
    }

    // 4. Flags consistency
    if (diffHours < 24 && !event.SubDay)
      throw new Error(`Event ${event.UID} should be SubDay`);
    if (diffHours === 24 && !event.Day)
      throw new Error(`Event ${event.UID} should be Day`);
    if (diffHours > 24 && !event.MultiDay)
      throw new Error(`Event ${event.UID} should be MultiDay`);

    // 5. Human Labels
    const isToday = eStart.toDateString() === new Date().toDateString();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const isTomorrow = eStart.toDateString() === tomorrow.toDateString();

    if (isToday && !event.HumanStart.includes("TODAY")) {
      throw new Error(
        `Event ${event.UID} occurs today but HumanStart missing TODAY label!`,
      );
    }
    if (isTomorrow && !event.HumanStart.includes("TOMORROW")) {
      throw new Error(
        `Event ${event.UID} occurs tomorrow but HumanStart missing TOMORROW label!`,
      );
    }
  });

  console.log("Passed logic validation.");
}

try {
  // Run multiple window sizes
  runTest(0);
  runTest(1);
  runTest(7);
  console.log("\n✅ All programmatic checks passed!");
} catch (err) {
  console.error("\n❌ Logic Validation Failed:");
  console.error(err.message);
  process.exit(1);
}
