# CLAUDE.md - Development Guide for gazelle-foreign-cc

This guide provides essential information for future coding agents working on the gazelle-foreign-cc project.

## Project Overview

**gazelle-foreign-cc** is a Bazel Gazelle plugin that generates C++ BUILD rules (`cc_library`, `cc_binary`, `cc_test`) from existing CMake projects. It bridges the gap between CMake-based C++ projects and Bazel's build system, allowing developers to automatically generate Bazel BUILD files from CMake project definitions.

### Project Motivation

- **Seamless Integration**: Enable CMake-based C++ projects to work within Bazel workspaces without manual BUILD file creation
- **Automated Build Rule Generation**: Parse CMakeLists.txt files and generate appropriate Bazel cc_* rules automatically  
- **Foreign Dependencies**: Support external CMake libraries as Bazel dependencies using `gazelle:cmake` directives
- **Development Workflow**: Reduce friction when incorporating existing CMake projects into Bazel-based builds

## Code Structure

### Core Directories

```
gazelle-foreign-cc/
├── gazelle/               # Gazelle plugin configuration and core logic
│   ├── config.go         # CMake configuration and directive handling  
│   ├── generate.go       # Rule generation logic (regex-based fallback)
│   └── plugin.go         # Plugin registration (placeholder)
├── language/             # Language implementation for Gazelle
│   ├── cmake.go          # Main language interface implementation
│   └── cmake_api.go      # CMake File API integration
├── examples/             # Example projects (separate Bazel module)
│   ├── simple_hello/     # Basic C++ example
│   └── libzmq/          # Real-world ZeroMQ example  
├── testdata/             # Test projects with various CMake structures
└── cmd/                  # Command-line tools
```

### Key Components

1. **gazelle/config.go**: Handles configuration and CMake directives like `gazelle:cmake_executable` and `gazelle:cmake`
2. **language/cmake.go**: Implements the `language.Language` interface required by Gazelle
3. **language/cmake_api.go**: Integrates with CMake's File API for robust project parsing
4. **gazelle/generate.go**: Provides regex-based CMake parsing as a fallback

## Bazel Module Structure

### Root Module (MODULE.bazel)
```starlark
module(
    name = "gazelle-foreign-cc",
    version = "0.1.0",
)

bazel_dep(name = "rules_go", version = "0.54.1")        # Go toolchain for plugin
bazel_dep(name = "gazelle", version = "0.43.0")        # Gazelle framework
bazel_dep(name = "bazel_skylib", version = "1.7.1")    # Bazel utilities
bazel_dep(name = "rules_cc", version = "0.1.1")        # C++ rules support
```

### Examples Module (examples/MODULE.bazel)
Separate module that references the main plugin via local path override, allowing real-world testing of the plugin.

## Bazel Module Purpose

- **Plugin Development**: Uses Go and Gazelle dependencies to build the plugin
- **C++ Rule Support**: Generates standard `rules_cc` targets (`cc_library`, `cc_binary`, `cc_test`)
- **Modular Testing**: Examples module provides isolation for testing plugin functionality
- **Dependency Management**: Uses bzlmod for consistent, reproducible builds

## Code Style & Conventions

### Go Code Style
- Follow standard Go conventions (gofmt, golint)
- Use structured logging with descriptive messages
- Package-level documentation for public APIs
- Error handling with context preservation

### Bazel Conventions
- Use `load()` statements at the top of BUILD files
- Consistent target naming (lowercase with underscores)
- Visibility declarations where appropriate
- Comment directives for Gazelle configuration

### Directory Patterns
```
package/
├── BUILD.bazel           # Bazel build rules
├── CMakeLists.txt       # CMake project definition (if applicable)
├── *.go                 # Go source files
└── *_test.go           # Go test files
```

## Bazel-Centric Approach

### Rule Generation Strategy
1. **CMake File API First**: Use modern CMake File API for robust parsing when available
2. **Regex Fallback**: Fall back to regex-based parsing for compatibility
3. **Bazel Labels**: Generate proper Bazel labels for dependencies and sources
4. **External Repository Support**: Handle external CMake projects via `@repo//` labels

### Key Gazelle Integration Points

#### Language Interface Implementation
```go
// Required methods for language.Language interface
Name() string                    // Returns "cmake"
Kinds() map[string]rule.KindInfo // cc_library, cc_binary, cc_test
GenerateRules() GenerateResult   // Main rule generation logic
Configure()                      // Handle directive configuration
```

#### Directive Handling
- `gazelle:cmake @repo//:srcs` - Process external CMake repository
- `gazelle:cmake_executable /path/to/cmake` - Configure CMake binary
- `gazelle:cmake_define KEY VALUE` - Pass CMake definitions

#### Build Rule Generation
```starlark
cc_library(
    name = "target_name",
    srcs = ["src/file.cpp"],
    hdrs = ["include/header.h"], 
    includes = ["include"],
    deps = [":dependency"],
)
```

## Development Workflow

### Building the Plugin
```bash
bazel build //gazelle:gazelle-foreign-cc
```

### Testing with Examples
```bash
cd examples
bazel run //:gazelle          # Generate BUILD rules
bazel build //...             # Build all examples
```

### Module Dependencies
After modifying MODULE.bazel:
```bash
bazel mod tidy                # Update MODULE.bazel.lock
```

## Testing Strategy

### Test Data Structure
- `testdata/` contains various CMake project patterns
- Each test project represents different CMake constructs
- Integration tests verify end-to-end rule generation

### CI Requirements
- Ensure MODULE.bazel.lock stays synchronized
- Build success for plugin and examples
- Go tests pass for all packages

## Implementation Notes

### Current Status
- ✅ Basic C++ rule generation by convention
- ✅ CMake directive handling
- ✅ External repository support
- 🚧 Full CMakeLists.txt parsing (File API + regex)
- 🚧 Sophisticated dependency resolution
- 🚧 Include scanning for automatic deps

### Future Development
- Enhanced CMake parsing coverage
- Cross-language dependency resolution
- Performance optimization for large projects
- Advanced CMake feature support (conditionals, generators)

## Quick Reference

### Common Commands
```bash
# Build the plugin
bazel build //gazelle:gazelle-foreign-cc

# Run on a directory
bazel run //:gazelle -- path/to/cmake/project

# Update dependencies
bazel mod tidy

# Run tests
bazel test //...
```

### Important Files to Modify
- `language/cmake.go` - Main language logic
- `gazelle/config.go` - Configuration and directives  
- `language/cmake_api.go` - CMake File API integration
- `examples/` - Add new example projects
- `testdata/` - Add test cases for new features