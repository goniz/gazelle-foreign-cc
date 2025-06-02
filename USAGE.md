# Gazelle CMake Directive Usage

This document describes how to use the `gazelle:cmake` directive to generate BUILD rules from external CMake projects.

## Workflow

### 1. Define External Project in MODULE.bazel

```starlark
http_archive(
    name = "somelib",
    urls = ["https://github.com/example/somelib/archive/v1.0.0.tar.gz"],
    strip_prefix = "somelib-1.0.0",
    build_file_content = '''
filegroup(
    name = "srcs",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)
'''
)
```

### 2. Create Package with CMake Directive

Create a package directory (e.g., `thirdparty/somelib/`) with a BUILD.bazel file:

```starlark
# gazelle:cmake @somelib//:srcs
```

### 3. Run Gazelle

Run the gazelle binary to generate BUILD rules:

```bash
bazel run //:gazelle -- thirdparty/somelib
```

This will populate the BUILD.bazel file with `cc_library`, `cc_binary`, and `cc_test` rules based on the CMake targets found in the external repository.

## How It Works

1. **Directive Detection**: Gazelle finds the `gazelle:cmake` directive in BUILD.bazel files
2. **Label Parsing**: The directive value (e.g., `@somelib//:srcs`) is parsed to extract the repository name
3. **Repository Resolution**: The external repository is located in the Bazel external directory
4. **CMake Processing**: CMakeLists.txt files in the external repository are processed using:
   - CMake File API (preferred)
   - Regex-based parsing (fallback)
5. **Rule Generation**: Bazel BUILD rules are generated based on the discovered CMake targets

## Supported CMake Constructs

- `add_library()` → `cc_library`
- `add_executable()` → `cc_binary`
- `target_include_directories()`
- `target_link_libraries()`

## Example Generated Output

For a CMakeLists.txt with:
```cmake
add_library(mylib src/mylib.cpp)
target_include_directories(mylib PUBLIC include)

add_executable(myapp src/main.cpp)
target_link_libraries(myapp mylib)
```

Gazelle will generate:
```starlark
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")

cc_library(
    name = "mylib",
    srcs = ["src/mylib.cpp"],
    includes = ["include"],
)

cc_binary(
    name = "myapp", 
    srcs = ["src/main.cpp"],
    deps = [":mylib"],
)
```

## Limitations

- Only basic CMake constructs are currently supported
- External repository resolution uses standard Bazel locations
- Complex CMake logic may not be fully captured
- File paths are relative to the external repository root

## Configuration

You can configure the CMake executable used:

```starlark
# gazelle:cmake_executable /usr/bin/cmake3
# gazelle:cmake @somelib//:srcs
```