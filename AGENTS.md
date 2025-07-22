# AGENTS.md - Development Commands & Guidelines

## Build/Test Commands
```bash
bazel build //...                    # Build all targets
bazel test //...                     # Run all tests
bazel test //gazelle:gazelle_tests   # Run gazelle package tests
bazel test //language:language_tests # Run language package tests
bazel build //gazelle:gazelle-foreign-cc # Build main plugin
cd examples && bazel run //:gazelle && bazel build //... # Test with examples
```

## Development Cycle
1. `cd examples` - Move to examples directory
2. `bazel run //:gazelle` - Run gazelle to generate BUILD files
3. `bazel build //libzmq:all` - Build specific example to test
4. Fix gazelle plugin if build fails, repeat until issue is resolved

## Code Style Guidelines
- **Go**: Use standard Go conventions (gofmt, golint), structured logging with `log` package
- **Imports**: Group stdlib, third-party, local packages with blank lines between groups
- **Naming**: Use camelCase for Go, snake_case for Bazel targets, descriptive variable names
- **Types**: Prefer explicit types, use interfaces for testability (e.g., `language.Language`)
- **Error Handling**: Return errors explicitly, use `fmt.Errorf` for wrapping, log errors before returning
- **Bazel**: Place `load()` statements at top, use lowercase_underscore naming, declare visibility
- **BUILD Files**: NEVER edit auto-generated BUILD files except for adding gazelle directives
- **Tooling**: NEVER use Go tooling directly (go build, go test, etc.), always use Bazel
- **Git**: NEVER push directly to main branch, always use feature branches and PRs
- **Testing**: Before pushing to git, always test both the Go tests AND at least one example

## Project Structure
- `gazelle/` - Plugin configuration and core logic
- `language/` - Gazelle language implementation (main interface)
- `common/` - Shared types and utilities
- `testdata/` - CMake test patterns for unit tests
- `examples/` - Integration test projects (separate Bazel module)