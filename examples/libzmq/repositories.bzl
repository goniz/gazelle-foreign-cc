# Repository rule to fetch libzmq source
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def libzmq_repositories():
    """Fetch libzmq source code."""
    
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
""",
    )