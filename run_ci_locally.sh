#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "=== Building Gazelle Foreign CC Plugin ==="
bazel build //gazelle:gazelle-foreign-cc

echo ""
echo "=== Setting up Gazelle Plugin Path ==="
# Determine the correct path to the built plugin
# Based on the target //gazelle:gazelle-foreign-cc, the output is typically bazel-bin/gazelle/gazelle-foreign-cc
PLUGIN_BINARY_PATH="bazel-bin/gazelle/gazelle-foreign-cc"

if [ ! -f "$PLUGIN_BINARY_PATH" ]; then
  echo "ERROR: Plugin binary not found at $PLUGIN_BINARY_PATH"
  # Attempt to find it, as the exact path can sometimes vary with Bazel versions or configurations
  # This is a common alternative for non-main repositories or specific target naming.
  ALTERNATIVE_PATH="bazel-bin/gazelle/gazelle-foreign-cc_foreign_cc_/gazelle-foreign-cc"
  if [ -f "$ALTERNATIVE_PATH" ]; then
    PLUGIN_BINARY_PATH=$ALTERNATIVE_PATH
    echo "Found plugin at alternative path: $PLUGIN_BINARY_PATH"
  else
    echo "Also checked $ALTERNATIVE_PATH, not found. Please verify the Bazel output path."
    exit 1
  fi
fi

PLUGIN_PATH_DIR=$(mktemp -d)
echo "Temporary directory for plugin: $PLUGIN_PATH_DIR"
cp "$PLUGIN_BINARY_PATH" "$PLUGIN_PATH_DIR/gazelle-cmake"
export PATH="$PLUGIN_PATH_DIR:$PATH"
echo "gazelle-cmake added to PATH from $PLUGIN_PATH_DIR"

echo ""
echo "=== Running Gazelle on testdata/simple_cc_project ==="
# Set BUILD_WORKSPACE_DIRECTORY as Gazelle might need it (though often not for direct CLI runs)
export BUILD_WORKSPACE_DIRECTORY="$(pwd)" 
bazel run @gazelle//:gazelle -- testdata/simple_cc_project

echo ""
echo "=== Checking for generated BUILD.bazel file ==="
if [ -f "testdata/simple_cc_project/BUILD.bazel" ]; then
  echo "BUILD.bazel found in testdata/simple_cc_project."
  echo "--- Contents of testdata/simple_cc_project/BUILD.bazel ---"
  cat testdata/simple_cc_project/BUILD.bazel
  echo "--- End of BUILD.bazel ---"
else
  echo "ERROR: BUILD.bazel NOT found in testdata/simple_cc_project."
  # Cleanup before exiting with error
  echo "Cleaning up temporary directory: $PLUGIN_PATH_DIR"
  rm -rf "$PLUGIN_PATH_DIR"
  exit 1
fi

echo ""
echo "=== Cleaning up ==="
echo "Removing temporary directory: $PLUGIN_PATH_DIR"
rm -rf "$PLUGIN_PATH_DIR"

echo ""
echo "=== Local CI script finished successfully! ==="
