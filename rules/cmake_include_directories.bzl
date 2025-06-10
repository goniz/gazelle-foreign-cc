"""Generic rule for building external CMake projects with proper include paths."""

def _cmake_include_directories_impl(ctx):
    """Implementation of cmake_include_directories rule."""

    # Get all source files from the external sources
    srcs_depset = ctx.attr.srcs[DefaultInfo].files
    src_files = []
    hdr_files = []

    # Separate source and header files
    for file in srcs_depset.to_list():
        if file.extension in ["cpp", "c", "cc", "cxx"]:
            src_files.append(file)
        elif file.extension in ["hpp", "h", "hxx", "hh"]:
            hdr_files.append(file)

    # Build include paths - start with the external repo root for relative paths
    include_paths = [ctx.attr.srcs.label.workspace_root] if ctx.attr.srcs.label.workspace_root else []

    # Add user-specified include directories relative to the source repo
    for include_dir in ctx.attr.includes:
        if ctx.attr.srcs.label.workspace_root:
            include_paths.append(ctx.attr.srcs.label.workspace_root + "/" + include_dir)
        else:
            include_paths.append(include_dir)

    # Create compilation context with proper includes
    compilation_context = cc_common.create_compilation_context(
        includes = depset(include_paths),
        headers = depset(hdr_files + ctx.files.additional_hdrs),
        defines = depset(ctx.attr.defines),
    )

    # Create linking context (empty for now, will be populated by cc_library)
    linking_context = cc_common.create_linking_context(linker_inputs = depset([]))

    # Return CCInfo provider
    return [
        DefaultInfo(files = depset(src_files + hdr_files + ctx.files.additional_hdrs)),
        CcInfo(
            compilation_context = compilation_context,
            linking_context = linking_context,
        ),
    ]

cmake_include_directories = rule(
    implementation = _cmake_include_directories_impl,
    attrs = {
        "srcs": attr.label(
            doc = "The external CMake sources filegroup (e.g., @repo//:srcs)",
            mandatory = True,
        ),
        "includes": attr.string_list(
            doc = "Include directories relative to the source repository root",
            default = ["include", ".cmake-build"],
        ),
        "additional_hdrs": attr.label_list(
            doc = "Additional header files to include",
            allow_files = [".h", ".hpp", ".hxx", ".hh"],
            default = [],
        ),
        "defines": attr.string_list(
            doc = "Preprocessor defines",
            default = [],
        ),
    },
    fragments = ["cpp"],
    provides = [CcInfo],
    doc = "Provides proper include directories for external CMake sources with relative includes",
)
