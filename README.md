# Tadoku Monorepo

![Build](https://github.com/tadoku/tadoku/actions/workflows/build.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=tadoku/tadoku)](https://dependabot.com)

## Getting Started

You will need to [install Bazel](https://docs.bazel.build/versions/4.1.0/install.html#installing-bazel-1) in order to build and run the monorepo.

## Commands

### Build all targets

```sh
$ bazel build //...
```

### Run all tests
```sh
$ bazel test //...
```

### Copy go modules dependencies to Bazel

```sh
$ bazel run //:gazelle -- update-repos --from_file=go.mod -to_macro=go_third_party.bzl%go_deps
```