#!/bin/bash

set -e
set -o pipefail

cd examples
bazel run :gazelle
bazel build //thirdparty/libzmq
bazel build //libzmq:all