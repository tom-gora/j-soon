#!/usr/bin/env bash
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
TEST_FILE="$SCRIPT_DIR/calendars/calendar.ics"

echo "Testing pipeline (stdin) support..."
RAW_JSON=$(cat "$TEST_FILE" | (cd "$SCRIPT_PARENT_DIR" && ./bin/jsoon -u 10 -f stdout))

# Validate JSON and check if events were found
COUNT=$(echo "$RAW_JSON" | jq 'length')

if [ "$COUNT" -gt 0 ]; then
  echo "Success: Piped input correctly processed ($COUNT events found)."
  exit 0
else
  echo "Error: No events found when reading from stdin."
  exit 1
fi
