"""Tadoku's Bazel-owned dev-cli metadata; lifecycle stays in the CLI."""

def _impl(ctx):
    metadata = json.decode(ctx.attr.metadata)
    metadata["workloadTemplate"] = ctx.file.workload_template.short_path
    output = ctx.actions.declare_file(ctx.label.name + ".dev.json")
    ctx.actions.write(output, json.encode(metadata) + "\n")
    return [DefaultInfo(files = depset([output]))]

dev_deployable = rule(
    implementation = _impl,
    attrs = {
        "metadata": attr.string(mandatory = True),
        "workload_template": attr.label(allow_single_file = [".yaml"], mandatory = True),
        "deps": attr.label_list(allow_files = True),
    },
)
