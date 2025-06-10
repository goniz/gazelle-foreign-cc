# Pre-commit Setup

This repository uses [pre-commit](https://pre-commit.com/) to automatically format Bazel files using `buildifier`.

## Installation

1. Install pre-commit:
```bash
pip install pre-commit
```

2. Install the git hook scripts:
```bash
pre-commit install
```

## Usage

The buildifier hook will automatically run when you commit changes. It will:

- Format all Bazel files (BUILD, *.bazel, *.bzl) using `buildifier -r`
- Automatically stage the formatted files so they are included in the commit
- Ensure consistent formatting across the codebase

### Manual execution

To run the hooks on all files manually:
```bash
pre-commit run --all-files
```

To run only the buildifier hook:
```bash
pre-commit run buildifier --all-files
```

## Configuration

The pre-commit configuration is defined in `.pre-commit-config.yaml` and uses a custom script `.buildifier-recursive-hook.sh` that:

1. Runs `buildifier -r -mode=fix .` to format all Bazel files recursively
2. Stages any modified files so they are included in the commit

This ensures that all Bazel files in the repository maintain consistent formatting.