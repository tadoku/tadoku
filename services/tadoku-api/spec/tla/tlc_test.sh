#!/usr/bin/env bash
# Usage: tlc_test.sh <tlc launcher> <Spec.tla> <Spec.cfg>
# Copies the model into a writable directory and runs TLC there, passing TLC's
# exit status through: 0 when every check holds, non-zero on any violation.
set -euo pipefail

tlc="$PWD/$1"
spec="$2"
config="$3"
work="$TEST_TMPDIR/model"

mkdir -p "$work"
cp "$spec" "$config" "$work/"
cd "$work"

export JAVA_RUNFILES="${RUNFILES_DIR:-$TEST_SRCDIR}"
exec "$tlc" -workers 2 -cleanup -metadir "$TEST_TMPDIR/states" \
  -config "$(basename "$config")" "$(basename "$spec")"
