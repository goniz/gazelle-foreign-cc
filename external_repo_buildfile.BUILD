package(default_visibility = ["//visibility:public"])

# Create platform.hpp for libzmq build
genrule(
    name = "platform_hpp",
    srcs = ["@gazelle-foreign-cc//:libzmq_platform.hpp"],
    outs = ["src/platform.hpp"],
    cmd = "cp $< $@",
)

filegroup(
	name = "srcs",
	srcs = glob(["**"]),
)

exports_files(glob(["**"]))