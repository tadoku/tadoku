"""Runs TLC over a TLA+ model as a Bazel test."""

load("@rules_shell//shell:sh_test.bzl", "sh_test")

def tlc_test(name, spec, config, size = "small", **kwargs):
    """Fails when TLC finds a violated invariant, property or deadlock in spec under config."""
    sh_test(
        name = name,
        size = size,
        srcs = ["tlc_test.sh"],
        args = [
            "$(rootpath :tlc)",
            "$(rootpath %s)" % spec,
            "$(rootpath %s)" % config,
        ],
        data = [
            ":tlc",
            spec,
            config,
        ],
        **kwargs
    )
