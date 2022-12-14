load("@bazel_skylib//lib:paths.bzl", "paths")
load("@rules_openapi//openapi:def.bzl", _openapi_generate = "openapi_generate")

def _openapi_generate_gen_impl(ctx):
    openapi_generator = ctx.toolchains["@rules_openapi//openapi:toolchain"].openapi_generator_cli
    arguments = [
        "-jar", paths.join("bazel-{}".format(ctx.workspace_name), openapi_generator.path),
        "generate",
        "-i", ctx.file.spec.short_path,
        "-g", ctx.attr.generator,
        "-o", paths.join(ctx.attr.out.label.package, ctx.attr.out.label.name),
    ]

    if ctx.attr.template_dir:
        arguments += ["-t", ctx.attr.template_dir.label.package]

    java_home = ctx.attr.java_runtime[java_common.JavaRuntimeInfo].java_home
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    ctx.actions.write(
        output = out_file,
        content = """\
#!/usr/bin/env bash
cd "$BUILD_WORKSPACE_DIRECTORY"
{cmd} {args}
        """.format(cmd = paths.join("bazel-{}".format(ctx.workspace_name), java_home, "bin/java"), args = " ".join(arguments)),
        is_executable = True,
    )
    runfiles = ctx.runfiles(files = ctx.files.java_runtime + [openapi_generator])
    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

_openapi_generate_gen = rule(
    _openapi_generate_gen_impl,
    executable = True,
    attrs = {
        "spec": attr.label(
            mandatory = True,
            allow_single_file = [".json", ".yaml"],
        ),
        "out": attr.label(mandatory = True, allow_files = True),
        "deps": attr.label_list(allow_files = True),
        "generator": attr.string(mandatory = True),
        "template_dir": attr.label(),
        "java_runtime": attr.label(
            default = Label("@bazel_tools//tools/jdk:current_host_java_runtime"),
            providers = [java_common.JavaRuntimeInfo],
        ),
    },
    toolchains = [
        "@rules_openapi//openapi:toolchain",
    ],
)

def openapi_generate(name, **kwargs):
    _openapi_generate(name = name, **kwargs)
    _openapi_generate_gen(name = name + "_gen", **kwargs)
