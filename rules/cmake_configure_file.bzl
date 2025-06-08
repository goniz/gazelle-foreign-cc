"""CMake configure_file rule for Bazel."""

def _cmake_configure_file_impl(ctx):
    """Implementation of cmake_configure_file rule."""
    input_file = ctx.file.src
    output_file = ctx.outputs.out
    
    # Create substitution arguments
    substitutions = {}
    for key, value in ctx.attr.defines.items():
        # Handle both @VAR@ and ${VAR} formats
        substitutions["@%s@" % key] = value
        substitutions["${%s}" % key] = value
    
    # Create the configure action
    ctx.actions.expand_template(
        template = input_file,
        output = output_file,
        substitutions = substitutions,
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
    },
    doc = "Configures a file by performing variable substitution, similar to CMake's configure_file command.",
)