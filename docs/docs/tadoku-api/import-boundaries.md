---
title: Import boundaries
description: How Bazel visibility, package groups and CI checks enforce the Tadoku API layer and provider dependency boundaries, and how to verify an import change.
sidebar_position: 8
---

# Import boundaries

Read this when you add an import, a package or a feature, or change a Bazel
`visibility`.

## Visibility and package groups

Bazel target `visibility` and the package groups in
`services/tadoku-api/BUILD.bazel` (`:feature_consumers`, `:domain_consumers`,
`:infrastructure_consumers` and `:internal_consumers`) define the local layer
boundaries described in [Code ownership](./index.md#code-ownership).

- Feature libraries are visible only to application, startup and E2E packages.
- A feature's generated sqlc package is visible only to that feature.
- Tests follow their package's layer policy. Startup (`cmd/tadoku-api`) and E2E
  packages are assembly boundaries.
- `bazel build //services/tadoku-api/...` checks the boundaries for every Go
  package and test it builds. `rules_go` requires direct imports to be declared
  in `deps`, and Gazelle's diff check keeps `deps` aligned with source imports.

## Adding a package

- Gazelle resolves Go imports to Bazel `deps` and preserves existing
  visibility, but it creates new Go libraries as public. After running Gazelle,
  replace that default with the narrowest matching scoped visibility.
- CI rejects public or out-of-service Tadoku API `go_library` visibility.
  Omitted visibility is Bazel-private.
- Do not widen visibility to `public` merely to make a build pass.
- When an import boundary intentionally changes, update the relevant target
  visibility or package group and this page in the same change, and explain the
  new dependency direction in the PR.

## Provider dependencies

`./tools/ci/check_tadoku_api_provider_deps.sh` checks Bazel's direct dependency
graph:

- Only `features/leaderboard`, `services/tadoku-api/infra/valkey/`, startup and E2E may depend
  directly on `valkey-go`.
- Only `internal/permissions`, startup and E2E may depend directly on the raw
  Keto client (`services/tadoku-api/infra/keto`).

This keeps other packages on the shared permission checker. Keep shared ban and
administrator relation lookups in `internal/permissions`; the check cannot
detect a raw `banned` or `admins` relation literal inside an allowed package.

## Verifying an import change

Before publishing a PR that changes imports, packages or visibility, run from
the repository root:

```sh
bazel run //:gazelle -- -mode=diff
./scripts/check-tadoku-api-visibility.sh
./tools/ci/check_tadoku_api_provider_deps.sh
bazel build //services/tadoku-api/...
```

CI runs the same checks.

## Depolicy backstop

The depolicy check (`bazel run //tools/ci/depolicy`, configured in
`.depolicy.yaml`) runs in CI as a temporary backstop. It scans every Go file,
including tests and inactive build-tag files. Its YAML is not the source of
truth for new package boundaries and does not define the preferred package
structure.

## Known gaps

- Visibility is owned by the imported target, so Bazel does not restrict Tadoku
  API imports from the retained public `services/common` infrastructure packages.
- A normal build does not inspect Go files excluded by the active build
  configuration.
- The CI visibility guard prevents a new public library but cannot tell whether
  its chosen scope matches its architectural role.
- Import rules do not enforce same-package service and repository
  responsibilities or business signatures; those still need review.

Close or accept these gaps before relying on Bazel alone.
