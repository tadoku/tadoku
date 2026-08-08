# Tadoku Paper static-container resource measurement plan

Status: Phase 0 measurement contract; no static Paper image exists yet

## Baseline and limitation

The live legacy Next.js styleguide currently requests `50m` CPU and `200Mi` memory and limits `250m` CPU and `512Mi` memory. A read-only cgroup snapshot on 2026-08-08 showed approximately 59.0 MB `memory.current`, 66.0 MB `memory.peak`, and PID 1 RSS approximately 50 MiB. The pod was ready with zero restarts. These figures describe the Node/Next workload and are a comparison baseline, not values to copy.

The production cluster does not expose the Kubernetes Metrics API, so `kubectl top` cannot collect Paper measurements. Installing Metrics Server or another telemetry stack is outside Tadoku Paper's deployment scope. The initial recommendation must come from local container metrics plus cgroup v2 counters in the production-shaped Paper pod.

## Decision rule

Do not merge final Paper requests/limits based only on “nginx is small” or on the legacy allocation. Build the exact static image first, run the scenarios below three times, preserve the raw observations, and derive the manifest values with the formulas in this document.

The initial hypothesis to test is:

| Resource | Provisional floor | Provisional ceiling |
| --- | ---: | ---: |
| CPU request | `10m` | measured value may justify more |
| CPU limit | `100m` | increase if legitimate burst throttling exceeds the gate |
| Memory request | `32Mi` | increase if the margin formula exceeds it |
| Memory limit | `64Mi` | increase if peak/margin exceeds it |

These are measurement starting points, not the Phase 1 result. A failed start or OOM at the provisional limit invalidates the run and requires rerunning at a safely higher limit.

## Artifacts to record

For each run, preserve in the Phase gate report:

- Tadoku commit and immutable image digest;
- static-server base image and config revision;
- uncompressed Vite `dist` bytes and file count;
- compressed registry image size/platform manifest;
- test environment (local Docker or production namespace), node architecture, and pod name;
- configured requests/limits and probe timings;
- readiness duration excluding and including cold image pull where observable;
- cgroup `memory.current`, `memory.peak`, `memory.events` and `cpu.stat` snapshots;
- PID 1 `VmRSS`/`VmHWM` for diagnostic comparison only;
- request count, concurrency, routes, cache mode, elapsed time, response failures, and latency summary;
- restart count, probe failures, OOM events, and CPU throttling delta;
- three-run median and worst observed value;
- calculated and selected requests/limits, including rounding.

## Instrumentation

### Local production-image pass

Run the exact production image with the provisional cgroup limits and its read-only/security settings. Use `docker stats --no-stream` for snapshots and `docker inspect` for state/restart/OOM evidence. Use the same static server port, config, and health endpoints as Kubernetes.

Local results catch image/config failures and give fast repeatability. They are not the final request recommendation because kernel, node, and ingress conditions differ from production.

### In-cluster pass

The Paper container must expose cgroup v2 counters readable by its unprivileged user:

```text
/sys/fs/cgroup/memory.current
/sys/fs/cgroup/memory.peak
/sys/fs/cgroup/memory.events
/sys/fs/cgroup/cpu.stat
```

Collect one snapshot immediately after readiness, at the end of idle, before/after each traffic scenario, and five minutes after traffic. `cpu.stat` is cumulative; CPU use for an interval is the delta of `usage_usec`. Throttling uses the deltas of `nr_periods`, `nr_throttled`, and `throttled_usec`.

Repeated `kubectl exec` adds a small observer cost inside the container cgroup. Use one short sampling process per scenario or take only boundary snapshots, and apply the same method to all three runs. Record this limitation. If later cluster/container metrics become available, prefer them and cross-check one run against cgroup values.

## Scenarios

Run in this order with a fresh pod for each repetition.

| Scenario | Procedure | Primary measurements |
| --- | --- | --- |
| Cold startup | Delete only the test pod through its Deployment rollout; observe scheduling, pull, container start, startup/readiness probes | time to ready, startup CPU delta, peak memory, restarts/probe failures |
| Warm startup | Restart with image already present on node where scheduling allows | server startup/readiness budget independent of pull |
| Idle | After readiness, no synthetic traffic for 15 minutes; probes remain active | median/worst `memory.current`, CPU delta per minute, throttle events |
| Cold catalogue load | One browser/session requests `/`, HTML, CSS, JS, fonts, icons with empty client cache | peak memory/CPU, bytes, failures, response headers |
| Direct deep links | Sequentially request every canonical registry route as a fresh document request | fallback correctness, CPU total, failures, memory peak |
| Representative navigation | Browser opens search and navigates at least 20 foundation/component/pattern pages using a warm cache | server request shape, stability; client behavior is a smoke gate, not server CPU alone |
| Modest sustained traffic | Five minutes at 1 request/second split across `/`, nested routes, and hashed assets | steady CPU/memory and throttling |
| Short burst | 60 seconds, concurrency 10, same route mix; no write/API traffic | limit headroom, error count, throttling, peak memory |
| Recovery idle | Five minutes with only probes after burst | memory returns to stable range; no restarts/OOM |

Do not use unique query strings for hashed assets in the representative-cache scenario. A separate no-cache document scenario may use request headers, but record it distinctly.

## Calculations

Use the worst valid observation across all three in-cluster runs.

### Memory

```text
working_peak = max(memory.peak during startup and all traffic scenarios)
request = max(32Mi, round_up_4Mi(max(idle_p95 * 1.5, navigation_p95 * 1.25)))
limit   = max(64Mi, round_up_8Mi(working_peak * 2.0))
```

The selected limit must also leave at least 16 MiB above the observed peak. If `memory.events` reports `oom`, `oom_kill`, or sustained `high` events, invalidate the run and raise the provisional limit before repeating.

### CPU

For each interval:

```text
average_cores = delta(usage_usec) / interval_usec
millicores = average_cores * 1000
```

Then select:

```text
request = max(10m, round_up_5m(max(idle_p95 * 1.5, sustained_p95 * 1.25)))
limit   = max(100m, round_up_10m(burst_p99 * 2.0))
```

If the 100m provisional limit creates throttling during cold startup or representative navigation, rerun with a higher provisional limit before calculating. A tiny static server may show noisy percentiles; retain raw deltas and use the three-run worst case rather than false precision.

## Acceptance gate

The selected manifest passes when all are true:

- three production-shaped runs complete with zero failed responses, restarts, OOM events, or probe failures;
- warm readiness completes within 3 seconds in all runs;
- memory peak is at most 50% of the selected memory limit and at most 80% of the request during idle/representative navigation;
- sustained traffic CPU is at most 80% of the request on average;
- representative navigation has zero throttled periods; short burst throttling is below 5% of cgroup periods and causes no failed/slow health probes;
- five-minute recovery memory is within 10% of the pre-traffic idle level;
- rolling update with `maxUnavailable: 0` retains one ready endpoint;
- the exact selected values, evidence table, and image digest are recorded in the Phase 1 gate report.

If the formula produces values above the legacy Node allocation, investigate the static image/server configuration before accepting them.

## Phase 4 retuning

Repeat the full measurement after the catalogue is complete because fonts, examples, search index, and asset count will have grown. Compare by digest and report deltas in `dist` size, image size, cold-load bytes, startup peak, idle, navigation, and burst.

Only lower requests/limits when all three complete-catalogue runs pass. Raise limits immediately if the complete build crosses 70% memory-limit usage, experiences probe failures, or sustains more than 5% CPU throttling. Resource tuning is a GitOps PR separate from application/source changes and must preserve the previous manifest values in the rollback record.

## Evidence table template

| Digest | Run | Ready warm/cold | Idle CPU | Sustained CPU | Burst p99 CPU | Idle memory | Peak memory | Throttled periods | Restarts/OOM | Result |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| `sha256:…` | 1 |  |  |  |  |  |  |  |  |  |
| `sha256:…` | 2 |  |  |  |  |  |  |  |  |  |
| `sha256:…` | 3 |  |  |  |  |  |  |  |  |  |

Final recommendation:

| Resource | Selected | Formula result | Headroom | Rationale |
| --- | ---: | ---: | ---: | --- |
| CPU request |  |  |  |  |
| CPU limit |  |  |  |  |
| Memory request |  |  |  |  |
| Memory limit |  |  |  |  |
