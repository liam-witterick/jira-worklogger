#!/bin/bash

# Script to create a new release tag and push it to GitHub

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.1"
    exit 1
fi

VERSION=$1

echo "Creating release v$VERSION..."

# Check if on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    echo "Warning: You are not on the main or master branch."
    echo "Current branch: $CURRENT_BRANCH"
    read -p "Do you want to continue? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting..."
        exit 1
    fi
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: You have uncommitted changes. Please commit or stash them before creating a release."
    exit 1
fi

# Update version in version.go
sed -i "s/^	Version   = \".*\"/	Version   = \"$VERSION\"/" version.go

# Commit the version change
git add version.go
git commit -m "Bump version to v$VERSION"

# Create tag
git tag -a "v$VERSION" -m "Release v$VERSION"

# Push commit and tag
echo "Pushing changes to remote..."
git push origin "$(git branch --show-current)"
git push origin "v$VERSION"

echo "Release v$VERSION created and pushed to GitHub."
echo "GitHub Actions workflow will now build and publish the release."
echo "Check the status at: https://github.com/liam-witterick/jira-worklogger/actions"
