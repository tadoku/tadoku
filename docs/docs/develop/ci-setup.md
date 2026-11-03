---
title: CI setup verification
description: Exact-input setup verification receipts, runtime invalidation and cache trust boundaries in backend CI.
sidebar_position: 4
---

# CI setup verification

Read this when changing backend setup checks or investigating a setup cache hit.

Service health probes run every second during startup, then every ten seconds.
PostgreSQL readiness uses loopback TCP so its temporary first-boot Unix socket
cannot report ready before the final server starts.
CI runs depolicy with `GO111MODULE=off`: its import-only parser still reads the
module and policy files directly, while standard-library classification uses
the pinned SDK's source tree without spawning `go list` for external imports.

Push and PR builds can reuse successful wrapper/publication and OpenAPI checks
through exact GitHub Actions cache lookups. `lookup-only` restores no files:
checked-in generated outputs stay in place. A receipt is saved only after a
successful job; missing receipts, probe failures, cache errors and partial key
matches run the checks. There are no fallback restore keys. Cache entries are
immutable, and GitHub's branch scope prevents main from reading PR merge-ref or
fork-written entries. See [GitHub cache scope](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching#restrictions-for-accessing-a-cache).

| Receipt | Closed inputs hashed by the workflow |
| --- | --- |
| Wrapper and publication | Bazel hook and version, CI Python checks and their embedded fixtures, image runner, workflow and native runtime |
| OpenAPI | All contract inputs and generated outputs, generator script and isolated workspace, root SDK/module/lock/build/rule/patch metadata, hook and versions, workflow and native runtime |

`tools/ci/setup_fingerprint.py` identifies the runtime using installed package
versions, native executable hashes and versions, Python version, kernel and
architecture, UID, relevant tool/locale environment, Bazel and Go configuration,
Docker client/server/runtime/security/storage configuration and resource limits.
It rejects RC imports because they can refer to files outside the known input
set. Unsupported hosts rerun the checks. Do not add an external input to a
cached check without extending its fingerprint. The OpenAPI receipt includes
the checked-in outputs, so a hand-edited generated file forces regeneration
and the existing diff check. New untracked generated files also fail validation.

The cache action logs the exact matched key and originating cache scope is
inspectable through GitHub's cache API. On a miss, `bazel-wrapper` retains the
real execution reports; the OpenAPI receipt records its run and revision in
`bazel-performance`. On a hit, consult the successful seed run for those reports.

Run the fingerprint's native stability, invalidation and unavailable-tool check
with Python 3, Docker, Bazelisk and the wrapper utilities installed:

```sh
python3 tools/ci/check_setup_fingerprint.py --evidence-dir /tmp/tadoku-setup-fingerprint
```

Successful test XML for every executed target is retained with profiles and
build events in the existing performance artifact in both normal and race jobs.

