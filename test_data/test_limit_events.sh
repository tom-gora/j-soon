#!/usr/bin/env bash
LIMIT=2
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
CALENDARS_CONF_FILE="$SCRIPT_DIR/test_calendars.conf"

echo "Testing event limit flag (-l $LIMIT)..."
RAW_JSON=$(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -l "$LIMIT" -c "$CALENDARS_CONF_FILE")

COUNT=$(echo "$RAW_JSON" | jq 'length')
if [ "$COUNT" -ne "$LIMIT" ]; then
    echo "Error: Expected $LIMIT events, but found $COUNT."
    exit 1
fi

echo "Success: Limit flag correctly restricted output to $LIMIT events."
exit 0
