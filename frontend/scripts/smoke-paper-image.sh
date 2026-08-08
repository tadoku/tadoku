#!/bin/sh
set -eu

image_ref="$1"
container_name="paper-smoke-$$"
headers_file="$(mktemp)"

cleanup() {
  docker stop "$container_name" >/dev/null 2>&1 || true
  rm -f "$headers_file"
}
trap cleanup EXIT INT TERM

docker run --rm --name "$container_name" -d -p 127.0.0.1::8080 "$image_ref" >/dev/null
port="$(docker port "$container_name" 8080/tcp | sed -n 's/.*://p' | head -n 1)"

attempt=0
until curl --fail --silent "http://127.0.0.1:${port}/healthz" >/dev/null; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    docker logs "$container_name"
    exit 1
  fi
  sleep 1
done

index_html="$(curl --fail --silent "http://127.0.0.1:${port}/")"
printf '%s' "$index_html" | grep -q '<div id="root"'
curl --fail --silent "http://127.0.0.1:${port}/components/actions/button" | grep -q '<div id="root"'
curl --fail --silent --compressed --dump-header "$headers_file" \
  "http://127.0.0.1:${port}/" >/dev/null
grep -qi '^cache-control: no-cache' "$headers_file"

asset_path="$(printf '%s' "$index_html" | sed -n 's/.*src="\([^"]*\.js\)".*/\1/p' | head -n 1)"
test -n "$asset_path"
curl --fail --silent --header 'Accept-Encoding: gzip' --dump-header "$headers_file" \
  "http://127.0.0.1:${port}${asset_path}" >/dev/null
grep -qi '^content-encoding: gzip' "$headers_file"
grep -qi '^cache-control: public, max-age=31536000, immutable' "$headers_file"

missing_status="$(curl --silent --output /dev/null --write-out '%{http_code}' "http://127.0.0.1:${port}/assets/not-a-real-paper-asset.js")"
test "$missing_status" = "404"

for sensitive_path in '/.env' '/.git/config' '/api/.env' '/config.env'; do
  sensitive_status="$(curl --silent --output /dev/null --write-out '%{http_code}' "http://127.0.0.1:${port}${sensitive_path}")"
  test "$sensitive_status" = "404"
done

echo "Paper static image smoke passed on port ${port}."
