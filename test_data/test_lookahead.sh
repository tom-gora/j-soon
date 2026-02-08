#!/bin/bash

# Configuration
BINARY="./bin/jsoon"
CONFIG="test_data/test_config.json"
FAILED=0

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "--- Testing Lookahead Logic ---"

run_test() {
  local days=$1
  echo -n "Testing -u $days: "

  # Calculate window boundaries
  # Use explicit YYYY-MM-DD to avoid ambiguity in relative date parsing
  local base_date=$(date +%Y-%m-%d)
  local start_ts=$(date -d "$base_date" +%s)
  local end_ts=$(date -d "$base_date + $((days + 1)) days - 1 second" +%s)

  # Run jsoon and use jq to validate every event
  local output=$($BINARY -u "$days" -c "$CONFIG" -f stdout 2>/dev/null)

  if [ $? -ne 0 ]; then
    echo -e "${RED}FAILED (Binary exited with error)${NC}"
    FAILED=1
    return
  fi

  # Validation Logic:
  # 1. UnixEnd >= start_ts (Event overlaps with window from the beginning)
  # 2. UnixStart <= end_ts (Event overlaps with window before it ends)
  # 3. Ongoing check: if UnixStart < start_ts, Ongoing must be true
  # 4. If UnixStart >= start_ts, Ongoing must be false
  local validation=$(echo "$output" | jq -r --argjson s "$start_ts" --argjson e "$end_ts" '
        .[] | select(
            (.UnixEnd < $s) or 
            (.UnixStart > $e) or 
            (.UnixStart < $s and .Ongoing != true) or
            (.UnixStart >= $s and .Ongoing != false)
        ) | "Error: Event \(.UID) out of bounds or flag mismatch (Start: \(.UnixStart), End: \(.UnixEnd), Window: \($s)-\($e), Ongoing: \(.Ongoing))"
    ')

  if [ -z "$validation" ]; then
    echo -e "${GREEN}PASSED${NC}"
  else
    echo -e "${RED}FAILED${NC}"
    echo "$validation"
    FAILED=1
  fi
}

# Ensure binary exists
if [ ! -f "$BINARY" ]; then
  echo "Binary not found. Running make build..."
  make build >/dev/null
fi

# Run tests for different lookahead windows
run_test 0
run_test 1
run_test 7

if [ $FAILED -eq 0 ]; then
  echo -e "\n${GREEN}✅ All lookahead checks passed!${NC}"
  exit 0
else
  echo -e "\n${RED}❌ Lookahead validation failed!${NC}"
  exit 1
fi
