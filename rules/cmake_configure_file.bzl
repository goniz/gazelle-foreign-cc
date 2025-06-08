"""CMake configure_file rule for Bazel that uses cmake binary for file generation."""

def _cmake_configure_file_impl(ctx):
    """Implementation of cmake_configure_file rule that calls cmake binary."""
    input_file = ctx.file.src
    output_file = ctx.outputs.out
    cmake_binary = ctx.attr.cmake_binary
    
    # Create a temporary CMakeLists.txt file that performs the configure_file operation
    cmake_script = ctx.actions.declare_file(ctx.label.name + "_configure.cmake")
    
    # Build cmake -D arguments for variable definitions
    define_args = []
    for key, value in ctx.attr.defines.items():
        define_args.append("-D%s=%s" % (key, value))
    
    # Create CMake script content
    script_content = """
# Set variables
"""
    for key, value in ctx.attr.defines.items():
        script_content += 'set(%s "%s")\n' % (key, value)
    
    script_content += '''
# Configure the file
configure_file("${INPUT_FILE}" "${OUTPUT_FILE}" @ONLY)
'''
    
    # Write the CMake script
    ctx.actions.write(
        output = cmake_script,
        content = script_content,
    )
    
    # Run cmake in script mode to perform the configure_file operation
    ctx.actions.run(
        inputs = [input_file, cmake_script],
        outputs = [output_file],
        executable = cmake_binary,
        arguments = [
            "-DINPUT_FILE=" + input_file.path,
            "-DOUTPUT_FILE=" + output_file.path,
        ] + define_args + [
            "-P", cmake_script.path,
        ],
        mnemonic = "CMakeConfigureFile",
        progress_message = "Configuring %s with cmake" % output_file.short_path,
    )
    
    return [DefaultInfo(files = depset([output_file]))]

cmake_configure_file = rule(
    implementation = _cmake_configure_file_impl,
    attrs = {
        "src": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The input template file to configure",
        ),
        "out": attr.output(
            mandatory = True,
            doc = "The output file to generate",
        ),
        "defines": attr.string_dict(
            default = {},
            doc = "Dictionary of variable definitions for substitution",
        ),
        "cmake_binary": attr.label(
            default = "@cmake//:cmake",
            executable = True,
            cfg = "exec",
            doc = "The cmake binary to use for file configuration",
        ),
    },
    doc = "Configures a file using cmake binary, similar to CMake's configure_file command.",
)