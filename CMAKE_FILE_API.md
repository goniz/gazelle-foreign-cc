# CMake File API Integration for Bazel Gazelle

This document describes the CMake File API integration implemented for the Bazel Gazelle foreign-cc plugin.

## Overview

The CMake File API integration replaces regex-based parsing of CMakeLists.txt files with a more robust and accurate approach using CMake's built-in File API. This provides better dependency resolution, target information, and build configuration details.

## Implementation

### Core Components

1. **cmake_api.go** - Complete CMake File API implementation
   - JSON structures matching CMake File API v1 schema
   - API query generation and response parsing
   - Target, source, and dependency extraction

2. **generate.go** - Integration with Gazelle generation
   - Primary File API usage with fallback to regex parsing
   - Enhanced target processing and rule generation

3. **resolve.go** - Improved dependency resolution
   - Better cross-target dependency handling using File API data

### Key Features

- **Accurate Target Parsing**: Extracts executables, libraries (static/shared), and their properties
- **Source File Detection**: Identifies C/C++ source files, headers, and their relationships
- **Dependency Resolution**: Maps target dependencies and linked libraries
- **Include Directory Handling**: Processes both global and target-specific include paths
- **Subdirectory Support**: Handles complex projects with multiple CMakeLists.txt files
- **Error Handling**: Graceful fallback to regex parsing when File API fails

## Usage

The integration is automatically used when generating Bazel rules for CMake projects:

```bash
# Generate rules for a CMake project
bazel run //:gazelle

# Or use the standalone test program
bazel run //gazelle:gazelle-foreign-cc -- --source_dir=/path/to/cmake/project
```

## CMake File API Details

### Query Generation

The implementation creates a File API query requesting:
- Codemodel information (targets, sources, dependencies)
- Cache variables
- CMake configuration details

### Response Processing

Parses the JSON response to extract:
- Target definitions (name, type, sources)
- Compile information (include directories, definitions)
- Link information (libraries, dependencies)
- Source file groups and properties

### Supported Target Types

- **Executables**: Converted to `cc_binary` rules
- **Static Libraries**: Converted to `cc_library` rules
- **Shared Libraries**: Converted to `cc_library` rules with appropriate linkage
- **Interface Libraries**: Header-only libraries

## Testing

The implementation includes comprehensive tests:

### Test Projects

1. **Simple Project**: Basic executable and library
2. **Complex Project**: Multiple targets, subdirectories, dependencies
3. **Invalid Project**: Tests fallback behavior

### Running Tests

```bash
# Run all tests
./test_cmake_api.sh

# Test specific project
cd testdata/simple_cc_project
bazel run //gazelle:gazelle-foreign-cc
```

### Expected Output

```
Successfully parsed X targets using CMake File API:
  Target: app (type: executable)
    Sources: [main.cc]
    Headers: []
    Include dirs: []
    Linked libs: [my_lib]
  Target: my_lib (type: library)
    Sources: [lib.cc]
    Headers: [lib.h]
    Include dirs: []
    Linked libs: []
```

## Error Handling

### CMake Configuration Failures

When CMake configuration fails (invalid syntax, missing dependencies), the system:
1. Logs the CMake error
2. Falls back to regex-based parsing
3. Continues with best-effort rule generation

### File API Unavailability

If CMake File API is not available (older CMake versions):
1. Automatically detects the limitation
2. Uses regex parsing as primary method
3. Logs the fallback for debugging

## Benefits Over Regex Parsing

1. **Accuracy**: No parsing ambiguities or edge cases
2. **Completeness**: Access to all CMake target information
3. **Maintainability**: No complex regex patterns to maintain
4. **Robustness**: Handles complex CMake constructs correctly
5. **Future-proof**: Leverages CMake's official API

## Limitations

1. **CMake Version**: Requires CMake 3.14+ for full File API support
2. **Configuration**: Requires successful CMake configuration
3. **Build Directory**: Creates temporary build directories for API queries

## File Structure

```
gazelle/
├── cmake_api.go          # CMake File API implementation
├── generate.go           # Rule generation with File API integration
├── resolve.go            # Dependency resolution
├── main.go              # Test program
└── BUILD.bazel          # Build configuration

testdata/
├── simple_cc_project/    # Basic test case
├── complex_cc_project/   # Advanced test case
└── invalid_cmake_project/ # Error handling test
```

## Future Enhancements

1. **Caching**: Cache File API responses for performance
2. **Incremental Updates**: Only re-query when CMakeLists.txt changes
3. **Advanced Features**: Support for custom properties, generators
4. **Integration**: Better integration with Bazel workspace rules