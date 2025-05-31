# Gazelle Foreign CC Plugin

[![Build Status](https://github.com/OWNER/REPO/actions/workflows/build.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/build.yml)

A Bazel Gazelle plugin to generate C++ (`cc_library`, `cc_binary`, `cc_test`) rules from existing CMake projects.

## Examples

The [examples directory](/examples) contains real-world demonstrations of the gazelle-foreign-cc plugin. The examples directory is a separate Bazel module that references the main plugin using a local path override.

Available examples:
- **simple_hello**: Basic C++ "Hello, World!" project
- **libzmq**: Real-world example using ZeroMQ messaging library

To build and test examples:
```bash
cd examples
bazel build //...
```

To run gazelle on examples:
```bash
cd examples
bazel run //:gazelle
```

## Current Status

This plugin is under active development. Currently, it can:
*   Be built using Bazel.
*   Recognize basic C++ source files and generate simple `cc_library`, `cc_binary`, or `cc_test` rules by convention.
*   Handle basic configuration directives (e.g., `gazelle:cmake_executable`).

**Note:** Full CMake `CMakeLists.txt` parsing and sophisticated dependency resolution are not yet implemented.

## TODO (Key Next Steps)
*   Implement robust parsing of `CMakeLists.txt`.
*   Implement C++ include scanning for dependency resolution.
*   Add comprehensive tests for various CMake project structures.

## Prerequisites
*   Bazel
*   Go
*   A C++ toolchain configured for Bazel

## Building (the plugin itself)
```bash
bazel build //gazelle:gazelle-foreign-cc
```

## Development

### Module Dependencies
This project uses Bazel's module system (bzlmod). After making changes to `MODULE.bazel`, ensure the lock file is up-to-date:

```bash
bazel mod tidy
```

The CI will fail if `MODULE.bazel.lock` is not synchronized with `MODULE.bazel`.

## Usage
(Details to be added once the plugin is more feature-complete)

```bash
# Example of how Gazelle is typically run
# bazel run //:gazelle -- path/to/your/cmake_project
```

---
*This README is a work in progress and will be updated as the plugin evolves.*
