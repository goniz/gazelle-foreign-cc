"""CMake configure_file rule for Bazel that runs cmake configure and copies generated files."""

def _create_compilation_context(ctx, output_file):
    """Create a compilation context for the generated header file."""

    # Create a virtual header that makes the config.h accessible via relative paths
    # For external repositories that use "../config.h", we need to make our generated
    # config.h findable via that path. We do this by adding appropriate include directories.

    # Add the directory containing the output file
    genfiles_include = output_file.dirname

    # Also add a quote include directory that can resolve "../config.h"
    # from external repo src directories to our generated file
    quote_includes = [genfiles_include]

    compilation_context = cc_common.create_compilation_context(
        headers = depset([output_file]),
        includes = depset([genfiles_include]),
        quote_includes = depset(quote_includes),
    )

    return compilation_context

def _cmake_configure_file_impl(ctx):
    """Implementation of cmake_configure_file rule that runs cmake configure and copies the generated file."""
    output_file = ctx.outputs.out
    cmake_binary = ctx.executable.cmake_binary
    cmake_source_dir = ctx.attr.cmake_source_dir

    # Validate inputs
    if not output_file:
        fail("out attribute is required")
    if not cmake_binary:
        fail("cmake_binary attribute is required")
    if not cmake_source_dir:
        fail("cmake_source_dir attribute is required")

    # Create a temporary build directory
    build_dir = ctx.actions.declare_directory(ctx.label.name + "_cmake_build")

    # Build cmake -D arguments for variable definitions if any
    define_args = []

    # Add CMAKE_CURRENT_SOURCE_DIR internally
    define_args.append("-DCMAKE_CURRENT_SOURCE_DIR=%s" % cmake_source_dir)
    for key, value in ctx.attr.defines.items():
        # Skip CMAKE_CURRENT_SOURCE_DIR if user provided it - we set it automatically
        if key != "CMAKE_CURRENT_SOURCE_DIR":
            define_args.append("-D%s=%s" % (key, value))

    # Get source files and derive the actual source directory
    inputs = []
    actual_source_dir = "."
    if ctx.attr.cmake_source_files:
        inputs.extend(ctx.files.cmake_source_files)

        # Find CMakeLists.txt in the inputs to determine the source directory
        for file in ctx.files.cmake_source_files:
            if file.basename == "CMakeLists.txt":
                # Use the directory containing CMakeLists.txt
                actual_source_dir = file.dirname
                break

    # Run cmake configure to generate files
    ctx.actions.run(
        inputs = inputs,
        outputs = [build_dir],
        executable = cmake_binary,
        arguments = [
            "-S",
            actual_source_dir,
            "-B",
            build_dir.path,
        ] + define_args,
        mnemonic = "CMakeConfigure",
        progress_message = "Running cmake configure",
        use_default_shell_env = True,
    )

    # Copy the generated file from the cmake build directory
    generated_file_path = ctx.attr.generated_file_path
    if not generated_file_path:
        # Default to the output file short path if not specified
        generated_file_path = output_file.short_path

    # Use a simple shell command to copy the file
    ctx.actions.run_shell(
        inputs = [build_dir],
        outputs = [output_file],
        command = """
        if [ ! -f "{build_dir}/{generated_path}" ]; then
            echo "Generated file not found: {build_dir}/{generated_path}"
            echo "Available files in build directory:"
            find {build_dir} -type f | head -20 || true
            exit 1
        fi
        cp "{build_dir}/{generated_path}" "{output_file}"
        """.format(
            build_dir = build_dir.path,
            generated_path = generated_file_path,
            output_file = output_file.path,
        ),
        mnemonic = "CMakeCopyFile",
        progress_message = "Copying cmake generated file",
        use_default_shell_env = True,
    )

    # ---------------------------------------------------------------------
    # Extra convenience copy for headers included with "../config.h"
    # ---------------------------------------------------------------------
    # We want exactly two include-able paths for the generated header:
    #   1) .cmake-build/generated/config.h  (what CMake targets reference)
    #   2) .cmake-build/config.h            (so "../config.h" from src/ works)

    compilation_headers = [output_file]
    include_dirs = [output_file.dirname]  # e.g. .cmake-build/generated

    # If the output lives in .../generated/, create the sibling copy one level up.
    if output_file.short_path.startswith(".cmake-build/generated/"):
        parent_copy = ctx.actions.declare_file(".cmake-build/config.h")
        ctx.actions.run_shell(
            inputs = [output_file],
            outputs = [parent_copy],
            command = "cp {} {}".format(output_file.path, parent_copy.path),
            mnemonic = "ConfigHParentCopy",
            progress_message = "Creating .cmake-build/config.h copy",
        )

        compilation_headers.append(parent_copy)
        include_dirs.append(parent_copy.dirname)  # .cmake-build

    # Always expose <repo_root> and <repo_root>/src to catch "../config.h" look-ups.
    include_dirs.extend(["", "src"])

    compilation_context = cc_common.create_compilation_context(
        headers = depset(compilation_headers),
        includes = depset(include_dirs),
        quote_includes = depset(include_dirs),
    )

    return [DefaultInfo(files = depset(compilation_headers)), CcInfo(compilation_context = compilation_context)]

cmake_configure_file = rule(
    implementation = _cmake_configure_file_impl,
    attrs = {
        "out": attr.output(
            mandatory = True,
            doc = "The output file to generate",
        ),
        "defines": attr.string_dict(
            default = {},
            doc = "Dictionary of variable definitions for cmake",
        ),
        "cmake_binary": attr.label(
            executable = True,
            cfg = "exec",
            mandatory = True,
            doc = "The cmake binary to use",
        ),
        "cmake_source_dir": attr.string(
            mandatory = True,
            doc = "The directory containing CMakeLists.txt",
        ),
        "cmake_source_files": attr.label_list(
            allow_files = True,
            doc = "CMakeLists.txt and related files",
        ),
        "generated_file_path": attr.string(
            doc = "Path to the generated file relative to cmake build directory. Defaults to the basename of 'out' if not specified.",
        ),
    },
    doc = "Runs cmake configure and copies the generated file.",
)
