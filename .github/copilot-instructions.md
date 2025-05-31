# GitHub Copilot Instructions for gazelle-foreign-cc

## Project Overview

This repository contains a Bazel Gazelle plugin that generates C++ BUILD files (`cc_library`, `cc_binary`, `cc_test`) from existing CMake projects. The plugin is written in Go and integrates with Bazel's module system (bzlmod).

## Tech Stack

- **Language**: Go 1.22.9+
- **Build System**: Bazel with bzlmod
- **Target**: C++ projects with CMake
- **Dependencies**: 
  - `github.com/bazelbuild/bazel-gazelle v0.43.0`
  - `github.com/bazelbuild/buildtools`

## Repository Structure

```
├── gazelle/           # Core Gazelle plugin implementation
│   ├── config.go      # Configuration handling and directives
│   ├── generate.go    # Rule generation logic
│   ├── resolve.go     # Dependency resolution
│   ├── plugin.go      # Plugin registration
│   └── main.go        # Entry point
├── language/          # Language interface implementation
│   ├── cmake.go       # CMake language extension
│   └── cmake_api.go   # CMake File API integration
├── examples/          # Example projects
├── testdata/          # Test data for unit tests
└── .openhands/        # OpenHands microagents configuration
```

## Key Files and Their Purpose

- **`gazelle/config.go`**: Handles Gazelle directives like `gazelle:cmake_executable`
- **`gazelle/generate.go`**: Parses CMakeLists.txt and generates BUILD rules
- **`gazelle/resolve.go`**: Resolves C++ dependencies via #include analysis and linked libraries
- **`language/cmake.go`**: Implements Gazelle's Language interface for CMake
- **`language/cmake_api.go`**: Uses CMake File API for sophisticated parsing

## Build and Development Guidelines

### Building the Plugin
```bash
bazel build //gazelle:gazelle-foreign-cc
```

### Module Dependencies
After making changes to `MODULE.bazel`, always run:
```bash
bazel mod tidy
```
The CI will fail if `MODULE.bazel.lock` is not synchronized with `MODULE.bazel`.

### Testing
```bash
bazel test //...
```

Test files follow the pattern `*_test.go` and use standard Go testing practices.

## Code Style and Conventions

### Go Conventions
- Follow standard Go naming conventions
- Use `log.Printf` for debug output (seen throughout the codebase)
- Package imports should be grouped: standard library, external dependencies, internal packages
- Use meaningful variable names (e.g., `cmakeFilePath`, `cmakeTargets`)

### Gazelle Plugin Patterns
- Implement the `language.Language` interface in `language/cmake.go`
- Configuration should be handled through the `CMakeConfig` struct
- Use `rule.NewRule()` to create new BUILD rules
- Store metadata in private attributes using `SetPrivateAttr()` for use in resolve phase

### Error Handling
- Always check for file existence before parsing: `os.Stat(cmakeFilePath)`
- Log informative messages for debugging: `log.Printf("CMake File API failed for %s: %v", args.Rel, err)`
- Gracefully fall back to simpler parsing when sophisticated methods fail

### BUILD Rule Generation
- Generate `cc_library` for libraries with `add_library()`
- Generate `cc_binary` for executables with `add_executable()`
- Generate `cc_test` for test targets
- Set `srcs` attribute for source files, `hdrs` for headers
- Use `SetPrivateAttr()` to store CMake metadata for dependency resolution

## Important Rules and Constraints

### Development Rules
1. **Always ensure Bazel targets build successfully after changes**
2. **Never extend scope without user approval**
3. **Never push directly to main branch**
4. **Always push code to feature branches**

### Dependencies and Resolution
- The plugin handles two types of dependencies:
  1. Linked libraries from CMake (`target_link_libraries`)
  2. Header dependencies from `#include` statements
- Use regex pattern `#include\s*([<"])([^>"]+)([>"])` for include parsing
- Dependency resolution happens in the `Resolve` phase after rule generation

### CMake Integration
- Primary method: CMake File API (modern, structured)
- Fallback method: Regex parsing of CMakeLists.txt
- Support for basic CMake directives like `add_library`, `add_executable`, `target_link_libraries`
- Handle include directories and compiler options

### Testing Guidelines
- Create test data in `testdata/` directory
- Use `createMockGenerateArgs()` helper for test setup
- Test both successful parsing and fallback scenarios
- Verify both generated rules and empty rules for Gazelle's update mechanism

## Current Limitations

- Full CMake `CMakeLists.txt` parsing is not yet complete
- Sophisticated dependency resolution is still under development
- Only basic C++ source file recognition is implemented
- CMake File API integration is partially implemented

## Key Gazelle Concepts

- **GenerateRules**: Called to generate new BUILD rules from source files
- **Resolve**: Called to resolve dependencies between rules
- **Configure**: Called to handle directives from BUILD files
- **UpdateRules**: Called to update existing rules
- **Language Interface**: Must implement all required methods for Gazelle integration

## Examples and Usage

See the `examples/` directory for sample projects. The plugin is typically run as:
```bash
bazel run //:gazelle -- path/to/your/cmake_project
```

## Notes for Contributors

- This project is under active development
- Focus on robustness and error handling
- Maintain compatibility with Bazel's bzlmod system
- Follow existing logging patterns for debugging
- Test with various CMake project structures