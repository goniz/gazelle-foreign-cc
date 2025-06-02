# Examples

This directory contains example projects demonstrating how to use the gazelle-foreign-cc plugin.

This examples directory is a separate Bazel module that references the main gazelle-foreign-cc module using a local path override, allowing it to fetch real-world CMake projects and demonstrate the plugin's capabilities.

## Available Examples

### Simple Hello World

A basic C++ "Hello, World!" example that demonstrates how to use the plugin with a simple project.

- [simple_hello/README.md](simple_hello/README.md)

### LibZMQ (ZeroMQ)

A real-world example using libzmq (ZeroMQ), a popular high-performance messaging library that uses CMake. This example demonstrates how the gazelle plugin can parse complex CMake projects and generate appropriate Bazel BUILD files.

- [libzmq/README.md](libzmq/README.md)

## Building Examples

To build all examples:

```bash
cd examples
bazel build //...
```

To run gazelle on the examples to generate/update BUILD files:

```bash
cd examples
bazel run //:gazelle
```

## Adding a New Example

To add a new example:

1. Create a new directory with the name of your example
2. Add necessary files (C/C++ source, CMakeLists.txt, BUILD.bazel, etc.)
3. If using external dependencies, add them to the MODULE.bazel or create appropriate repository rules
4. Document how to build and run the example
5. Update this README to include your new example

## Module Structure

The examples module references the main gazelle-foreign-cc module using a local path override, which allows it to:

- Use the latest version of the plugin during development
- Demonstrate real-world usage patterns
- Test the plugin against actual CMake projects
- Serve as integration tests for the plugin