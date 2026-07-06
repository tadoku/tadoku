---
slug: /
sidebar_position: 1
title: Getting Started
---

# Tadoku

## Getting Started

You will need to [install Bazel](https://bazel.build/start) in order to build and run the monorepo.

## Commands

### Build all targets

```sh
$ bazel build //...
```

### Run all tests

```sh
$ bazel test //...
```

### Generate BUILD files for all Golang code

```sh
bazel run //:gazelle
```

CI verifies that BUILD files are up to date with `bazel run //:gazelle -- -mode=diff`, so run this after adding or removing Go files or changing imports.
