#!/usr/bin/env bash
#
# Runs a list of oci_push binaries (passed as runfiles paths in "$@") so all
# service images are pushed in a single `bazel run //:push_images` invocation,
# instead of one bazel analysis per image. Invoked by CI in
# .github/workflows/build-bazel.yaml; each push is wrapped in a GitHub Actions
# log group.
set -uo pipefail

runfiles_bash="bazel_tools/tools/bash/runfiles/runfiles.bash"
source "${RUNFILES_DIR:-/dev/null}/${runfiles_bash}" 2>/dev/null || \
  source "$(grep -sm1 "^${runfiles_bash} " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -d' ' -f2-)" 2>/dev/null || \
  source "$0.runfiles/${runfiles_bash}" 2>/dev/null || \
  source "$(grep -sm1 "^${runfiles_bash} " "$0.runfiles_manifest" | cut -d' ' -f2-)" 2>/dev/null || \
  source "$(grep -sm1 "^${runfiles_bash} " "$0.exe.runfiles_manifest" | cut -d' ' -f2-)" 2>/dev/null || \
  { echo "ERROR: cannot find ${runfiles_bash}" >&2; exit 1; }
runfiles_bash=
set -e

resolve_runfile() {
  local path="$1"
  local resolved=""

  if command -v rlocation >/dev/null 2>&1; then
    resolved="$(rlocation "${path}")"
    if [[ -n "${resolved}" && -x "${resolved}" ]]; then
      echo "${resolved}"
      return 0
    fi
  fi

  if [[ -x "${path}" ]]; then
    echo "${path}"
    return 0
  fi

  if [[ -x "../${path}" ]]; then
    echo "../${path}"
    return 0
  fi

  return 1
}

for pusher in "$@"; do
  if ! resolved_pusher="$(resolve_runfile "${pusher}")"; then
    echo "could not resolve ${pusher}" >&2
    exit 1
  fi

  echo "::group::${pusher}"
  "${resolved_pusher}"
  echo "::endgroup::"
done
