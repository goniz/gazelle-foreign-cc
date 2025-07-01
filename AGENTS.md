# AGENTS.md - Comprehensive Development Guide for gazelle-foreign-cc

**gazelle-foreign-cc** is a Bazel Gazelle plugin that generates C++ BUILD rules (`cc_library`, `cc_binary`, `cc_test`) from CMake projects, bridging CMake and Bazel build systems.

## Table of Contents
1. [Project Overview](#project-overview)
2. [Core Functionality](#core-functionality)
3. [Project Structure](#project-structure)
4. [Key Files for Development](#key-files-for-development)
5. [Usage Patterns](#usage-patterns)
6. [Configuration Directives](#configuration-directives)
7. [CMake File API Integration](#cmake-file-api-integration)
8. [Development Commands](#development-commands)
9. [Examples and Testing](#examples-and-testing)
10. [Git Hooks and Code Formatting](#git-hooks-and-code-formatting)
11. [Implementation Status](#implementation-status)
12. [Code Conventions](#code-conventions)
13. [Troubleshooting](#troubleshooting)

## Project Overview

This plugin serves as a bridge between CMake and Bazel build systems, automatically generating Bazel BUILD rules from existing CMake projects. It supports both local CMake projects within the repository and external CMake dependencies.

### Key Features
- Parses CMakeLists.txt files using CMake File API (primary) or regex (fallback)
- Generates Bazel `cc_*` rules automatically
- Supports external CMake dependencies via `gazelle:cmake` directives
- Handles complex project structures with subdirectories
- Provides graceful fallback mechanisms for compatibility

## Core Functionality

### Supported CMake Constructs
- `add_library()` → `cc_library`
- `add_executable()` → `cc_binary`  
- `target_include_directories()` → `includes` attribute
- `target_link_libraries()` → `deps` attribute
- Basic source file detection
- Subdirectory support with multiple CMakeLists.txt files

### Target Types Supported
- **Executables**: Converted to `cc_binary` rules
- **Static Libraries**: Converted to `cc_library` rules
- **Shared Libraries**: Converted to `cc_library` rules with appropriate linkage
- **Interface Libraries**: Header-only libraries

## Project Structure

```
gazelle-foreign-cc/
├── gazelle/               # Plugin configuration and core logic
│   ├── config.go         # CMake directives (gazelle:cmake_executable, gazelle:cmake)
│   ├── generate.go       # Regex-based parsing fallback
│   ├── cmake_api.go      # CMake File API implementation
│   ├── resolve.go        # Dependency resolution
│   ├── main.go          # Test program
│   └── plugin.go         # Plugin registration
├── language/             # Gazelle language implementation
│   ├── cmake.go          # Main language.Language interface
│   └── cmake_api.go      # CMake File API integration
├── examples/             # Test projects (separate Bazel module)
│   ├── simple_hello/     # Basic C++ "Hello, World!" project
│   ├── libzmq/          # Real-world ZeroMQ library integration
│   ├── cmake_directive/  # External CMake project usage
│   └── README.md        # Examples documentation
├── testdata/             # CMake test patterns
│   ├── simple_cc_project/    # Basic test case
│   ├── complex_cc_project/   # Advanced test case
│   └── invalid_cmake_project/ # Error handling test
├── hooks/                # Git hooks for code formatting
├── .github/              # GitHub configuration
└── .openhands/           # OpenHands microagents configuration
```

## Key Files for Development

### Core Implementation Files
- **`language/cmake.go`** - Main language logic, implements Gazelle interface
- **`gazelle/config.go`** - Configuration and directive handling
- **`language/cmake_api.go`** - CMake File API integration
- **`gazelle/cmake_api.go`** - Complete CMake File API implementation with JSON structures
- **`gazelle/generate.go`** - Rule generation with File API integration
- **`gazelle/resolve.go`** - Improved dependency resolution

### Testing and Examples
- **`examples/`** - Add example projects for testing
- **`testdata/`** - Add test cases for new CMake patterns
- **`test_cmake_api.sh`** - CMake File API testing script

### Configuration Files
- **`MODULE.bazel`** - Bazel module dependencies
- **`BUILD.bazel`** - Build configuration
- **`.bazelrc`** - Bazel configuration
- **`go.mod`** - Go module dependencies

## Usage Patterns

### 1. Local CMake Projects

For CMake projects within your repository:

```bash
# Run gazelle on a specific directory
bazel run //:gazelle -- path/to/cmake/project

# Run gazelle on the entire repository
bazel run //:gazelle
```

Example CMake project:
```cmake
# CMakeLists.txt
add_library(mylib src/lib.cpp)
add_executable(myapp src/main.cpp)
target_link_libraries(myapp mylib)
```

Generated Bazel rules:
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

For external CMake dependencies:

#### Step 1: Define External Project in MODULE.bazel
```starlark
http_archive = use_repo_rule("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "somelib",
    urls = ["https://github.com/example/somelib/archive/v1.0.0.tar.gz"],
    strip_prefix = "somelib-1.0.0",
    build_file = "@gazelle-foreign-cc//:external_repo_buildfile.BUILD",
)
```

#### Step 2: Create Package with CMake Directive
```starlark
# gazelle:cmake @somelib//:srcs
```

#### Step 3: Run Gazelle
```bash
bazel run //:gazelle -- thirdparty/somelib
```

## Configuration Directives

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

## CMake File API Integration

### Overview
The CMake File API integration provides robust and accurate parsing of CMake projects, replacing regex-based approaches with CMake's official API.

### Key Features
- **Accurate Target Parsing**: Extracts executables, libraries and their properties
- **Source File Detection**: Identifies C/C++ source files, headers, and relationships
- **Dependency Resolution**: Maps target dependencies and linked libraries
- **Include Directory Handling**: Processes global and target-specific include paths
- **Error Handling**: Graceful fallback to regex parsing when File API fails

### Implementation Details
- Creates File API queries requesting codemodel information
- Parses JSON responses to extract target definitions
- Supports CMake 3.14+ for full File API support
- Handles configuration failures with automatic fallback

### Benefits Over Regex Parsing
1. **Accuracy**: No parsing ambiguities or edge cases
2. **Completeness**: Access to all CMake target information
3. **Maintainability**: No complex regex patterns to maintain
4. **Robustness**: Handles complex CMake constructs correctly
5. **Future-proof**: Leverages CMake's official API

## Development Commands

### Building and Testing
```bash
# Build plugin
bazel build //gazelle:gazelle-foreign-cc

# Test with examples
cd examples && bazel run //:gazelle && bazel build //...

# Update dependencies  
bazel mod tidy

# Run tests
bazel test //...

# Test CMake File API specifically
./test_cmake_api.sh

# Run buildifier for code formatting
buildifier -r -mode=fix .
```

### Bazel Configuration

#### Root MODULE.bazel Dependencies
```starlark
bazel_dep(name = "rules_go", version = "0.54.1")     # Go toolchain
bazel_dep(name = "gazelle", version = "0.43.0")     # Gazelle framework  
bazel_dep(name = "rules_cc", version = "0.1.1")     # C++ rules
```

#### Generated Rule Pattern
```starlark
cc_library(
    name = "target_name",
    srcs = ["src/file.cpp"],
    hdrs = ["include/header.h"],
    includes = ["include"],
    deps = [":dependency"],
)
```

## Examples and Testing

### Available Examples

#### 1. Simple Hello World (`examples/simple_hello/`)
- Basic C++ "Hello, World!" project
- Demonstrates simple CMake to Bazel conversion
- Files: `CMakeLists.txt`, `main.cpp`, `BUILD.bazel`

#### 2. LibZMQ (`examples/libzmq/`)
- Real-world ZeroMQ messaging library integration
- Complex CMake project with multiple targets
- Demonstrates external dependency handling

#### 3. CMake Directive (`examples/cmake_directive/`)
- Shows complete workflow for `gazelle:cmake` directive
- External repository integration example
- Demonstrates BUILD file generation from external CMake projects

### Testing Strategy
```bash
# Build all examples
cd examples && bazel build //...

# Run gazelle on examples
cd examples && bazel run //:gazelle

# Test specific example
cd examples/simple_hello && bazel run //simple_hello:hello
```

### Test Projects in testdata/
- **simple_cc_project/**: Basic test case
- **complex_cc_project/**: Advanced test case with subdirectories
- **invalid_cmake_project/**: Error handling and fallback testing

## Git Hooks and Code Formatting

### Setup
The repository uses git hooks for automatic code formatting:

```bash
# Install git hooks
./install-hooks.sh

# Manual installation
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Functionality
- Automatically formats Bazel files using `buildifier`
- Runs on every commit
- Stages formatted files automatically
- Ensures consistent code formatting

### Manual Formatting
```bash
# Format all Bazel files
buildifier -r -mode=fix .

# Bypass hook (not recommended)
git commit --no-verify
```

## Implementation Status

### Current Capabilities ✅
- Basic C++ rule generation (`cc_library`, `cc_binary`, `cc_test`)
- Directive handling (`gazelle:cmake`, `gazelle:cmake_executable`)
- External repository support
- CMake File API integration
- Regex fallback parsing
- Source file detection and dependency mapping
- Include directory processing
- Git hooks for code formatting

### Work in Progress 🚧
- Advanced dependency resolution
- Complex CMake constructs (conditionals, loops, functions)
- Include path scanning for automatic dependency detection
- Comprehensive test rule generation
- Caching of File API responses
- Incremental updates based on file changes

### Future Enhancements 🔮
- Support for custom CMake properties
- Integration with Bazel workspace rules
- Advanced CMake features (generators, custom commands)
- Performance optimizations
- Better error reporting and debugging

## Code Conventions

### Go Code Standards
- Follow standard Go conventions (gofmt, golint)
- Use structured logging for debugging
- Implement proper error handling with fallbacks
- Write comprehensive tests for new features

### Bazel Standards
- Place `load()` statements at the top of BUILD files
- Use `lowercase_underscore` naming convention
- Declare appropriate visibility for targets
- Include comprehensive documentation

### CMake Integration
- Prioritize CMake File API over regex parsing
- Implement graceful fallbacks for compatibility
- Handle edge cases and error conditions
- Test with various CMake project structures

## Troubleshooting

### Common Issues

#### Gazelle doesn't generate expected rules
1. Ensure CMake is installed (`cmake --version`)
2. Check that CMakeLists.txt files are valid
3. Verify external repository configuration
4. Check gazelle logs for parsing errors

#### CMake File API failures
1. Verify CMake version (3.14+ required for full support)
2. Check for CMake configuration errors
3. Ensure build directory permissions
4. Review temporary build directory creation

#### Build failures after rule generation
1. Verify source file paths are correct
2. Check include directory mappings
3. Ensure dependency resolution is accurate
4. Review generated BUILD file syntax

### Debugging Commands
```bash
# Verbose gazelle output
bazel run //:gazelle -- --verbose path/to/project

# Test CMake File API directly
cd testdata/simple_cc_project
bazel run //gazelle:gazelle-foreign-cc

# Check CMake configuration
cmake -S . -B build --debug-output
```

### Development Rules
- Always ensure Bazel targets build successfully after changes
- Do NOT extend task scope without user approval
- Push code to GitHub feature branches when tasks are finished
- NEVER push directly to main branch
- Test changes against multiple example projects

### Prerequisites
- Bazel is installed and configured
- CMake is installed and available in PATH
- Go toolchain is properly set up
- C++ toolchain is configured for Bazel

---

This comprehensive guide provides all the necessary information for AI agents to effectively work on the gazelle-foreign-cc project, covering everything from basic usage to advanced development patterns and troubleshooting procedures.