#!/bin/bash

# GitHub Copilot Coding Agent setup script for gazelle-foreign-cc
# This script installs the necessary tools and dependencies for working on
# the Bazel Gazelle plugin for C/C++ projects using CMake.

set -e

echo "Setting up development environment for gazelle-foreign-cc..."

# Update package lists
apt-get update

# Install essential build tools
apt-get install -y \
    build-essential \
    curl \
    git \
    unzip \
    wget

# Install CMake (required for processing C++ projects)
apt-get install -y cmake

# Install Go (required for the Gazelle plugin development)
# The project requires Go 1.22+ as specified in go.mod
GO_VERSION="1.23.9"
if ! command -v go &> /dev/null || [[ $(go version | grep -o 'go[0-9.]*' | head -1) < "go1.22" ]]; then
    echo "Installing Go ${GO_VERSION}..."
    wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # Add Go to PATH
    export PATH="/usr/local/go/bin:$PATH"
    echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
fi

# Install Bazelisk (Bazel version manager - preferred over direct Bazel installation)
if ! command -v bazel &> /dev/null; then
    echo "Installing Bazelisk..."
    npm install -g @bazel/bazelisk
    
    # Ensure bazelisk is available as 'bazel'
    if [ ! -f /usr/local/bin/bazel ] && [ -f /usr/local/bin/bazelisk ]; then
        ln -s /usr/local/bin/bazelisk /usr/local/bin/bazel
    fi
fi

# Verify installations
echo "Verifying installations..."
echo "Go version: $(go version)"
echo "CMake version: $(cmake --version | head -1)"
echo "Bazel version: $(bazel version 2>/dev/null | head -1 || echo 'Bazel available via bazelisk')"

echo "Setup complete! You can now work on the gazelle-foreign-cc project."
echo ""
echo "To build the project:"
echo "  bazel build //gazelle:gazelle-foreign-cc"
echo ""
echo "To run the local CI script:"
echo "  ./run_ci_locally.sh"