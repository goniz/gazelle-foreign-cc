# libcurl Example

This example demonstrates how to use libcurl with Bazel and gazelle-foreign-cc.

## Overview

This example includes:
- A thirdparty package (`//thirdparty/libcurl`) that uses gazelle directives to build libcurl from source
- Two example applications that use libcurl:
  - `example_http_get`: Performs a simple HTTP GET request
  - `example_download`: Downloads a file with progress reporting

## Building

To build the examples:

```bash
cd examples
bazel build //libcurl:example_http_get
bazel build //libcurl:example_download
```

Or build all targets:

```bash
cd examples
bazel build //libcurl:all
```

## Running the Examples

### HTTP GET Example

The HTTP GET example fetches a URL and displays the response:

```bash
# Fetch the default URL (httpbin.org/get)
bazel run //libcurl:example_http_get

# Fetch a custom URL
bazel run //libcurl:example_http_get -- https://api.github.com
```

### File Download Example

The download example downloads a file from a URL with progress reporting:

```bash
# Download a file
bazel run //libcurl:example_download -- https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf output.pdf
```

## How It Works

### Thirdparty Package

The thirdparty package (`//thirdparty/libcurl/BUILD.bazel`) uses gazelle directives to configure the CMake build:

```starlark
# gazelle:cmake_source @curl
# gazelle:cmake_define BUILD_CURL_EXE OFF
# gazelle:cmake_define BUILD_TESTING OFF
# gazelle:cmake_define CURL_DISABLE_LDAP ON
# gazelle:cmake_define CURL_DISABLE_LDAPS ON
# gazelle:cmake_define CMAKE_USE_OPENSSL ON
```

These directives:
- Point to the external curl repository (`@curl`)
- Disable building the curl executable (we only need the library)
- Disable tests
- Disable LDAP support (simplifies dependencies)
- Enable OpenSSL support for HTTPS

### Dependencies

The example automatically links against system libraries:
- OpenSSL (`-lssl -lcrypto`) for HTTPS support
- zlib (`-lz`) for compression support

## Notes

- The thirdparty package includes platform-specific configuration (Linux support)
- The curl_config.h file is generated using `cmake_configure_file` rule
- The examples demonstrate both simple HTTP requests and file downloads with progress