---
title: SQL performance measurements
description: Frozen SQL measurement tooling, isolation and page evidence, baseline buffer certification, required A/A calibration and qualification journeys.
---

# SQL performance measurements

Read this before collecting SQL measurements or judging a candidate with queryclimb.

`tools/queryclimb/` contains the Bazel-built PostgreSQL measurement CLI and
synthetic data generator. Keep its binary and source revision frozen outside
the checkout whose SQL it measures. Store cases, templates, results and reports
outside Git.

The CLI permits only its owned loopback PostgreSQL instance and `qc_` databases.
Run queries as the runtime role with a transaction-local tenant; migrations and
fixtures use the owner role. Preserve production planner parameters and measure
prepared statements after five executions. Writes roll back, and their
fingerprints include affected table state with declared volatile columns masked.
Table fingerprints stream sorted SHA-256 row digests, preserving row order
independence, duplicate counts and NULLs without sorting or transferring full
JSON rows.

All commands share a host lock with `tools/queryclimb/run.sh`. Use that lock for
external builds, tests and data loads too. Run one timed workload at a time and
keep heavy work outside timed batches. A failed noise, equivalence, interface or
global-work gate cannot authorize a performance change. Dashboard publication
uses the same lock and the exact stored Draft revision.

The host-load check reserves half the visible CPUs, with a minimum ceiling of
two. An absolute load of two can be the measured PostgreSQL query's own parallel
workers on a larger machine. Builds remain excluded, and the independent A/A
noise probe still decides whether each batch is usable.

The default gate compares exact fingerprints and measured buffer counters.
When a baseline varies against itself, keep its raw failures. A boundary-cut
page can instead use `page-evidence` from the baseline SQL: the gate checks
full row values, multiplicities, ordering groups and the exact number selected
from each boundary group. Enable `include_rows` only in the inspection cases
that need this evidence. Empty pages and stable pages retain exact fingerprints.

For read-only SQL changes, `isolation` attests matching migrations, schema,
data, statistics and planner settings in the two named measurement databases.
Unchanged-query exceptions require byte-identical SQL, three distinct baseline
clones with buffer variance, identical full estimated plans and prepared counters.
Enable `include_estimate` for inspection.
Index or schema changes cannot use exceptions; work, latency and noise gates
stay strict. Gate output lists every replacement assertion. Oracles bind to the
baseline measurement database.

`isolation --baseline-buffer-case <case>` additionally certifies one target
case at `--scale-index`. Both roots must contain byte-identical read-only target
SQL. The proof retains its SQL digest, complete case key, all three full witness
plans, returned rows and all five identities hashed from the actual distinct clones; baseline
and candidate identities are reused. Only `gate --mode determinism` accepts
this certificate. It permits temporary-buffer variance while requiring exact
fingerprints, shared buffers, work, inputs, prepared counters and full plans.
Missing plans or identities fail closed. Other modes reject this certificate.

Candidate mode requires `--aa-calibration`, a comma-separated file per scale,
with the same target-case coverage as the measured metrics and paired latency.
Each case must have `aa_control: true`, at least 15 samples per side, both raw
five-sample probes and complete batch-noise samples. The gate recomputes A/A
noise and rejects missing, duplicate, malformed, wrong-case or over-10% controls.
The acceptance threshold is `max(2%, 2 * max(A/A noise, A/B noise))`, including
case regressions and geometric-mean improvement. Stored aggregate noise and
threshold fields do not decide acceptance. `attempt` requires the existing A/A
files through the same flag; it never substitutes A/B measurements for controls.

Run the CLI qualification counterexamples with the frozen binary and a new
external evidence directory:

```sh
python3 tools/queryclimb/qualification_journey.py <binary> <evidence-dir>
python3 tools/queryclimb/baseline_buffer_journey.py <binary> <evidence-dir>
python3 tools/queryclimb/aa_calibration_journey.py <binary> <evidence-dir>
python3 tools/queryclimb/isolation_journey.py <binary> <repo-root> <evidence-dir> <template>
python3 tools/queryclimb/page_journey.py <binary> <repo-root> <evidence-dir> <template> <cases>
```

The gate journeys use synthetic measurement artifacts. The identity journey
clones an owned PostgreSQL template and proves that changes to data,
schema, statistics or planner settings reject the identity check. The page
journey collects real rows and an oracle from boundary-cut fixtures, accepts a
legal tied alternative and rejects an invented row.
