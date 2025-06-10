#!/bin/bash
#
# Install git hooks for the gazelle-foreign-cc repository
# This script copies the pre-commit hook to .git/hooks/ and makes it executable

set -e

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "Error: This script must be run from the root of the git repository"
    exit 1
fi

# Check if buildifier is available
if ! command -v buildifier &> /dev/null; then
    echo "Warning: buildifier is not installed or not in PATH"
    echo "Please install buildifier before using the git hooks"
    echo "See: https://github.com/bazelbuild/buildtools/releases"
fi

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy the pre-commit hook
echo "Installing pre-commit hook..."
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo "Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will now automatically:"
echo "  - Format all Bazel files using buildifier -r"
echo "  - Stage the formatted files in your commit"
echo ""
echo "To disable the hook temporarily, use: git commit --no-verify"