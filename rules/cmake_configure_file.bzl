"""CMake configure_file rule for Bazel that runs cmake configure and copies generated files."""

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
    for key, value in ctx.attr.defines.items():
        define_args.append('-D%s=%s' % (key, value))
    
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
            "-S", actual_source_dir,
            "-B", build_dir.path,
        ] + define_args,
        mnemonic = "CMakeConfigure",
        progress_message = "Running cmake configure",
        use_default_shell_env = True,
    )
    
    # Copy the generated file from the cmake build directory
    generated_file_path = ctx.attr.generated_file_path
    if not generated_file_path:
        fail("generated_file_path attribute is required")
    
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
    
    return [DefaultInfo(files = depset([output_file]))]

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
            allow_single_file = True,
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
            mandatory = True,
            doc = "Path to the generated file relative to cmake build directory",
        ),
    },
    doc = "Runs cmake configure and copies the generated file.",
)