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
