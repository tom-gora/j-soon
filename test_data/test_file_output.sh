#!/usr/bin/env bash
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
CALENDARS_CONF_FILE="$SCRIPT_DIR/test_calendars.conf"
TEST_OUT_FILE="$SCRIPT_DIR/test_output.json"

# Cleanup from previous runs
rm -f "$TEST_OUT_FILE"

echo "Testing file output flag (-f $TEST_OUT_FILE)..."
(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -f "$TEST_OUT_FILE" -c "$CALENDARS_CONF_FILE")

# 1. Check if file exists
if [ ! -f "$TEST_OUT_FILE" ]; then
    echo "Error: Output file was not created at $TEST_OUT_FILE"
    exit 1
fi

# 2. Check if it's valid JSON
if ! jq -e . "$TEST_OUT_FILE" >/dev/null 2>&1; then
    echo "Error: Output file does not contain valid JSON."
    rm -f "$TEST_OUT_FILE"
    exit 1
fi

COUNT=$(jq 'length' "$TEST_OUT_FILE")
echo "Success: File output contains valid JSON ($COUNT events)."

# 3. Cleanup
rm -f "$TEST_OUT_FILE"
exit 0
