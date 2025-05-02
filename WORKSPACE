workspace(name = "tadoku")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

# http_archive(
#     name = "rules_proto",
#     sha256 = "80d3a4ec17354cccc898bfe32118edd934f851b03029d63ef3fc7c8663a7415c",
#     strip_prefix = "rules_proto-5.3.0-21.5",
#     urls = [
#         "https://github.com/bazelbuild/rules_proto/archive/refs/tags/5.3.0-21.5.tar.gz",
#     ],
# )

# ## Setup protobuf toolchain
# load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

# rules_proto_dependencies()

# rules_proto_toolchains()

# Setup k8s
http_archive(
    name = "io_bazel_rules_k8s",
    sha256 = "ce5b9bc0926681e2e7f2147b49096f143e6cbc783e71bc1d4f36ca76b00e6f4a",
    strip_prefix = "rules_k8s-0.7",
    urls = ["https://github.com/bazelbuild/rules_k8s/archive/refs/tags/v0.7.tar.gz"],
)

load("@io_bazel_rules_k8s//k8s:k8s.bzl", "k8s_repositories")

k8s_repositories()

# load("@io_bazel_rules_k8s//k8s:k8s_go_deps.bzl", k8s_go_deps = "deps")

# k8s_go_deps()