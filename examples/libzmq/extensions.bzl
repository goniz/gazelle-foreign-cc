"""Module extension for non-module dependencies."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def _non_module_deps_impl(module_ctx):
    """Implementation of the non_module_deps extension."""
    
    # Fetch libzmq
    http_archive(
        name = "libzmq",
        url = "https://github.com/zeromq/libzmq/releases/download/v4.3.5/zeromq-4.3.5.tar.gz",
        sha256 = "6653ef5910f17954861fe72332e68b03ca6e4d9c7160eb3a8de5a5a913bfab43",
        strip_prefix = "zeromq-4.3.5",
        build_file_content = """
# Placeholder BUILD file for libzmq
# This will be replaced by gazelle-generated content

filegroup(
    name = "all_files",
    srcs = glob(["**/*"]),
    visibility = ["//visibility:public"],
)

# CMakeLists.txt should be processed by gazelle
exports_files(["CMakeLists.txt"])

# Basic library target that gazelle should replace
cc_library(
    name = "zmq",
    visibility = ["//visibility:public"],
)
""",
    )

non_module_deps = module_extension(
    implementation = _non_module_deps_impl,
)