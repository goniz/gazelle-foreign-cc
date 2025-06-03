#!/bin/bash

echo "=== CMake File API Integration Test ==="
echo

# Build the test program
echo "Building test program..."
cd /workspace/gazelle-foreign-cc
bazel build //cmd/gazelle-foreign-cc:gazelle-foreign-cc
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "✓ Build successful"
echo

# Test 1: Simple project
echo "Test 1: Simple CMake project"
echo "----------------------------"
cd testdata/simple_cc_project
/workspace/gazelle-foreign-cc/bazel-bin/cmd/gazelle-foreign-cc/gazelle-foreign-cc_/gazelle-foreign-cc
echo

# Test 2: Complex project with subdirectories
echo "Test 2: Complex CMake project with subdirectories"
echo "------------------------------------------------"
cd ../complex_cc_project
/workspace/gazelle-foreign-cc/bazel-bin/cmd/gazelle-foreign-cc/gazelle-foreign-cc_/gazelle-foreign-cc
echo

# Test 3: Project with CMake errors (fallback scenario)
echo "Test 3: Invalid CMake project (fallback scenario)"
echo "------------------------------------------------"
cd ../invalid_cmake_project
/workspace/gazelle-foreign-cc/bazel-bin/cmd/gazelle-foreign-cc/gazelle-foreign-cc_/gazelle-foreign-cc
echo

echo "=== All tests completed ==="