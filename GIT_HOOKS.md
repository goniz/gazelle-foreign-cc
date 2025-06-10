# Git Hooks Setup

This repository uses native git hooks to automatically format Bazel files using `buildifier`.

## Installation

1. Make sure `buildifier` is installed and available in your PATH:
   ```bash
   # Install buildifier (choose one method)
   go install github.com/bazelbuild/buildtools/buildifier@latest
   # OR download from releases: https://github.com/bazelbuild/buildtools/releases
   ```

2. Install the git hooks by running the installation script:
   ```bash
   ./install-hooks.sh
   ```

Alternatively, you can manually install the hooks:
```bash
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Usage

The buildifier hook will automatically run when you commit changes. It will:

- Format all Bazel files (BUILD, *.bazel, *.bzl) using `buildifier -r`
- Automatically stage the formatted files so they are included in the commit
- Ensure consistent formatting across the codebase

### Bypassing the hook

To temporarily bypass the pre-commit hook (not recommended):
```bash
git commit --no-verify
```

### Manual execution

To run buildifier manually on all files:
```bash
buildifier -r -mode=fix .
```

## Configuration

The pre-commit hook is defined in `hooks/pre-commit` and:

1. Runs `buildifier -r -mode=fix .` to format all Bazel files recursively
2. Stages any modified files so they are included in the commit

This ensures that all Bazel files in the repository maintain consistent formatting.