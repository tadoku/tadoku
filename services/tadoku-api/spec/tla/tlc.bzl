"""Runs TLC over a TLA+ model as a Bazel test."""

load("@rules_shell//shell:sh_test.bzl", "sh_test")

_JAVA_RUNTIME = "@rules_java//toolchains:current_java_runtime"
_TLA2TOOLS = "@tla2tools//file"

def tlc_test(name, spec, config, size = "small", tags = [], **kwargs):
    """Fails when TLC finds a violated invariant, property or deadlock in spec under config.

    A model outside CI's small budget must be tagged manual.
    """
    if size != "small" and "manual" not in tags:
        fail("%s: a tlc_test larger than small must be tagged manual so CI stays fast" % name)

    sh_test(
        name = name,
        size = size,
        srcs = ["tlc_test.sh"],
        args = [
            "$(JAVA_ROOTPATH)",
            "$(rootpath %s)" % _TLA2TOOLS,
            "$(rootpath %s)" % spec,
            "$(rootpath %s)" % config,
        ],
        data = [
            _JAVA_RUNTIME,
            _TLA2TOOLS,
            spec,
            config,
        ],
        tags = tags + ["cpu:2"],
        toolchains = [_JAVA_RUNTIME],
        **kwargs
    )
