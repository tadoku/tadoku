#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo 'Usage: publish-image.sh <webv2|auth|admin|styleguide> --repository <repo> --tag <tag> [--tag <tag> ...]' >&2
  exit 2
}

app=${1:-}
case "$app" in webv2|auth|admin|styleguide) ;; *) usage ;; esac
shift
repository=
tags=()
while (($#)); do
  case "$1" in
    --repository) (($# >= 2)) || usage; repository=$2; shift 2 ;;
    --tag) (($# >= 2)) || usage; tags+=("$2"); shift 2 ;;
    *) usage ;;
  esac
done
[[ "$repository" =~ ^[a-z0-9][a-z0-9.-]*(:[0-9]+)?(/[a-z0-9]+([._-][a-z0-9]+)*)+$ ]] || usage
((${#tags[@]})) || usage
for tag in "${tags[@]}"; do
  [[ "$tag" =~ ^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127}$ ]] || usage
done

root=${BUILD_WORKSPACE_DIRECTORY:-$(git rev-parse --show-toplevel)}
image="$repository:${tags[0]}"
git -C "$root" archive --format=tar HEAD:frontend |
  docker build --platform linux/amd64 --build-arg "PROJECT_NAME=$app" -t "$image" -
for tag in "${tags[@]}"; do
  if [[ "$repository:$tag" != "$image" ]]; then
    docker tag "$image" "$repository:$tag"
  fi
  docker push "$repository:$tag"
done
