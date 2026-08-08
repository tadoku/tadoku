# Bazel modernization migration

## Status

Phase 0 baseline captured. Rollout in progress.

Last updated: 2026-08-08

The current repository uses Bazel 8.7.0 with Bzlmod, `rules_go` 0.61.1,
Gazelle 0.51.3, `rules_pkg` 1.2.0, and `rules_oci` 2.3.0. The target is
Bazel 9.2.0 with explicit shell rules, current Go build rules, a regenerated
lockfile, and CI that rejects stale dependency state.

The supporting audit is available in Presentr:
[Tadoku Bazel modernization report](https://presentr.lab/html/tadoku-bazel-modernization-2026-08).

## Progress tracker

Update this table in the same pull request that advances a phase. Use only the
statuses `Not started`, `In progress`, `Blocked`, and `Done`. A phase is not
`Done` until its exit criteria have passed on `main`.

| Phase | Status | Pull request | Depends on | Evidence / notes |
| --- | --- | --- | --- | --- |
| 0. Capture the baseline | Done | [#765](https://github.com/tadoku/tadoku/pull/765) | — | Merged 2026-08-08 as `b6ba1845`; baseline evidence recorded below. |
| 1. Make rules Bazel 9-ready on Bazel 8 | Done | [#766](https://github.com/tadoku/tadoku/pull/766) | Phase 0 | Merged 2026-08-08 as `be66af7e`; [full Bazel 8 CI passed](https://github.com/tadoku/tadoku/actions/runs/31237113928). |
| 2. Upgrade Bazel to 9.2.0 | Done | [#767](https://github.com/tadoku/tadoku/pull/767) | Phase 1 merged | Merged 2026-08-08 as `42bba884`; [full Bazel 9 CI passed](https://github.com/tadoku/tadoku/actions/runs/31237357627). |
| 3. Harden CI and dependency updates | In progress | [#768](https://github.com/tadoku/tadoku/pull/768) | Phase 2 merged | Local gates and the [full branch PR workflow](https://github.com/tadoku/tadoku/actions/runs/31237719680) passed; post-merge cache validation pending. |
| 4. Clean up latent Bazel debt | Not started | — | Phase 3 merged | — |
| 5. Complete rollout and close migration | Not started | — | Phases 0–4 | — |

## Phase 0 baseline evidence

Captured on 2026-08-08 from commit `935b8737`, based on remote `main` at
`f3dba363`, before any dependency or rule changes:

| Check | Result |
| --- | --- |
| Version pins | Bazel 8.7.0; `rules_go` 0.61.1; Gazelle 0.51.3; `rules_pkg` 1.2.0; `rules_oci` 2.3.0; Go 1.26.5 from `go.mod` |
| Lockfile format | 24 |
| Frozen dependencies | Pre-existing failure: the implementation of the `rules_python` `pip` extension or one of its transitive `.bzl` files changed |
| Gazelle | Passed with lockfile updates disabled; no diff |
| Build | All 142 targets passed |
| Tests | All 16 tests passed |
| OCI coverage | `load_images` and `push_images` each exactly covered six production service images |
| Image loading | All six production service images loaded successfully |

The frozen-dependency failure was observed without modifying
`MODULE.bazel.lock`. Subsequent baseline commands used
`--lockfile_mode=off` to keep Phase 0 read-only.

## Working rules

1. Complete phases in order. Do not start the Bazel 9 lockfile rewrite before
   the Bazel 8 rule-readiness change is merged.
2. Use one pull request per phase. Do not mix application behavior, database
   migrations, Go dependency upgrades, or unrelated generated changes into a
   Bazel migration pull request.
3. Base every phase on the latest `main` after the previous phase is merged.
4. Commit the complete `MODULE.bazel.lock` output produced by Bazel. Do not
   hand-edit, partially stage, or discard unexpected generated lockfile
   changes.
5. Use Bazel for backend build and test commands. Do not substitute direct
   `go build` or `go test` commands.
6. Record the pull request, completion date, CI run, and any deviations in the
   progress table before marking a phase `Done`.
7. If a phase fails, keep its pull request isolated and revert that phase. The
   earlier merged phase must remain independently valid.

## Scope

### Goals

- Move the repository from Bazel 8.7.0 to the active Bazel 9.2.0 LTS.
- Replace the removed native `sh_binary` symbol with an explicit
  `rules_shell` dependency.
- Update `rules_go` and Gazelle to versions verified with Bazel 9.
- Regenerate and commit the Bazel 9 lockfile without omitting generated
  changes.
- Make CI fail when the checked-in lockfile is stale.
- Make Bazel installation and caching explicit in GitHub Actions.
- Automate routine Bzlmod dependency update pull requests.

### Non-goals

- Changing application behavior or service APIs.
- Changing the Go version or application Go dependencies.
- Changing database schemas or data.
- Replacing `rules_oci`, `rules_pkg`, or the distroless base image.
- Adding multi-architecture container publishing.
- Enabling experimental Bazel 10 behavior in production builds.
- Introducing remote execution or a new organization-wide build service.

## Target version set

| Component | Current | Target | Migration action |
| --- | --- | --- | --- |
| Bazel | 8.7.0 | 9.2.0 | Upgrade in Phase 2 |
| `rules_go` | 0.61.1 | 0.62.0 | Upgrade in Phase 1 |
| Gazelle | 0.51.3 | 0.52.2 | Upgrade in Phase 1 |
| `rules_shell` | Transitive only | 0.8.0 direct | Add in Phase 1 |
| `rules_pkg` | 1.2.0 | 1.2.0 | Keep |
| `rules_oci` | 2.3.0 | 2.3.0 | Keep |
| Go SDK | From `go.mod` | Unchanged | Keep |

## Phase 0: capture the baseline

### Purpose

Create a reproducible record of the last known-good Bazel 8 state before any
dependency or rule changes. This distinguishes pre-existing failures from
migration regressions.

### Steps

- [x] Set Phase 0 to `In progress` in the progress tracker.
- [x] Create a dedicated baseline pull request or issue and link it in the
      tracker.
- [x] Confirm the checkout is based on the latest `main` and has no unrelated
      changes.
- [x] Record the current values from `.bazelversion`, `MODULE.bazel`, and
      `go.mod` in the pull request description.
- [x] Record the current lockfile format:

  ```sh
  jq .lockFileVersion MODULE.bazel.lock
  ```

- [x] Run the full dependency resolution check:

  ```sh
  bazel mod deps --lockfile_mode=error
  ```

- [x] If the lockfile check fails, record the exact error as pre-existing.
      Do not update the lockfile in Phase 0.
- [x] Verify Gazelle produces no diff:

  ```sh
  bazel run //:gazelle -- -mode=diff
  ```

- [x] Build and test all Bazel targets:

  ```sh
  bazel build //...
  bazel test //...
  ```

- [x] Run the two production image coverage queries from
      `.github/workflows/build-bazel.yaml` and confirm `//:load_images` and
      `//:push_images` cover every expected production target.
- [x] Run `bazel run //:load_images` and confirm all backend images load
      locally. Do not push images from a baseline branch.
- [x] Attach command output or the successful CI run to the tracking issue or
      pull request.
- [x] Update this phase to `Done` only after the baseline evidence is recorded.

### Exit criteria

- Bazel 8.7.0 builds every target and all tests pass, or every pre-existing
  failure is explicitly recorded.
- Gazelle and image aggregation state are known.
- The lockfile's current status is known without modifying it.

### Rollback

Phase 0 is read-only apart from this tracker. No build rollback is required.

## Phase 1: make rules Bazel 9-ready on Bazel 8

### Purpose

Remove the known Bazel 9 analysis failure while still running Bazel 8.7.0.
Keeping the core Bazel version unchanged makes rule and dependency failures
easy to isolate.

### Steps

- [x] Branch from `main` after Phase 0 is complete.
- [x] Set Phase 1 to `In progress` and link the pull request.
- [x] Add the direct shell dependency to `MODULE.bazel`:

  ```starlark
  bazel_dep(name = "rules_shell", version = "0.8.0")
  ```

- [x] Load `sh_binary` explicitly at the top of the root `BUILD.bazel`:

  ```starlark
  load("@rules_shell//shell:sh_binary.bzl", "sh_binary")
  ```

- [x] Keep both existing `sh_binary` targets functionally unchanged.
- [x] Update `rules_go` from 0.61.1 to 0.62.0.
- [x] Update Gazelle from 0.51.3 to 0.52.2.
- [x] Keep `rules_pkg` and `rules_oci` at 1.2.0 and 2.3.0.
- [x] Regenerate the complete lockfile with Bazel 8:

  ```sh
  bazel mod deps --lockfile_mode=update
  ```

- [x] Run module cleanup and inspect any proposed `MODULE.bazel` changes:

  ```sh
  bazel mod tidy
  ```

- [x] Commit every lockfile change generated by the pinned Bazel 8 version.
- [x] Verify that dependency resolution is frozen after regeneration:

  ```sh
  bazel mod deps --lockfile_mode=error
  ```

- [x] Verify Gazelle, build, and tests:

  ```sh
  bazel run --lockfile_mode=error //:gazelle -- -mode=diff
  bazel build --lockfile_mode=error //...
  bazel test --lockfile_mode=error //...
  ```

- [x] Re-run the production image coverage queries.
- [x] Run `bazel run --lockfile_mode=error //:load_images` and confirm every
      image still loads.
- [x] Review the lockfile diff for provenance and completeness, not for
      minimizing generated output.
- [x] Merge only when Bazel 8.7.0 remains green on `main`.
- [x] Record the merged pull request and CI run, then mark Phase 1 `Done`.

### Exit criteria

- `rules_shell` is a declared direct dependency and `sh_binary` is explicitly
  loaded.
- `rules_go` 0.62.0 and Gazelle 0.52.2 pass on Bazel 8.7.0.
- Frozen dependency resolution, Gazelle, full build, tests, and image loading
  all pass.
- The pull request contains no Bazel core-version change.

### Rollback

Revert the Phase 1 commit. Because Bazel remains at 8.7.0, the original native
shell rule and previous dependency versions remain usable.

## Phase 2: upgrade Bazel to 9.2.0

### Purpose

Move the core build tool and lockfile format after every BUILD target is ready
for externally loaded shell rules.

### Steps

- [x] Branch from `main` after Phase 1 is merged.
- [x] Set Phase 2 to `In progress` and link the pull request.
- [x] Change `.bazelversion` from 8.7.0 to 9.2.0.
- [x] Remove the stale `.bazelrc` comment that says the repository still needs
      a Bzlmod migration.
- [x] Remove `common --enable_workspace=false`; Bazel 9 has removed WORKSPACE
      support and the flag is redundant.
- [x] Confirm Bazelisk resolves the new pin:

  ```sh
  bazel --version
  ```

- [x] Regenerate the lockfile with Bazel 9.2.0:

  ```sh
  bazel mod deps --lockfile_mode=update
  ```

- [x] Confirm the lockfile format changed to the Bazel 9 format and commit the
      complete generated rewrite:

  ```sh
  jq .lockFileVersion MODULE.bazel.lock
  ```

- [x] Run module graph and tidy checks:

  ```sh
  bazel mod graph --lockfile_mode=error
  bazel mod tidy
  git diff --exit-code MODULE.bazel
  ```

- [x] Verify frozen dependency resolution:

  ```sh
  bazel mod deps --lockfile_mode=error
  ```

- [x] Verify Gazelle, build, and tests:

  ```sh
  bazel run --lockfile_mode=error //:gazelle -- -mode=diff
  bazel build --lockfile_mode=error //...
  bazel test --lockfile_mode=error //...
  ```

- [x] Re-run both production image coverage queries.
- [x] Run `bazel run --lockfile_mode=error //:load_images`.
- [x] Run the same Trivy scan performed by the Bazel GitHub Actions workflow.
- [x] Confirm the pull request changes only Bazel configuration and generated
      dependency state.
- [x] Merge and verify the first `main` build completes with Bazel 9.2.0.
- [x] Record the merge commit and CI run, then mark Phase 2 `Done`.

### Exit criteria

- `.bazelversion` pins exactly Bazel 9.2.0.
- The lockfile was generated by Bazel 9.2.0 and passes `error` mode.
- All targets build and all tests pass.
- Production OCI targets still load and pass the existing security scan.
- No application or database behavior changed.

### Rollback

Revert the complete Phase 2 commit, including `.bazelversion`, `.bazelrc`, and
`MODULE.bazel.lock`. Do not restore only the old version pin because the Bazel
8 and Bazel 9 lockfile formats are coupled to their corresponding pins.

## Phase 3: harden CI and dependency updates

### Purpose

Prevent silent lockfile drift and make the pinned Bazel/Bazelisk and cache
behavior explicit on GitHub-hosted runners.

### Steps

- [x] Branch from `main` after Phase 2 is merged.
- [x] Set Phase 3 to `In progress` and link the pull request.
- [x] Add `pull_request` to `.github/workflows/build-bazel.yaml` with the same
      relevant path filters as the push workflow.
- [x] Add top-level `permissions: contents: read`.
- [x] Grant `packages: write` only to the image publish job if it is required
      for GHCR publishing.
- [x] Add workflow concurrency so a newer commit cancels a superseded build
      for the same branch or pull request.
- [x] Replace the manual `/home/runner/.cache/bazel` cache with
      `bazel-contrib/setup-bazel`.
- [x] Pin the setup action to a reviewed commit SHA, with its release version
      documented in a comment.
- [x] Enable the Bazelisk download cache, repository cache, and per-workflow
      disk cache.
- [x] Prevent pull requests from saving shared caches while still allowing
      cache restoration.
- [x] Add a dedicated frozen lockfile step before analysis:

  ```sh
  bazel mod deps --lockfile_mode=error
  ```

- [x] Pass `--lockfile_mode=error` to Gazelle, build, test, query, and image
      load invocations in CI where the command accepts it.
- [x] Add `.github/dependabot.yml` with a weekly `bazel` package ecosystem
      update at the repository root.
- [x] Group compatible Bazel rule updates so the regenerated lockfile is
      reviewed once per update set.
- [x] Do not configure Dependabot to combine Bazel updates with Go, pnpm, or
      GitHub Actions updates.
- [x] Validate the workflow on a pull request from a branch and, if practical,
      from a fork.
- [ ] Merge and confirm the next scheduled or `main` run restores and saves
      the intended caches.
- [ ] Record the merge and CI evidence, then mark Phase 3 `Done`.

### Exit criteria

- Pull requests receive the full Bazel validation workflow.
- CI cannot modify or tolerate a stale checked-in lockfile.
- Bazelisk and cache setup are explicit and version-controlled.
- Publish credentials are unavailable to validation-only jobs.
- Dependabot can open isolated Bzlmod update pull requests with regenerated
  lockfiles.

### Rollback

Revert the Phase 3 workflow/configuration commit. This does not require
reverting Bazel 9; the build remains runnable locally and with the previous CI
cache arrangement.

## Phase 4: clean up latent Bazel debt

### Purpose

Remove or repair build files that are currently unreachable and therefore not
covered by normal Bazel analysis.

### Steps

- [ ] Branch from `main` after Phase 3 is merged.
- [ ] Set Phase 4 to `In progress` and link the pull request.
- [ ] Confirm whether `build/openapi.bzl` and `build/oapi-codegen.bzl` have any
      intended consumer.
- [ ] For `build/openapi.bzl`, either delete it as obsolete or declare and test
      its `rules_openapi` dependency before loading it from a BUILD target.
- [ ] For `build/oapi-codegen.bzl`, either delete it as obsolete or restore an
      explicit, hermetic code-generator target instead of the absent
      `//third_party/oapi-codegen` label.
- [ ] Keep OpenAPI generator cleanup separate from generated OpenAPI API
      changes. If generation output would change, stop and plan that as a
      dedicated follow-up.
- [ ] Run Buildifier or the repository's accepted Starlark formatter/checker
      across active BUILD and `.bzl` files.
- [ ] Consider adding the documented Bazel lockfile merge driver entry to
      `.gitattributes`; keep developer-local merge-driver setup optional.
- [ ] Verify frozen dependencies, Gazelle, build, and tests again.
- [ ] Record the outcome for both generator files in this document.
- [ ] Merge, record the pull request, and mark Phase 4 `Done`.

### Exit criteria

- No dormant `.bzl` file references an undeclared or missing dependency without
  an explicit documented reason.
- Cleanup does not change generated API behavior.
- The full Bazel validation suite remains green.

### Rollback

Revert the cleanup commit. It is intentionally separated from the Bazel 9
upgrade and CI changes.

## Phase 5: complete rollout and close migration

### Purpose

Confirm the merged sequence works in normal development and publishing, then
turn this document into the durable record of the migration.

### Steps

- [ ] Confirm Phases 0–4 are marked `Done` with pull requests and evidence.
- [ ] Confirm at least one post-upgrade pull request completed the Bazel 9 CI
      workflow using frozen lockfile mode.
- [ ] Confirm at least one scheduled or `main` workflow restored and saved the
      new Bazel caches.
- [ ] Confirm the `main` publish job built, scanned, and published every
      production backend image expected by the coverage queries.
- [ ] Smoke-test the published backend images in the normal deployment flow.
- [ ] Review the first automated Bazel dependency pull request and adjust its
      grouping only if the result is too broad to review safely.
- [ ] Record final Bazel and rule versions in the progress tracker notes.
- [ ] Record any intentionally deferred items under Follow-ups.
- [ ] Change the top-level status to `Migration complete`.
- [ ] Mark Phase 5 `Done` in the progress tracker.

### Exit criteria

- Bazel 9.2.0 is used on `main`, in CI, and for published backend images.
- The version pin and lockfile remain synchronized.
- Routine dependency updates have an owner and automated path.
- No migration phase or undocumented blocker remains open.

### Rollback

If a production-only build or image regression appears, first revert the
smallest responsible phase. Revert Phase 2 as one unit only when the failure is
specific to Bazel 9; keep the Phase 1 explicit rule loads if they are not the
cause.

## Required validation matrix

Record a link or concise result for every row in the phase pull request.

| Check | Phase 0 | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
| --- | --- | --- | --- | --- | --- |
| `bazel mod deps --lockfile_mode=error` | Observe | Pass | Pass | Pass in CI | Pass |
| Gazelle diff | Pass | Pass | Pass | Pass in CI | Pass |
| `bazel build //...` | Pass | Pass | Pass | Pass in CI | Pass |
| `bazel test //...` | Pass | Pass | Pass | Pass in CI | Pass |
| OCI target coverage queries | Pass | Pass | Pass | Pass in CI | Pass |
| `//:load_images` | Pass | Pass | Pass | Pass in CI | Pass |
| Trivy backend image scan | Record | Pass | Pass | Pass in CI | Pass |
| Clean Git worktree after generators | Record | Pass | Pass | Pass in CI | Pass |

## Follow-ups

These items are deliberately outside this migration and should be tracked in
separate issues if they become valuable:

- Add an allowed-to-fail scheduled compatibility job for a Bazel 10 rolling
  release after Bazel 9 has been stable on `main`.
- Evaluate remote cache or remote execution only with measured CI or developer
  build-time data.
- Add Linux ARM64 OCI images as a separate container-platform project.
- Reassess Bazel and rule versions at least once per Bazel minor release or
  quarterly, whichever is less frequent.
