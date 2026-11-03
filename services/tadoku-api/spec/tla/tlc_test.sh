#!/usr/bin/env bash
# Usage: tlc_test.sh <java> <tla2tools.jar> <Spec.tla> <Spec.cfg>
# Copies the model into a writable directory and runs TLC there, passing TLC's
# exit status through: 0 when every check holds, non-zero on any violation.
# TLC uses two worker threads, a 512 MiB heap and a private temporary directory.
set -euo pipefail

java="$PWD/$1"
jar="$PWD/$2"
spec="$3"
config="$4"
work="$TEST_TMPDIR/model"

mkdir -p "$work"
cp "$spec" "$config" "$work/"
cd "$work"

exec "$java" -XX:+UseParallelGC -XX:ParallelGCThreads=2 -Xmx512m -Djava.io.tmpdir="$TEST_TMPDIR" \
  -cp "$jar" tlc2.TLC \
  -workers 2 -fp 0 -seed 0 -cleanup -metadir "$TEST_TMPDIR/states" \
  -config "$(basename "$config")" "$(basename "$spec")"
