#!/usr/bin/env bash

# set -x
DAYS_LIMIT="$1"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)

CALENDARS_CONF_FILE="$SCRIPT_DIR/test_calendars.conf"

echo "Emitting notifications for $DAYS_LIMIT days..."
RAW_JSON_ARR=$(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -u "$DAYS_LIMIT" -c "$CALENDARS_CONF_FILE")

COUNT=$(echo "$RAW_JSON_ARR" | jq 'length')
echo "Found $COUNT events to emit."

echo "$RAW_JSON_ARR" | jq -c '.[]' | while read -r obj_row; do

  HEADER=$(echo "$obj_row" | jq -r ". | .HumanStart")
  SUMMARY=$(echo "$obj_row" | jq -r ". | .Summary")
  BODY="$(echo "$obj_row" | jq -r ". | .Summary")\n\n$(echo "$obj_row" | jq -r ". | .Description")"

  echo "  -> Emitting: $SUMMARY"
  # notify-send might fail in headless environments, so we ignore errors
  notify-send "$HEADER" "$BODY" -u normal -t 10000 2>/dev/null || true
  sleep 0.25 #stagger

done

exit 0
