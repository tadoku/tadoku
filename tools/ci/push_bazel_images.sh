#!/usr/bin/env bash
set -euo pipefail

if [[ -n "${RUNFILES_DIR:-}" && -f "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
  # shellcheck source=/dev/null
  source "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -n "${RUNFILES_MANIFEST_FILE:-}" && -f "${RUNFILES_MANIFEST_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "$(grep -m1 "^bazel_tools/tools/bash/runfiles/runfiles.bash " "${RUNFILES_MANIFEST_FILE}" | cut -d' ' -f2-)"
fi

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
