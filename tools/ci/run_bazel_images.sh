#!/usr/bin/env bash
#
# Runs a list of oci_load or oci_push binaries passed as runfiles paths in
# "$@", allowing CI to load or push all service images after one Bazel
# analysis. Each operation is wrapped in a GitHub Actions log group.
set -uo pipefail

runfiles_bash="bazel_tools/tools/bash/runfiles/runfiles.bash"
# shellcheck disable=SC1090
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

for image_operation in "$@"; do
  if ! resolved_operation="$(resolve_runfile "${image_operation}")"; then
    echo "could not resolve ${image_operation}" >&2
    exit 1
  fi

  echo "::group::${image_operation}"
  "${resolved_operation}"
  echo "::endgroup::"
done
