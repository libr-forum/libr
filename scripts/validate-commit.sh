#!/bin/bash

# Get the commit message from the file Git provides
COMMIT_MSG_FILE=$1
COMMIT_MSG=$(head -n1 "$COMMIT_MSG_FILE")

# Skip if it's a merge commit
if grep -q "^Merge " "$COMMIT_MSG_FILE"; then
    exit 0
fi

# Validate the commit message
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$REPO_ROOT/scripts/validate-commit.sh" ]; then
    "$REPO_ROOT/scripts/validate-commit.sh" "$COMMIT_MSG"
    exit $?
else
    echo "⚠️  Commit validator not found, skipping validation"
    exit 0
fi
