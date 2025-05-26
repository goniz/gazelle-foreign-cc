# Gazelle Foreign CC Plugin

[![Build Status](https://github.com/OWNER/REPO/actions/workflows/build.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/build.yml)

A Bazel Gazelle plugin to generate C++ (`cc_library`, `cc_binary`, `cc_test`) rules from existing CMake projects.

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

## Usage
(Details to be added once the plugin is more feature-complete)

```bash
# Example of how Gazelle is typically run
# bazel run //:gazelle -- path/to/your/cmake_project
```

---
*This README is a work in progress and will be updated as the plugin evolves.*
