#!/usr/bin/env bash
# Deletes old versions of the production-branch images under ghcr.io/tadoku/tadoku/branches/.
# Lists what it would delete unless DRY_RUN=false. Needs GH_TOKEN with read:packages, and
# delete:packages (or the repository's admin role on each package) to delete.
set -euo pipefail

RETENTION_DAYS="${RETENTION_DAYS:-14}"
QUIET_HOURS="${QUIET_HOURS:-24}"
DRY_RUN="${DRY_RUN:-true}"
IMAGES=(tadoku-api tadoku-worker frontend-webv2 frontend-auth frontend-admin tenant)
MANIFEST_TYPES="application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json,\
application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json"

cutoff="$(date -u -d "-${RETENTION_DAYS} days" +%Y-%m-%dT%H:%M:%SZ)"
quiet_since="$(date -u -d "-${QUIET_HOURS} hours" +%Y-%m-%dT%H:%M:%SZ)"
action="would delete"
if [[ "$DRY_RUN" == false ]]; then
  action="deleting"
fi
echo "retention cutoff $cutoff, skipping packages pushed since $quiet_since, dry run: $DRY_RUN"

for image in "${IMAGES[@]}"; do
  package="/orgs/tadoku/packages/container/tadoku%2Fbranches%2F${image}"
  repository="tadoku/tadoku/branches/${image}"

  # Pushing a new tag onto an existing digest bumps only the package's updated_at, not the
  # version's, so an old version may back a live branch until the package has been quiet.
  updated="$(gh api "$package" --jq .updated_at)"
  if [[ ! "$updated" < "$quiet_since" ]]; then
    echo "$image: skipped, last push $updated"
    continue
  fi

  versions="$(gh api --paginate --slurp "$package/versions?per_page=100" | jq 'add | sort_by(.updated_at) | reverse')"
  keep="$(jq -r --arg cutoff "$cutoff" \
    '(map(select(.metadata.container.tags != [])) | .[0].name // empty), (.[] | select(.updated_at >= $cutoff) | .name)' \
    <<<"$versions" | sort -u)"

  # Kept image indexes still need their platform and attestation manifests.
  token="$(curl -fsS "https://ghcr.io/token?scope=repository:${repository}:pull" | jq -r .token)"
  children="$(for digest in $keep; do
    curl -fsS -H "Authorization: Bearer $token" -H "Accept: $MANIFEST_TYPES" \
      "https://ghcr.io/v2/${repository}/manifests/${digest}" | jq -r '.manifests[]?.digest'
  done)"
  keep_json="$(printf '%s\n' "$keep" "$children" | jq -R . | jq -sc .)"

  # Tagged versions go first so no index outlives the manifests it references.
  deletions="$(jq -r --argjson keep "$keep_json" '
    map(select(.name | IN($keep[]) | not))
    | sort_by(.metadata.container.tags == [])
    | .[]
    | [.id, .name, .updated_at, (.metadata.container.tags | join(",") | if . == "" then "untagged" else . end)]
    | @tsv' <<<"$versions")"

  echo "$image: $(jq length <<<"$versions") versions, $(grep -c . <<<"$deletions" || true) to delete"
  while IFS=$'\t' read -r id digest version_updated tags; do
    [[ -n "$id" ]] || continue
    echo "$image: $action $digest updated $version_updated tags $tags"
    if [[ "$DRY_RUN" == false ]]; then
      gh api --method DELETE "$package/versions/$id" --silent
    fi
  done <<<"$deletions"
done
