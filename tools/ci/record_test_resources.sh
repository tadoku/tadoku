#!/usr/bin/env bash
# Copies runner and service cgroup counters into $RUNNER_TEMP for the bazel-performance artifact.
set -euo pipefail
cgroup="/sys/fs/cgroup$(sed -n 's/^0:://p' /proc/self/cgroup)"
if [[ ! -d "$cgroup" ]]; then cgroup=/sys/fs/cgroup; fi
for counter in memory.peak memory.max memory.events memory.stat cpu.max cpu.stat; do
  if [[ -r "$cgroup/$counter" ]]; then
    cat "$cgroup/$counter" > "$RUNNER_TEMP/bazel-cgroup-$counter"
  fi
done
for service in postgres pgbouncer valkey; do
  variable="${service^^}_CONTAINER"
  container="${!variable}"
  if [[ -n "$container" ]]; then
    docker exec "$container" sh -c 'cgroup="/sys/fs/cgroup$(sed -n "s/^0:://p" /proc/self/cgroup)"; if ! test -d "$cgroup"; then cgroup=/sys/fs/cgroup; fi; for counter in memory.peak memory.events cpu.stat; do echo "$counter"; cat "$cgroup/$counter"; done' \
      > "$RUNNER_TEMP/bazel-cgroup-$service" 2>&1 || true
  fi
done
