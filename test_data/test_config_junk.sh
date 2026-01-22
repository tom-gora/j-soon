#!/usr/bin/env bash
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
CLEAN_CONF="$SCRIPT_DIR/test_calendars.conf"
JUNK_CONF="$SCRIPT_DIR/test_junk.conf"

# Create a "junk" config file
cat <<EOF > "$JUNK_CONF"
# This is a comment

  ./test_data/calendars/calendar.ics  
    
# Another comment with spaces after
   

./test_data/calendars/birthdays.ics

# End of file with spaces
  
EOF

echo "Testing config parser robustness with 'junk' entries..."

# 1. Run with clean config
CLEAN_JSON=$(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -c "$CLEAN_CONF" -u 30)
CLEAN_COUNT=$(echo "$CLEAN_JSON" | jq 'length')

# 2. Run with junk config
JUNK_JSON=$(cd "$SCRIPT_PARENT_DIR" && ./bin/jfi -c "$JUNK_CONF" -u 30)
JUNK_COUNT=$(echo "$JUNK_JSON" | jq 'length')

# Clean up
rm -f "$JUNK_CONF"

if [ "$CLEAN_COUNT" -eq "$JUNK_COUNT" ]; then
    echo "Success: Config parser correctly handled junk lines. Counts match: $CLEAN_COUNT"
    exit 0
else
    echo "FAIL: Config parser results differ when using junk config!"
    echo "Clean count: $CLEAN_COUNT"
    echo "Junk count: $JUNK_COUNT"
    # Note: User explicitly asked NOT to fix this yet, just report it.
    exit 1
fi
