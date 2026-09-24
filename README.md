# Tadoku Monorepo

[![Documentation](https://img.shields.io/badge/docs-online-6969FF.svg)](https://tadoku.github.io/tadoku/)
![Build Bazel](https://github.com/tadoku/tadoku/actions/workflows/build-bazel.yaml/badge.svg)
![Build Frontend webv2](https://github.com/tadoku/tadoku/actions/workflows/build-frontend-webv2.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)

Developer documentation lives in [`docs/docs`](docs/docs) and is published at https://tadoku.github.io/tadoku/.

## Dev Environment

Branches run on the shared `homelab-dev` cluster with DevCLI. Installation,
branch databases, seed accounts, routing and cleanup are in
[Development environment](docs/docs/develop/environment.md).

To verify a change in the browser, follow the
[verify-tadoku skill](.agents/skills/verify-tadoku/SKILL.md) (`$verify-tadoku`
where repository skill discovery is supported) and its
[feature map](.agents/skills/verify-tadoku/references/features/README.md).
