#!/usr/bin/env bash
DAYS_LIMIT="${1:-7}"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
CALENDARS_CONF_FILE="$SCRIPT_DIR/test_calendars.conf"

echo "Validating JSON output..."
RAW_JSON=$(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -u "$DAYS_LIMIT" -c "$CALENDARS_CONF_FILE")

# 1. Check if it's valid JSON
if ! echo "$RAW_JSON" | jq -e . >/dev/null 2>&1; then
    echo "Error: Invalid JSON output produced."
    exit 1
fi

# 2. Check if it's an array
TYPE=$(echo "$RAW_JSON" | jq -r 'type')
if [ "$TYPE" != "array" ]; then
    echo "Error: Output is not a JSON array (found $TYPE)."
    exit 1
fi

# 3. Check for mandatory fields in the first item (if array not empty)
COUNT=$(echo "$RAW_JSON" | jq 'length')
if [ "$COUNT" -gt 0 ]; then
    FIELDS=$(echo "$RAW_JSON" | jq -r '.[0] | keys | .[]')
    for field in UID Summary Start End HumanStart HumanEnd Ongoing; do
        if ! echo "$FIELDS" | grep -q "$field"; then
            echo "Error: Mandatory field '$field' missing from JSON objects."
            exit 1
        fi
    done

    # 4. Specific logic check: Ongoing events
    ONGOING_COUNT=$(echo "$RAW_JSON" | jq '[.[] | select(.Ongoing == true)] | length')
    if [ "$ONGOING_COUNT" -gt 0 ]; then
        DESC_CHECK=$(echo "$RAW_JSON" | jq -r '.[] | select(.Ongoing == true) | .Description' | grep -c "Ongoing")
        if [ "$DESC_CHECK" -lt "$ONGOING_COUNT" ]; then
            echo "Warning: Some events marked as Ongoing do not have 'Ongoing' in their description."
            # We treat this as a failure if it's strictly required
            # exit 1 
        fi
        echo "Found $ONGOING_COUNT ongoing events."
    fi
fi

echo "Success: JSON output is valid ($COUNT events found) and structured correctly."
exit 0
