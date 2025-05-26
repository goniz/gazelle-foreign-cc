load("@rules_foreign_cc//foreign_cc:defs.bzl", "cmake")

filegroup(
    name = "all_srcs",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

cmake(
    name = "libzmq",
    cache_entries = {
        "CMAKE_BUILD_TYPE": "Release",
        "ZMQ_BUILD_TESTS": "OFF",
        "ENABLE_CPACK": "OFF",
        "ENABLE_CURVE": "OFF",  # Disable curve encryption to avoid libsodium dependency
        "WITH_DOCS": "OFF",
        "BUILD_SHARED": "ON",
        "BUILD_STATIC": "ON",
    },
    lib_source = ":all_srcs",
    out_static_libs = ["libzmq.a"],
    out_shared_libs = ["libzmq.so"],
    visibility = ["//visibility:public"],
)