# Simple Hello World Example

This is a simple example demonstrating how to use the gazelle-foreign-cc plugin with a basic C++ project.

## Building

To build this example using Bazel and Gazelle:

```bash
# Generate BUILD.bazel file from CMakeLists.txt
bazel run //gazelle:gazelle-foreign-cc -- path/to/this/directory

# Build the project
bazel build //examples/simple_hello:hello
```

## Running

To run the resulting binary:

```bash
bazel run //examples/simple_hello:hello
```