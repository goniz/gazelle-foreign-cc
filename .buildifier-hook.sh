#!/bin/bash
# Pre-commit hook script for buildifier that automatically stages formatted files

# Exit on any error
set -e

# Store the list of files passed to the script
FILES=("$@")

# Run buildifier on the files
buildifier -mode=fix "${FILES[@]}"

# Add the formatted files back to the staging area
for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        git add "$file"
    fi
done