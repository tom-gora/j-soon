#!/usr/bin/env bash
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
SCRIPT_PARENT_DIR=$(cd -- "$SCRIPT_DIR/.." &>/dev/null && pwd)
BINARY="$SCRIPT_PARENT_DIR/bin/jsoon"

echo "Testing informational flags..."

# 1. Test Version
echo "Checking -V (version)..."
if ! "$BINARY" -V | grep -q "Version"; then
  echo "Error: -V flag did not report version correctly."
  exit 1
fi

# 2. Test Help (usage)
echo "Checking -h (help)..."
# Help often prints to stderr, and -h usually exits with 0 or 2 depending on implementation
# flag package in Go exits with 2 for -h if not overridden, but we check if it prints the usage
USAGE=$("$BINARY" -h 2>&1)
if ! echo "$USAGE" | grep -q "Usage:"; then
  echo "Error: -h flag did not display usage instructions."
  exit 1
fi

echo "Success: Informational flags are working."
exit 0
