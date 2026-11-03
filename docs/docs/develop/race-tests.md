---
title: Race tests
description: The daily race-instrumented Tadoku API test workflow, its schedule, manual runs, runner setup and failure alerts.
---

# Race tests

Read this when a race-test run fails, when you add code that needs race
coverage, or when you change `.github/workflows/race-bazel.yaml`.

`race-bazel.yaml` runs the Tadoku API's race-instrumented targets, including the
HTTP E2E target, with `--@rules_go//go/config:race`. It passes
`--nocache_test_results`, so every listed target runs fresh each time instead of
reusing a cached pass; compilation still uses the Bazel cache. Add a target to
its list when new code needs race coverage. It is scheduled daily at 15:37 UTC.
GitHub starts this repository's scheduled runs hours late and can skip them, so
runs are not exactly 24 hours apart. Pushes and pull requests do not run it,
except pull requests that change the workflow itself. Start it manually from the
Actions tab, or with `gh workflow run race-bazel.yaml --ref main`, to validate a
fix or recover a missed run.

It uses the same `tadoku-ci` runner, digest-pinned services, private Bazel
cache, Lab CA and resource budget as `build-bazel.yaml`, and retains the same
`bazel-performance` artifact. Keep its service images identical to
`build-bazel.yaml`'s, which the runner Pods pre-pull.

When a scheduled or manual run fails, times out or is cancelled, its `alert`
job posts a `warning` to the homelab Notify service on topic `ci` with the
result, revision and run link; Notify routes warnings to the owner. Notify must
admit `tadoku-ci` runner Pods as a producer through the homelab network
policies. The job fails with an error annotation when Notify does not accept
the alert, so the run still shows the failure.
