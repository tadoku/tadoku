load("@bazel_skylib//lib:shell.bzl", "shell")

_CODEGEN_TOOL = "//third_party/oapi-codegen/cmd/oapi-codegen"

def _oapi_codegen_impl(ctx):
    args = []
    if len(ctx.attr.generate) > 0:
        args += ["-generate", ",".join(ctx.attr.generate)]

    if ctx.attr.package != None:
        args += ["-package", ctx.attr.package]

    args.append(ctx.file.spec.path)
    ctx.actions.run_shell(
        outputs = [ctx.outputs.out],
        inputs = [ctx.file.spec],
        tools = [ctx.file.oapi_codegen_tool],
        command = """{cmd} {args} > {out}""".format(
            cmd = "$(pwd)/" + ctx.file.oapi_codegen_tool.path,
            args = " ".join(args),
            out = ctx.outputs.out.path,
        ),
    )

_oapi_codegen = rule(
    _oapi_codegen_impl,
    attrs = {
        "spec": attr.label(
            mandatory = True,
            allow_single_file = [".json", ".yaml"],
        ),
        "out": attr.output(mandatory = True),
        "package": attr.string(),
        "generate": attr.string_list(),
        "oapi_codegen_tool": attr.label(
            default = Label(_CODEGEN_TOOL),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
    },
)

def _oapi_codegen_gen_impl(ctx):
    args = ["-o", ctx.file.out.short_path]
    if len(ctx.attr.generate) > 0:
        args += ["-generate", ",".join(ctx.attr.generate)]

    if ctx.attr.package != None:
        args += ["-package", ctx.attr.package]

    args.append(ctx.file.spec.short_path)

    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    ctx.actions.write(
        output = out_file,
        content = """\
#!/usr/bin/env bash
cd "$BUILD_WORKSPACE_DIRECTORY"
{cmd} {args}
        """.format(cmd = ctx.executable.oapi_codegen_tool.path, args = " ".join(args)),
        is_executable = True,
    )
    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = ctx.runfiles(files = [ctx.executable.oapi_codegen_tool]),
        executable = out_file,
    )]

_oapi_codegen_gen = rule(
    _oapi_codegen_gen_impl,
    executable = True,
    attrs = {
        "spec": attr.label(
            mandatory = True,
            allow_single_file = [".json", ".yaml"],
        ),
        "out": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "package": attr.string(),
        "generate": attr.string_list(),
        "oapi_codegen_tool": attr.label(
            default = Label(_CODEGEN_TOOL),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
    },
)

# rename _oapi_codegen to oapi_codegen
def oapi_codegen(name, **kwargs):
    _oapi_codegen(name = name, **kwargs)
    _oapi_codegen_gen(name = name + "_gen", **kwargs)
