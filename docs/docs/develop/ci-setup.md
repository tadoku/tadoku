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

Push and PR builds can reuse successful wrapper/publication and dependency/OpenAPI
checks through exact GitHub Actions cache lookups. `lookup-only` restores no files:
checked-in generated outputs stay in place. A receipt is saved only after a
successful job; missing receipts, probe failures, cache errors and partial key
matches run the checks. There are no fallback restore keys. Cache entries are
immutable, and GitHub's branch scope prevents main from reading PR merge-ref or
fork-written entries. See [GitHub cache scope](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching#restrictions-for-accessing-a-cache).

| Receipt | Closed inputs hashed by the workflow |
| --- | --- |
| Wrapper and publication | Bazel hook and version, CI Python checks and their embedded fixtures, image runner, workflow and native runtime |
| Frozen dependencies and OpenAPI | All contract inputs and generated outputs, generator script and isolated workspace, root SDK/module/lock/build/rule/patch metadata, hook and versions, workflow and native runtime |

Source keys include the selected tracked paths, Git blob identities, actual file
modes and raw file bytes. Content-only hashes can miss same-byte renames.
Dirty tracked inputs, untracked selected inputs and symlinks whose targets are
outside the selected closure disable reuse. The successful receipt contains no
source or generated output archive.
Ignored untracked inputs also disable reuse: generation can read files Git ignores.

`tools/ci/setup_fingerprint.py` identifies the runtime using installed package
versions, native executable hashes and versions, Python version, kernel and
architecture, UID, relevant tool/locale environment, Bazel and Go configuration,
Docker client/server/runtime/security/storage configuration and resource limits.
It rejects RC imports because they can refer to files outside the known input
set. Unsupported hosts rerun the checks. Do not add an external input to a
cached check without extending its fingerprint. The OpenAPI receipt includes
the frozen module graph and checked-in outputs, so a hand-edited generated file
forces regeneration and the existing diff check. New untracked generated files also fail validation.
The dependency check uses the same frozen graph: Bazel's
[`error` lockfile mode](https://bazel.build/external/lockfile#lockfile-usage)
performs no registry network requests during resolution. Builds and tests still
use that mode on every run.

The cache action logs the exact matched key and originating cache scope is
inspectable through GitHub's cache API. On a miss, `bazel-wrapper` retains the
real execution reports; the OpenAPI receipt records its run and revision in
`bazel-performance`. On a hit, consult the successful seed run for those reports.

Run the fingerprint's native stability, invalidation and unavailable-tool check
with Python 3, Docker, Bazelisk and the wrapper utilities installed and no
other Bazel client holding the per-UID lock:

```sh
python3 tools/ci/check_setup_fingerprint.py --evidence-dir /tmp/tadoku-setup-fingerprint
python3 tools/ci/check_setup_inputs.py --evidence-dir /tmp/tadoku-setup-inputs
```

Successful test XML for every executed target is retained with profiles and
build events in the existing performance artifact in both normal and race jobs.

Gazelle validation runs as `//tools/ci/setup:gazelle_test` inside the existing
full build/test invocation. The native Gazelle rule is local, unsandboxed and
uncached: every run checks the real workspace, including changed Go and BUILD
files. Its workspace anchor is the exported root policy file. This avoids a
separate Bazel invocation and repeats no analysis solely to run Gazelle.

Normal builds allow four local test jobs within the four-action limit.
The wrapper supplies `--local_resources=memory=HOST_RAM*0.25` unless the invocation
explicitly overrides it. This command-line default takes precedence over Bazel RC
settings; container resource limits are enforced separately. Race execution keeps
its separate profile. The performance artifact also retains runner cgroup limits, peaks and
events plus each disposable service's cgroup peak/events. Missing counters or
unavailable services remain visible in these diagnostic files. These snapshots
include setup and the complete normal build/test phase; they do not replace a
measurement of scheduler or runner-factory queue time.

Workspace import, repository and Go comment policies run uncached as
`//tools/ci/setup:policies_test` in the same full test invocation. The test uses
the existing checker binaries against the real workspace, with the same
`GO111MODULE=off` standard-library classification. The standalone checker
commands remain available.

Self-hosted normal tests use Bazel's minimal remote-output download mode.
Local actions still receive their required inputs, and the full target set is
built and tested. XML and logs are explicitly downloaded even for cached tests.
Release builds and image entrypoints use their existing download behavior.
