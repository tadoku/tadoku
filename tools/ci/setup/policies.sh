#!/usr/bin/env bash
set -euo pipefail

source "$TEST_SRCDIR/bazel_tools/tools/bash/runfiles/runfiles.bash"
BUILD_WORKSPACE_DIRECTORY="$(dirname "$(readlink "$(rlocation "$1")")")"
export BUILD_WORKSPACE_DIRECTORY
export GO111MODULE=off
shift

for tool in "$@"; do
  "$(rlocation "$tool")"
done
