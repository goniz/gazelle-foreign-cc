# GitHub Copilot Coding Agent Setup

This directory contains setup files for GitHub Copilot Coding Agent to work effectively with the gazelle-foreign-cc repository.

## Files

- `setup.sh` - Setup script that installs all necessary tools and dependencies

## Dependencies Installed

The setup script installs the following tools required for developing and testing the Bazel Gazelle plugin:

- **Build tools**: build-essential, curl, git, unzip, wget
- **CMake**: For processing C++ projects that this plugin targets
- **Go 1.23.9**: For developing the Gazelle plugin itself (Go 1.22+ required)
- **Bazelisk**: Bazel version manager for building the plugin

## Usage

This setup is automatically executed by GitHub Copilot Coding Agent when working on this repository. The script can also be run manually:

```bash
chmod +x .github/copilot/setup.sh
./.github/copilot/setup.sh
```

## Verification

After setup, you can verify the installation by building the project:

```bash
bazel build //gazelle:gazelle-foreign-cc
```

Or run the local CI script:

```bash
./run_ci_locally.sh
```