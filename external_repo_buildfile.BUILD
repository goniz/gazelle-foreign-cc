load("@rules_cc//cc:defs.bzl", "cc_library")

package(default_visibility = ["//visibility:public"])

# Create platform.hpp for libzmq build
genrule(
    name = "platform_hpp",
    srcs = ["@gazelle-foreign-cc//:libzmq_platform.hpp"],
    outs = ["src/platform.hpp"],
    cmd = "cp $< $@",
)

# Create a cc_library that provides platform.hpp as a header
cc_library(
    name = "platform_headers",
    hdrs = [":src/platform.hpp"],
    strip_include_prefix = "src",
)

filegroup(
	name = "srcs",
	srcs = glob(["**"]),
)

exports_files(glob(["**"]))