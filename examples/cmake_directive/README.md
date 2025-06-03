# Example Integration with Gazelle Foreign-CC

This directory demonstrates the complete workflow for using the `gazelle:cmake` directive.

## Setup

1. **External Repository Definition** (MODULE.bazel):
```starlark
http_archive(
    name = "example_cmake_lib",
    urls = ["https://github.com/example/cmake-lib/archive/v1.0.tar.gz"],
    strip_prefix = "cmake-lib-1.0",
    build_file_content = '''
filegroup(
    name = "srcs",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)
'''
)
```

2. **Package with CMake Directive** (thirdparty/example_cmake_lib/BUILD.bazel):
```starlark
# gazelle:cmake @example_cmake_lib//:srcs
```

3. **Run Gazelle**:
```bash
bazel run //:gazelle -- thirdparty/example_cmake_lib
```

## Expected Output

For an external CMake project with:
```cmake
cmake_minimum_required(VERSION 3.10)
project(ExampleLib)

add_library(example_lib 
    src/example.cpp
    src/utils.cpp
)

target_include_directories(example_lib PUBLIC include)

add_executable(example_app src/main.cpp)
target_link_libraries(example_app example_lib)
```

Gazelle will populate the BUILD.bazel file with:
```starlark
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")

# gazelle:cmake @example_cmake_lib//:srcs

cc_library(
    name = "example_lib",
    srcs = [
        "src/example.cpp", 
        "src/utils.cpp",
    ],
    includes = ["include"],
    visibility = ["//visibility:public"],
)

cc_binary(
    name = "example_app",
    srcs = ["src/main.cpp"],
    deps = [":example_lib"],
)
```

## Notes

- The generated rules reference source files from the external repository
- Include directories are properly mapped to the `includes` attribute
- Library dependencies are resolved and added to the `deps` attribute
- The original `gazelle:cmake` directive is preserved in the BUILD file