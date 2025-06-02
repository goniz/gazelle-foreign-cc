# LibZMQ Example

This example demonstrates how to use the gazelle-foreign-cc plugin with libzmq, a popular messaging library that uses CMake.

## Overview

LibZMQ (ZeroMQ) is a high-performance asynchronous messaging library that provides message queues, but unlike message-oriented middleware, a ZeroMQ system can run without a dedicated message broker.

## Building

To build this example:

```bash
cd examples/libzmq
bazel build //...
```

To run gazelle on this example:

```bash
cd examples
bazel run //:gazelle -- libzmq
```

## Files

- `BUILD.bazel` - Bazel build configuration
- `libzmq/` - Downloaded libzmq source code (fetched automatically)
- `example_client.cpp` - Simple example client using libzmq
- `example_server.cpp` - Simple example server using libzmq

## Expected Output

After running gazelle, BUILD files should be generated for the libzmq CMake targets, allowing you to build and use libzmq from Bazel.