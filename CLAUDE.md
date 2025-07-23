# CLAUDE.md - Development Guide for gazelle-foreign-cc

**gazelle-foreign-cc** is a Bazel Gazelle plugin that generates C++ BUILD rules (`cc_library`, `cc_binary`, `cc_test`) from CMake projects, bridging CMake and Bazel build systems.

## Core Functionality
- Parses CMakeLists.txt files using CMake File API exclusively
- Generates Bazel `cc_*` rules automatically
- Supports external CMake dependencies via `gazelle:cmake` directives

## Project Structure
```
gazelle-foreign-cc/
├── gazelle/               # Plugin configuration and core logic
│   ├── config.go         # CMake directives (`gazelle:cmake_executable`, `gazelle:cmake`)
│   ├── generate.go       # Deprecated (regex fallback removed)
│   └── plugin.go         # Plugin registration
├── language/             # Gazelle language implementation
│   ├── cmake.go          # Main language.Language interface
│   └── cmake_api.go      # CMake File API integration
├── examples/             # Test projects (separate Bazel module)
└── testdata/             # CMake test patterns
```

## Key Files for Development
- `language/cmake.go` - Main language logic, implements Gazelle interface
- `gazelle/config.go` - Configuration and directive handling
- `language/cmake_api.go` - CMake File API integration
- `examples/` - Add example projects for testing
- `testdata/` - Add test cases for new CMake patterns

## Bazel Configuration

### Root MODULE.bazel Dependencies
```starlark
bazel_dep(name = "rules_go", version = "0.54.1")     # Go toolchain
bazel_dep(name = "gazelle", version = "0.43.0")     # Gazelle framework  
bazel_dep(name = "rules_cc", version = "0.1.1")     # C++ rules
```

### Generated Rule Pattern
```starlark
cc_library(
    name = "target_name",
    srcs = ["src/file.cpp"],
    hdrs = ["include/header.h"],
    includes = ["include"],
    deps = [":dependency"],
)
```

## Development Commands
```bash
# Build plugin
bazel build //gazelle:gazelle-foreign-cc

# Test with examples
cd examples && bazel run //:gazelle && bazel build //...

# Update dependencies  
bazel mod tidy

# Run tests
bazel test //...
```

## Gazelle Integration

### Required Language Interface
```go
Name() string                    // Returns "cmake"
Kinds() map[string]rule.KindInfo // cc_library, cc_binary, cc_test
GenerateRules() GenerateResult   // Main generation logic
Configure()                      // Handle directives
```

### CMake Directives
- `gazelle:cmake @repo//:srcs` - Process external CMake repo
- `gazelle:cmake_executable /path/to/cmake` - Set CMake binary
- `gazelle:cmake_define KEY VALUE` - Pass CMake definitions

## Rule Generation Strategy
1. **CMake File API**: Modern CMake parsing (exclusive method)
2. **Bazel Labels**: Generate proper `@repo//target` references
3. **External Support**: Handle external CMake projects
4. **Error Handling**: Fail if CMake File API is unavailable

## Implementation Status
- ✅ Basic C++ rule generation, directive handling, external repos
- 🚧 Full CMake parsing, dependency resolution, include scanning

## Code Conventions
- **Go**: Standard conventions (gofmt, golint), structured logging
- **Bazel**: `load()` at top, lowercase_underscore naming, visibility declarations