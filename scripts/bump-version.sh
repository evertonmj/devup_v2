#!/bin/bash
# Script to bump version number

set -e

VERSION_FILE="VERSION"
ROOT_GO_FILE="cmd/root.go"

# Read current version
if [ ! -f "$VERSION_FILE" ]; then
    echo "1.0.0" > "$VERSION_FILE"
fi

CURRENT_VERSION=$(cat "$VERSION_FILE")

# Parse version
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

# Determine bump type (default: patch)
BUMP_TYPE="${1:-patch}"

case "$BUMP_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Unknown bump type: $BUMP_TYPE"
        echo "Usage: $0 [major|minor|patch]"
        exit 1
        ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"

# Update VERSION file
echo "$NEW_VERSION" > "$VERSION_FILE"

# Update root.go
if [ -f "$ROOT_GO_FILE" ]; then
    # macOS compatible sed
    sed -i '' "s/Version: \".*\"/Version: \"$NEW_VERSION\"/" "$ROOT_GO_FILE"
fi

echo "Version bumped: $CURRENT_VERSION → $NEW_VERSION"
echo "$NEW_VERSION"
