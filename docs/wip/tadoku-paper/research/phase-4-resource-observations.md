# Tadoku Paper Phase 4 raw resource observations

Measurement date: 2026-08-08 UTC

Source commit: `f297aa7899d54dba2302d6d26f377fe9fe9d570b`

Image digest: `sha256:4f15ab54ef5d6844df798ad3b08530f1a162ee0c6424a77977a362786a07b714`

Environment: `lke20948-ctx`, namespace `tdk-prod-paper-styleguide`, deployment `tadoku-paper-styleguide`, node architecture `x86_64`, public traffic through `https://paper.tadoku.app`.

Configured resources: CPU request `10m`, CPU limit `250m`, memory request `32Mi`, memory limit `64Mi`.

The reproducible driver is [scripts/measure-paper-resources.sh](scripts/measure-paper-resources.sh). Each run used a fresh rolling-restart pod, 15 minutes of probe-only idle, all static files, all 40 canonical routes, 20 representative navigation routes, 300 requests at 1 request/second, 60 seconds at concurrency 10, and five minutes of recovery. Every `memory.events` counter was zero in every snapshot. Short `kubectl exec` snapshots execute inside the measured cgroup and can add one throttled period; the observer-neutral navigation controls below separate that cost from traffic.

## Run 1

Pod `tadoku-paper-styleguide-5f9f6d4c76-7c5pn`; pod start to Ready 4,000 ms; container start to Ready 2,000 ms; zero restarts and warning events.

```csv
label,memory.current,memory.peak,cpu.usage_usec,cpu.nr_periods,cpu.nr_throttled,cpu.throttled_usec
ready,7979008,10321920,95121,5,1,52164
idle_00,8192000,10588160,128626,7,2,92212
idle_05,4448256,10588160,176566,32,3,164213
idle_10,4399104,10588160,219509,53,3,164213
idle_15,4411392,10588160,263275,74,4,209117
cold_catalogue_before,4714496,10588160,296198,77,5,278009
cold_catalogue_after,9035776,11456512,335088,85,5,278009
deep_links_before,9166848,11472896,367508,87,6,319014
deep_links_after,9052160,11476992,405165,98,7,361315
navigation_before,9142272,11476992,438965,100,8,405775
navigation_after,9109504,11476992,475011,106,9,438328
sustained_before,9043968,11481088,507598,108,10,490248
sustained_after,4419584,11481088,610670,240,11,537315
burst_before,4354048,11481088,644647,242,12,576071
burst_after,4599808,11481088,2953892,845,13,597265
recovery_05,4534272,11481088,3003835,866,13,597265
```

| Scenario | Requests | Failures | Average latency | Maximum latency |
| --- | ---: | ---: | ---: | ---: |
| Cold catalogue | 11 | 0 | 66.7 ms | 306.8 ms |
| Canonical deep links | 40 | 0 | 23.8 ms | 96.4 ms |
| Representative navigation | 20 | 0 | 21.4 ms | 22.8 ms |
| Sustained | 300 | 0 | 42.7 ms | 466.7 ms |
| Burst | 15,839 | 0 | 33.8 ms | 5,178.7 ms |

Sustained CPU averaged `0.328m`. Burst CPU averaged `37.9m`; 1 of 603 active cgroup periods was throttled (`0.17%`). Recovery memory was `4.2%` above the pre-burst level. PID 1 settled at `1,324–1,400 kB` RSS with a `5,512 kB` HWM.

## Run 2

Pod `tadoku-paper-styleguide-7b8bd66658-g6jk4`; pod start to Ready 4,000 ms; container start to Ready 2,000 ms; zero restarts and warning events.

```csv
label,memory.current,memory.peak,cpu.usage_usec,cpu.nr_periods,cpu.nr_throttled,cpu.throttled_usec
ready,8101888,10887168,99756,5,0,0
idle_00,8208384,10887168,132753,8,1,56152
idle_05,4894720,10887168,183111,31,2,96149
idle_10,6230016,10887168,297261,62,2,96149
idle_15,4481024,10887168,342509,83,3,168303
cold_catalogue_before,4255744,10887168,376659,85,4,169953
cold_catalogue_after,8933376,11313152,418185,91,5,227417
deep_links_before,8855552,11313152,452630,93,6,268603
deep_links_after,8843264,11313152,488367,104,7,310373
navigation_before,8871936,11313152,522785,106,8,362738
navigation_after,8601600,11313152,560550,112,8,362738
sustained_before,8548352,11313152,593785,114,8,362738
sustained_after,4366336,11313152,699412,246,9,378618
burst_before,4472832,11313152,730803,248,10,395829
burst_after,4894720,11313152,3332803,851,10,395829
recovery_05,4194304,11313152,3378665,874,11,429848
```

| Scenario | Requests | Failures | Average latency | Maximum latency |
| --- | ---: | ---: | ---: | ---: |
| Cold catalogue | 11 | 0 | 31.5 ms | 54.0 ms |
| Canonical deep links | 40 | 0 | 21.4 ms | 23.7 ms |
| Representative navigation | 20 | 0 | 21.4 ms | 23.4 ms |
| Sustained | 300 | 0 | 32.0 ms | 578.0 ms |
| Burst | 16,509 | 0 | 32.0 ms | 3,111.4 ms |

Sustained CPU averaged `0.336m`. Burst CPU averaged `42.7m`; 0 of 603 active cgroup periods was throttled. Recovery memory was `6.2%` below the pre-burst level. PID 1 settled at `1,328 kB` RSS with a `5,524 kB` HWM.

The first p99 sampler revision had a shell-quoting defect. It did not affect traffic or the authoritative boundary counters and its output is excluded.

## Run 3

Pod `tadoku-paper-styleguide-c948b7f95-x5clw`; pod start to Ready 3,000 ms; container start to Ready 1,000 ms; zero restarts and warning events.

```csv
label,memory.current,memory.peak,cpu.usage_usec,cpu.nr_periods,cpu.nr_throttled,cpu.throttled_usec
ready,8458240,10596352,104560,5,2,89784
idle_00,8642560,10850304,140816,7,2,89784
idle_05,4227072,10850304,184936,28,3,133889
idle_10,4169728,10850304,228527,47,4,175728
idle_15,4288512,10850304,271768,68,5,200377
cold_catalogue_before,4304896,10850304,307113,70,6,248377
cold_catalogue_after,8990720,11395072,347006,78,7,313171
deep_links_before,8663040,11395072,380605,80,8,316536
deep_links_after,8753152,11395072,417740,90,8,316536
navigation_before,8687616,11395072,449448,92,9,353994
navigation_after,8130560,11395072,485674,98,10,423211
sustained_before,8056832,11395072,518436,100,11,464516
sustained_after,4448256,11395072,624882,231,12,504515
burst_before,4546560,11395072,659756,233,12,504515
burst_after,4620288,11395072,3441610,834,13,564515
recovery_05,4210688,11395072,3494498,856,14,604513
```

| Scenario | Requests | Failures | Average latency | Maximum latency |
| --- | ---: | ---: | ---: | ---: |
| Cold catalogue | 11 | 0 | 75.8 ms | 525.6 ms |
| Canonical deep links | 40 | 0 | 23.0 ms | 43.2 ms |
| Representative navigation | 20 | 0 | 25.4 ms | 85.9 ms |
| Sustained | 300 | 0 | 34.3 ms | 499.7 ms |
| Burst | 15,995 | 0 | 33.2 ms | 5,183.3 ms |

Sustained CPU averaged `0.341m`. Burst CPU averaged `46.4m`; 1 of 601 active cgroup periods was throttled (`0.17%`). A valid 27-second one-second sampler measured `52.626m` p99 and zero throttling; the Kubernetes API server reset the long-running exec stream before 60 seconds. Recovery memory was `7.4%` below the pre-burst level. PID 1 settled at `1,324 kB` RSS with a `5,568 kB` HWM.

## Observer-neutral navigation controls

The normal before/after snapshots execute inside the measured cgroup. Three additional repetitions established the baseline after the observer became idle, issued the same 20 public navigation requests, then read the ending counter with shell builtins.

| Repetition | Requests | Failures | CPU use | Active periods | Throttled periods |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 | 0 | 4,944 µs | 8 | 0 |
| 2 | 20 | 0 | 3,908 µs | 6 | 0 |
| 3 | 20 | 0 | 4,317 µs | 6 | 0 |

## Aggregates

| Measurement | Median | Worst |
| --- | ---: | ---: |
| Container start to Ready | 2.0 s | 2.0 s |
| Settled idle CPU | 0.160m | 0.379m |
| Sustained CPU | 0.336m | 0.341m |
| Burst average CPU | 42.7m | 46.4m |
| Settled idle memory | 4.45 MB | 6.23 MB |
| Representative-navigation memory | 8.60 MB | 9.11 MB |
| Working peak memory | 11.40 MB | 11.48 MB |
| Burst throttling | 0.17% | 0.17% |
| Recovery delta | −6.2% | +4.2% |

Across the three primary runs, 49,456 responses completed with zero failure. The observer-neutral controls add 60 successful navigation responses.
