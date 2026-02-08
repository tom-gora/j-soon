#!/usr/bin/env bash

DAYS_LIMIT="$1"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROCESSOR_BIN="$SCRIPT_DIR/bin/jsoon"

# remove when using build binary
make build

RAW_JSON_ARR=$("$PROCESSOR_BIN" -u "$DAYS_LIMIT")

echo "$RAW_JSON_ARR" | jq -c '.[]' | while read -r obj_row; do

  HEADER=$(echo "$obj_row" | jq -r ". | .HumanStart")
  BODY="$(echo "$obj_row" | jq -r ". | .Summary")\n\n$(echo "$obj_row" | jq -r ". | .Description")"

  notify-send "$HEADER" "$BODY" -u normal -t 10000
  sleep 0.25 #stagger

done

# remove when using build binary
make clean

exit 0
