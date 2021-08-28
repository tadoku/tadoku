load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")

k8s_object(
  name = "api",
  kind = "deployment",
  template = ":api.yaml")

k8s_object(
  name = "migrate",
  kind = "deployment",
  template = ":migrate.yaml")
