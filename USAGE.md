# Gazelle Foreign CC Usage Guide

This document describes how to use the gazelle-foreign-cc plugin to generate Bazel BUILD rules from CMake projects.

## Quick Start

The gazelle-foreign-cc plugin can generate Bazel `cc_library`, `cc_binary`, and `cc_test` rules from:
- Local CMake projects
- External CMake dependencies

### Prerequisites

- Bazel with bzlmod enabled
- Go toolchain
- CMake (for projects using CMake File API)

## Usage Patterns

### 1. Local CMake Projects

For CMake projects within your repository:

```bash
# Run gazelle on a specific directory
bazel run //:gazelle -- path/to/cmake/project

# Run gazelle on the entire repository
bazel run //:gazelle
```

Example with a simple CMake project:
```cmake
# CMakeLists.txt
add_library(mylib src/lib.cpp)
add_executable(myapp src/main.cpp)
target_link_libraries(myapp mylib)
```

Gazelle will generate:
```starlark
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")

cc_library(
    name = "mylib",
    srcs = ["src/lib.cpp"],
)

cc_binary(
    name = "myapp",
    srcs = ["src/main.cpp"],
    deps = [":mylib"],
)
```

### 2. External CMake Projects

For external CMake dependencies, use the `gazelle:cmake` directive.

#### Step 1: Define External Project in MODULE.bazel

```starlark
# Use http_archive with bzlmod
http_archive = use_repo_rule("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "somelib",
    urls = ["https://github.com/example/somelib/archive/v1.0.0.tar.gz"],
    strip_prefix = "somelib-1.0.0",
    build_file = "@gazelle-foreign-cc//:external_repo_buildfile.BUILD",
)
```

Create an `external_repo_buildfile.BUILD` file in your repository root:
```starlark
filegroup(
    name = "srcs",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)
```

#### Step 2: Create Package with CMake Directive

Create a package directory (e.g., `thirdparty/somelib/`) with a BUILD.bazel file:

```starlark
# gazelle:cmake @somelib//:srcs
```

#### Step 3: Run Gazelle

```bash
bazel run //:gazelle -- thirdparty/somelib
```

This will populate the BUILD.bazel file with `cc_library`, `cc_binary`, and `cc_test` rules based on the CMake targets found in the external repository.

## Configuration Directives

The plugin supports several configuration directives:

### `gazelle:cmake`
Specifies an external CMake repository to process:
```starlark
# gazelle:cmake @somelib//:srcs
```

### `gazelle:cmake_executable`
Specifies the path to the CMake executable:
```starlark
# gazelle:cmake_executable /usr/bin/cmake3
```

### `gazelle:cmake_define`
Passes CMake definitions (not yet fully implemented):
```starlark
# gazelle:cmake_define CMAKE_BUILD_TYPE Release
```

## How It Works

1. **Directive Detection**: Gazelle finds `gazelle:cmake` directives in BUILD.bazel files
2. **Label Parsing**: The directive value (e.g., `@somelib//:srcs`) is parsed to extract the repository name
3. **Repository Resolution**: The external repository is located in the Bazel external directory
4. **CMake Processing**: CMakeLists.txt files are processed using:
   - CMake File API (preferred method)
   - Regex-based parsing (fallback for compatibility)
5. **Rule Generation**: Bazel BUILD rules are generated based on discovered CMake targets

## Supported CMake Constructs

Currently supported CMake constructs:
- `add_library()` → `cc_library`
- `add_executable()` → `cc_binary`  
- `target_include_directories()` → `includes` attribute
- `target_link_libraries()` → `deps` attribute
- Basic source file detection

## Examples

See the [examples directory](examples/) for working demonstrations:
- **simple_hello**: Basic "Hello, World!" CMake project
- **libzmq**: Real-world ZeroMQ library integration
- **cmake_directive**: External CMake project usage

To run the examples:
```bash
cd examples
bazel run //:gazelle    # Generate BUILD files
bazel build //...       # Build all examples
```

## Development Status

**Current Capabilities:**
- ✅ Basic C++ rule generation
- ✅ Directive handling
- ✅ External repository support
- ✅ CMake File API integration
- ✅ Regex fallback parsing

**Work in Progress:**
- 🚧 Advanced dependency resolution
- 🚧 Complex CMake constructs
- 🚧 Include path scanning
- 🚧 Test rule generation

## Limitations

- Complex CMake logic may not be fully captured
- Advanced CMake features not yet supported
- File paths are relative to the CMake project root
- Some CMake variables and conditionals not processed

## Troubleshooting

If gazelle doesn't generate expected rules:
1. Ensure CMake is installed (`cmake --version`)
2. Check that CMakeLists.txt files are valid
3. Verify external repository configuration
4. Check gazelle logs for parsing errors

For more information, see the [development guide](CLAUDE.md).