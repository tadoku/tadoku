#!/usr/bin/env bash

set -euo pipefail

context="${KUBE_CONTEXT:-lke20948-ctx}"
namespace="${KUBE_NAMESPACE:-tdk-prod-paper-styleguide}"
deployment="${KUBE_DEPLOYMENT:-tadoku-paper-styleguide}"
base_url="${PAPER_BASE_URL:-https://paper.tadoku.app}"
run="${1:?usage: measure-paper-resources.sh RUN_NUMBER}"

repo_root="$(git rev-parse --show-toplevel)"
catalog_dir="$repo_root/frontend/packages/paper-ui/src/catalog"
dist_dir="$repo_root/frontend/apps/paper-styleguide/dist"

snapshot() {
  local label="$1"
  local pod="$2"

  printf 'SNAPSHOT_BEGIN label=%s timestamp=%s pod=%s\n' "$label" "$(date -u +%FT%TZ)" "$pod"
  kubectl --context "$context" --namespace "$namespace" exec "$pod" -- sh -c '
    printf "memory.current="; cat /sys/fs/cgroup/memory.current
    printf "memory.peak="; cat /sys/fs/cgroup/memory.peak
    sed "s/^/memory.events./" /sys/fs/cgroup/memory.events
    sed "s/^/cpu.stat./" /sys/fs/cgroup/cpu.stat
    awk "/VmRSS|VmHWM/ {print \"pid1.\" \$1 \"=\" \$2 \" \" \$3}" /proc/1/status
  '
  printf 'SNAPSHOT_END label=%s\n' "$label"
}

latest_ready_pod() {
  kubectl --context "$context" --namespace "$namespace" get pods \
    -l 'app=tadoku-paper-styleguide' \
    --sort-by=.metadata.creationTimestamp \
    -o json | jq -r '[.items[] | select(any(.status.conditions[]?; .type == "Ready" and .status == "True"))][-1].metadata.name // empty'
}

wait_for_new_ready_pod() {
  local old_pod="$1"
  local deadline=$((SECONDS + 180))
  local candidate=""

  while (( SECONDS < deadline )); do
    candidate="$(latest_ready_pod)"
    if [[ -n "$candidate" && "$candidate" != "$old_pod" ]]; then
      printf '%s\n' "$candidate"
      return 0
    fi
    sleep 1
  done

  printf 'new ready pod did not appear within 180 seconds\n' >&2
  return 1
}

request_once() {
  local path="$1"
  curl --silent --show-error --output /dev/null \
    --write-out '%{http_code} %{time_total}\n' \
    --header 'Cache-Control: no-cache' \
    "$base_url$path"
}

summarize_requests() {
  local label="$1"
  awk -v label="$label" '
    { count += 1; total += $2; if ($1 < 200 || $1 >= 400) failures += 1; if ($2 > maximum) maximum = $2 }
    END { printf "TRAFFIC label=%s requests=%d failures=%d latency_avg_seconds=%.6f latency_max_seconds=%.6f\n", label, count, failures, count ? total / count : 0, maximum }
  '
}

mapfile -t canonical_routes < <(
  rg --no-filename --only-matching 'route: "[^"]+"' "$catalog_dir" \
    | sed -E 's/route: "([^"]+)"/\1/' \
    | sort -u
)

mapfile -t static_paths < <(
  find "$dist_dir" -type f -printf '/%P\n' | sort
)

printf 'RUN_BEGIN run=%s timestamp=%s commit=%s\n' "$run" "$(date -u +%FT%TZ)" "$(git rev-parse origin/main)"
printf 'ENV context=%s namespace=%s deployment=%s base_url=%s node_arch=%s\n' \
  "$context" "$namespace" "$deployment" "$base_url" "$(uname -m)"
printf 'ROUTES canonical=%s static_files=%s\n' "${#canonical_routes[@]}" "${#static_paths[@]}"

old_pod="$(latest_ready_pod)"
rollout_started_epoch="$(date +%s%3N)"
kubectl --context "$context" --namespace "$namespace" rollout restart "deployment/$deployment"
pod="$(wait_for_new_ready_pod "$old_pod")"
rollout_ready_epoch="$(date +%s%3N)"

pod_json="$(kubectl --context "$context" --namespace "$namespace" get pod "$pod" -o json)"
pod_started="$(jq -r '.status.startTime' <<<"$pod_json")"
container_started="$(jq -r '.status.containerStatuses[0].state.running.startedAt' <<<"$pod_json")"
pod_ready="$(jq -r '.status.conditions[] | select(.type == "Ready") | .lastTransitionTime' <<<"$pod_json")"
started_epoch="$(date -u -d "$pod_started" +%s%3N)"
container_started_epoch="$(date -u -d "$container_started" +%s%3N)"
ready_epoch="$(date -u -d "$pod_ready" +%s%3N)"
printf 'STARTUP pod=%s previous_pod=%s rollout_milliseconds=%s pod_ready_milliseconds=%s container_ready_milliseconds=%s pod_started=%s container_started=%s ready=%s image_id=%s\n' \
  "$pod" "$old_pod" "$((rollout_ready_epoch - rollout_started_epoch))" "$((ready_epoch - started_epoch))" \
  "$((ready_epoch - container_started_epoch))" "$pod_started" "$container_started" "$pod_ready" \
  "$(jq -r '.status.containerStatuses[0].imageID' <<<"$pod_json")"

snapshot ready "$pod"
snapshot idle_00 "$pod"
sleep 300
snapshot idle_05 "$pod"
sleep 300
snapshot idle_10 "$pod"
sleep 300
snapshot idle_15 "$pod"

snapshot cold_catalogue_before "$pod"
{
  request_once /index.html
  for path in "${static_paths[@]}"; do
    request_once "$path"
  done
} | summarize_requests cold_catalogue
snapshot cold_catalogue_after "$pod"

snapshot deep_links_before "$pod"
for path in "${canonical_routes[@]}"; do
  request_once "$path"
done | summarize_requests deep_links
snapshot deep_links_after "$pod"

navigation_routes=(
  /foundations/color
  /components/actions/button
  /components/forms/input
  /components/navigation/navbar
  /components/navigation/sidebar
  /components/navigation/breadcrumb
  /components/navigation/tabbar
  /components/navigation/vertical-tabbar
  /components/navigation/pagination
  /components/data-display/table
  /components/data-display/heatmap-chart
  /components/feedback/flash
  /components/feedback/toast
  /components/overlays/modal
  /patterns/logging
  /experiments/logging-v2
  /foundations/typography
  /foundations/layout
  /contributing
  /changelog
)

snapshot navigation_before "$pod"
for path in "${navigation_routes[@]}"; do
  request_once "$path"
done | summarize_requests navigation
snapshot navigation_after "$pod"

traffic_routes=(/ /components/actions/button /components/data-display/table /patterns/logging)
snapshot sustained_before "$pod"
{
  for ((request_index = 0; request_index < 300; request_index += 1)); do
    request_once "${traffic_routes[$((request_index % ${#traffic_routes[@]}))]}"
    sleep 1
  done
} | summarize_requests sustained
snapshot sustained_after "$pod"

burst_dir="$(mktemp -d)"
trap 'rm -rf -- "$burst_dir"' EXIT
snapshot burst_before "$pod"
for ((segment = 0; segment < 3; segment += 1)); do
  (
    sleep "$((segment * 20))"
    kubectl --context "$context" --namespace "$namespace" exec "$pod" -- sh -c '
      previous_usage="$(grep "^usage_usec " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
      previous_periods="$(grep "^nr_periods " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
      previous_throttled="$(grep "^nr_throttled " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
      sample=0
      while [ "$sample" -lt 20 ]; do
        sleep 1
        usage="$(grep "^usage_usec " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
        periods="$(grep "^nr_periods " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
        throttled="$(grep "^nr_throttled " /sys/fs/cgroup/cpu.stat | cut -d " " -f 2)"
        printf "%s %s %s\n" "$((usage - previous_usage))" "$((periods - previous_periods))" "$((throttled - previous_throttled))"
        previous_usage="$usage"
        previous_periods="$periods"
        previous_throttled="$throttled"
        sample=$((sample + 1))
      done
    '
  ) >"$burst_dir/cpu-samples-$segment" &
done
for ((worker = 0; worker < 10; worker += 1)); do
  worker_path="${traffic_routes[$((worker % ${#traffic_routes[@]}))]}"
  timeout 60 bash -c '
    while true; do
      curl --silent --show-error --output /dev/null \
        --write-out "%{http_code} %{time_total}\\n" \
        "$1$2"
    done
  ' _ "$base_url" "$worker_path" >"$burst_dir/worker-$worker" &
done
wait || true
cat "$burst_dir"/worker-* | summarize_requests burst
sort -n "$burst_dir"/cpu-samples-* | awk '
  { usage[NR] = $1; periods += $2; throttled += $3 }
  END {
    percentile = int((NR * 99 + 99) / 100)
    printf "BURST_CPU samples=%d p99_millicores=%.3f periods=%d throttled=%d throttled_percent=%.3f\n", NR, usage[percentile] / 1000, periods, throttled, periods ? throttled * 100 / periods : 0
  }
'
snapshot burst_after "$pod"

sleep 300
snapshot recovery_05 "$pod"

final_json="$(kubectl --context "$context" --namespace "$namespace" get pod "$pod" -o json)"
printf 'POD_FINAL restarts=%s ready=%s phase=%s\n' \
  "$(jq -r '.status.containerStatuses[0].restartCount' <<<"$final_json")" \
  "$(jq -r '.status.containerStatuses[0].ready' <<<"$final_json")" \
  "$(jq -r '.status.phase' <<<"$final_json")"
kubectl --context "$context" --namespace "$namespace" get events \
  --field-selector "involvedObject.name=$pod,type=Warning" \
  -o json | jq -c '{warning_events: [.items[] | {reason, message, count}]}'
printf 'RUN_END run=%s timestamp=%s\n' "$run" "$(date -u +%FT%TZ)"
