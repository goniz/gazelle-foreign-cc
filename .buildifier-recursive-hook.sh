#!/bin/bash
# Pre-commit hook script that runs buildifier with -r flag on all files

# Exit on any error  
set -e

# Run buildifier recursively on the current directory
buildifier -r -mode=fix .

# Stage all modified Bazel files
git diff --name-only --diff-filter=M | grep -E '(BUILD|BUILD\.bazel|WORKSPACE|WORKSPACE\.bazel|\.bzl|\.bazel)$' | while read -r file; do
    git add "$file"
done || true  # Don't fail if no files match