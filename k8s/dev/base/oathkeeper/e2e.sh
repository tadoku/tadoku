#!/usr/bin/env bash
set -euo pipefail

# Repeatable isolated gateway proof:
#   k8s/dev/base/oathkeeper/e2e.sh /tmp/tadoku-oathkeeper-evidence
#   OATHKEEPER_IMAGE=docker.io/oryd/oathkeeper:v25.4.0 k8s/dev/base/oathkeeper/e2e.sh /tmp/tadoku-oathkeeper-v25-evidence
# Requires Bazel, Docker, Git, Node, curl, and OpenSSL. It runs the real
# Oathkeeper, token-reflector, and Flipt processes. Synthetic Kubernetes-style
# JWTs and a local JWKS server replace cluster TokenRequest issuance and the
# Kubernetes JWKS relay; HTTP requests replace the separately tested Go client.

root=$(git rev-parse --show-toplevel)
evidence=${1:-/tmp/tadoku-oathkeeper-evidence}
work=$(mktemp -d)
prefix="tdk-oathkeeper-proof-$$"
oathkeeper_image=${OATHKEEPER_IMAGE:-docker.io/oryd/oathkeeper:v26.2.0}
containers=()

cleanup() {
  for container in "${containers[@]}"; do docker logs "$container" >"$evidence/${container#"$prefix-"}.log" 2>&1 || true; docker rm -f "$container" >/dev/null 2>&1 || true; done
  docker network rm "$prefix" >/dev/null 2>&1 || true
  rm -rf "$work"
}
trap cleanup EXIT
mkdir -p "$evidence"

bazel run //services/token-reflector:load >"$evidence/build-reflector.log" 2>&1
docker pull "$oathkeeper_image" >"$evidence/pull-oathkeeper.log"
docker pull docker.io/node:22-alpine >"$evidence/pull-node.log"
docker pull docker.flipt.io/flipt/flipt:v2.11.0@sha256:d20384874048ef6ac326f4937cee64f1db175a1878a87db32916cc8db46c740e >"$evidence/pull-flipt.log"

node - "$work" <<'NODE'
const { generateKeyPairSync, createSign } = require('node:crypto')
const fs = require('node:fs')
const path = require('node:path')
const out = process.argv[2]
const { privateKey, publicKey } = generateKeyPairSync('rsa', { modulusLength: 2048 })
const pub = publicKey.export({ format: 'jwk' }); pub.kid = 'fixture'; pub.use = 'sig'; pub.alg = 'RS256'
const priv = privateKey.export({ format: 'jwk' }); priv.kid = 'fixture'; priv.use = 'sig'; priv.alg = 'RS256'
fs.writeFileSync(path.join(out, 'jwks.json'), JSON.stringify({ keys: [pub] }))
fs.writeFileSync(path.join(out, 'signing.json'), JSON.stringify({ keys: [priv] }))
function token(sub, aud) {
  const encode = value => Buffer.from(JSON.stringify(value)).toString('base64url')
  const data = `${encode({ alg: 'RS256', typ: 'JWT', kid: 'fixture' })}.${encode({ sub, aud: [aud], exp: Math.floor(Date.now()/1000)+3600, iat: Math.floor(Date.now()/1000) })}`
  return `${data}.${createSign('RSA-SHA256').update(data).sign(privateKey).toString('base64url')}`
}
fs.writeFileSync(path.join(out, 'valid.token'), token('system:serviceaccount:tdk-dev-tadoku-api:tadoku-api', 'tadoku-api'))
const valid = token('system:serviceaccount:tdk-dev-tadoku-api:tadoku-api', 'tadoku-api')
const pieces = valid.split('.'); pieces[2] = (pieces[2][0] === 'A' ? 'B' : 'A') + pieces[2].slice(1)
fs.writeFileSync(path.join(out, 'bad-signature.token'), pieces.join('.'))
fs.writeFileSync(path.join(out, 'wrong-name.token'), token('system:serviceaccount:tdk-dev-tadoku-api:other', 'tadoku-api'))
fs.writeFileSync(path.join(out, 'wrong-namespace.token'), token('system:serviceaccount:other:tadoku-api', 'tadoku-api'))
fs.writeFileSync(path.join(out, 'wrong-audience.token'), token('system:serviceaccount:tdk-dev-tadoku-api:tadoku-api', 'other'))
fs.writeFileSync(path.join(out, 'legacy.token'), token('system:serviceaccount:tdk-dev-immersion-api:immersion-api', 'immersion-api'))
NODE

mkdir -p "$work/fixture" "$work/config"
cp "$work/jwks.json" "$work/fixture/jwks.json"
cat >"$work/fixture/fixture.mjs" <<'NODE'
import http from 'node:http'; import fs from 'node:fs'
const jwks = fs.readFileSync('/fixture/jwks.json')
http.createServer((req,res) => { if (req.url === '/jwks') { res.setHeader('content-type','application/json'); res.end(jwks) } else { res.statusCode=404; res.end() } }).listen(8080)
NODE

cat >"$work/oathkeeper.yaml" <<'YAML'
serve:
  proxy: {host: 0.0.0.0, port: 4455}
  api: {host: 0.0.0.0, port: 4456}
access_rules: {matching_strategy: glob, repositories: [file:///config/rules.yaml]}
errors: {fallback: [json], handlers: {json: {enabled: true, config: {verbose: true}}}}
authenticators:
  jwt: {enabled: true, config: {jwks_urls: [http://fixture:8080/jwks]}}
authorizers:
  allow: {enabled: true}
  remote_json: {enabled: true, config: {remote: http://127.0.0.1, payload: '{}'}}
mutators:
  noop: {enabled: true}
  id_token:
    enabled: true
    config: {issuer_url: http://oathkeeper-api/, jwks_url: file:///config/signing.json}
YAML

awk '
  /^- id: tadoku:flipt:evaluation:snapshot$/ {copy=1}
  /^- id: tadoku:flipt:management:feature-access:get$/ {copy=1}
  /^- id: tadoku:flipt:management:feature-access:update$/ {copy=1}
  /^- id: s2s:immersion-to-flipt-evaluation$/ {copy=1}
  /^- id: s2s:tadoku-to-flipt-evaluation$/ {copy=1}
  /^- id: s2s:tadoku-to-flipt-management$/ {copy=1}
  /^- id:/ && copy && $0 !~ /(evaluation:snapshot|feature-access:get|feature-access:update|immersion-to-flipt-evaluation|tadoku-to-flipt-evaluation|tadoku-to-flipt-management)$/ {copy=0}
  copy {print}
' "$root/k8s/dev/base/oathkeeper/rules.yaml" |
  sed -e 's#http://oathkeeper-proxy.tdk-dev-oathkeeper:4455#http://oathkeeper:4455#g' \
      -e 's#http://oathkeeper-api.tdk-dev-oathkeeper:4456#http://oathkeeper:4456#g' \
      -e 's#http://token-reflector.tdk-dev-token-reflector/jwks#http://fixture:8080/jwks#g' \
      -e 's#http://token-reflector.tdk-dev-token-reflector#http://reflector:8080#g' \
      -e 's#http://flipt.tdk-dev-flipt.svc.cluster.local:8080#http://flipt:8080#g' >"$work/rules.yaml"
cp "$work/oathkeeper.yaml" "$work/rules.yaml" "$work/signing.json" "$work/config/"

mkdir -p "$work/flipt/default" "$work/flipt-config"
cp "$root/k8s/dev/base/flipt/features.yaml" "$work/flipt/default/features.yaml"
git -C "$work/flipt" init --initial-branch=main >/dev/null
git -C "$work/flipt" -c user.name=fixture -c user.email=fixture@example.test add default/features.yaml
git -C "$work/flipt" -c user.name=fixture -c user.email=fixture@example.test commit -m fixture >/dev/null
cat >"$work/flipt-config/default.yml" <<'YAML'
version: "2.0"
server: {protocol: http, host: 0.0.0.0, http_port: 8080, grpc_port: 9000}
storage: {local: {name: local, branch: main, backend: {type: local, path: /var/lib/flipt}}}
environments: {local: {name: local, default: true, storage: local}}
meta: {check_for_updates: false, telemetry_enabled: false, state_directory: /tmp/flipt}
YAML

docker network create "$prefix" >/dev/null
create() { docker create --name "$prefix-$1" --network "$prefix" --network-alias "$1" --cpus 0.5 --memory 512m "${@:2}" >/dev/null; containers+=("$prefix-$1"); }
create fixture docker.io/node:22-alpine node /fixture/fixture.mjs
docker cp "$work/fixture" "$prefix-fixture:/fixture"
create reflector bazel/services/token-reflector:latest
create flipt --user 0 --entrypoint /flipt docker.flipt.io/flipt/flipt:v2.11.0@sha256:d20384874048ef6ac326f4937cee64f1db175a1878a87db32916cc8db46c740e server --config /etc/flipt/default.yml
docker cp "$work/flipt/." "$prefix-flipt:/var/lib/flipt"; docker cp "$work/flipt-config/default.yml" "$prefix-flipt:/etc/flipt/default.yml"
create oathkeeper -p 127.0.0.1::4455 -p 127.0.0.1::4456 --entrypoint oathkeeper "$oathkeeper_image" serve --config /config/oathkeeper.yaml
docker cp "$work/config" "$prefix-oathkeeper:/config"
for container in "${containers[@]}"; do docker start "$container" >/dev/null; done
port=$(docker port "$prefix-oathkeeper" 4455/tcp | sed 's/.*://')
api_port=$(docker port "$prefix-oathkeeper" 4456/tcp | sed 's/.*://')
base="http://127.0.0.1:$port"
ready=false
for _ in $(seq 1 60); do
  if curl --max-time 2 -fsS "http://127.0.0.1:$api_port/health/ready" >/dev/null 2>&1 && docker exec "$prefix-fixture" node -e "fetch('http://flipt:8080/health',{signal:AbortSignal.timeout(2000)}).then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"; then ready=true; break; fi
  sleep 1
done
test "$ready" = true || { echo 'Oathkeeper or Flipt did not become ready'; exit 1; }

request() { local want=$1 token=$2 path=$3 out=$4; status=$(curl -sS -o "$out" -w '%{http_code}' -H 'Host: oathkeeper:4455' -H "Authorization: Bearer $token" "$base$path"); test "$status" = "$want" || { cat "$out"; echo "status $status, want $want"; exit 1; }; }
request 401 invalid /token-exchange/flipt-evaluation/tadoku-api "$evidence/invalid-credential.json"
request 401 "$(cat "$work/bad-signature.token")" /token-exchange/flipt-evaluation/tadoku-api "$evidence/bad-signature.json"
request 403 "$(cat "$work/wrong-name.token")" /token-exchange/flipt-evaluation/tadoku-api "$evidence/wrong-name.json"
request 403 "$(cat "$work/wrong-namespace.token")" /token-exchange/flipt-evaluation/tadoku-api "$evidence/wrong-namespace.json"
request 401 "$(cat "$work/wrong-audience.token")" /token-exchange/flipt-evaluation/tadoku-api "$evidence/wrong-audience.json"
request 200 "$(cat "$work/valid.token")" /token-exchange/flipt-evaluation/tadoku-api "$evidence/evaluation-exchange.json"
request 200 "$(cat "$work/valid.token")" /token-exchange/flipt-management/tadoku-api "$evidence/management-exchange.json"
request 200 "$(cat "$work/legacy.token")" /token-exchange/flipt-evaluation/immersion-api "$evidence/legacy-exchange.json"
access_token() { node -e 'process.stdout.write(JSON.parse(require("node:fs").readFileSync(process.argv[1])).access_token)' "$1"; }
evaluation=$(access_token "$evidence/evaluation-exchange.json"); management=$(access_token "$evidence/management-exchange.json"); legacy=$(access_token "$evidence/legacy-exchange.json")
request 200 "$evaluation" /flipt/internal/v1/evaluation/snapshot/namespace/default "$evidence/snapshot.json"
request 200 "$management" /flipt-management/api/v2/environments/local/namespaces/default/resources/flipt.core.Segment/release-log-entry-v2-access "$evidence/segment.json"
node - "$evidence/segment.json" "$work/segment-update.json" <<'NODE'
const fs = require('node:fs'); const segment = JSON.parse(fs.readFileSync(process.argv[2])); segment.resource.payload.constraints[0].value = '["11111111-1111-4111-8111-111111111111"]'; fs.writeFileSync(process.argv[3], JSON.stringify({key: segment.resource.key, revision: segment.revision, payload: segment.resource.payload}))
NODE
status=$(curl -sS -o "$evidence/evaluation-cannot-update.json" -w '%{http_code}' -X PUT -H 'Host: oathkeeper:4455' -H 'Content-Type: application/json' -H "Authorization: Bearer $evaluation" --data-binary "@$work/segment-update.json" "$base/flipt-management/api/v2/environments/local/namespaces/default/resources")
test "$status" = 401 || { cat "$evidence/evaluation-cannot-update.json"; echo "evaluation update status $status, want 401"; exit 1; }
status=$(curl -sS -o "$evidence/segment-update.json" -w '%{http_code}' -X PUT -H 'Host: oathkeeper:4455' -H 'Content-Type: application/json' -H "Authorization: Bearer $management" --data-binary "@$work/segment-update.json" "$base/flipt-management/api/v2/environments/local/namespaces/default/resources")
test "$status" = 200 || { cat "$evidence/segment-update.json"; echo "management update status $status, want 200"; exit 1; }
request 200 "$management" /flipt-management/api/v2/environments/local/namespaces/default/resources/flipt.core.Segment/release-log-entry-v2-access "$evidence/segment-readback.json"
grep -q '11111111-1111-4111-8111-111111111111' "$evidence/segment-readback.json"
request 401 "$evaluation" /flipt-management/api/v2/environments/local/namespaces/default/resources/flipt.core.Segment/release-log-entry-v2-access "$evidence/evaluation-cannot-manage.json"
request 401 "$management" /flipt/internal/v1/evaluation/snapshot/namespace/default "$evidence/management-cannot-evaluate.json"
request 200 "$legacy" /flipt/internal/v1/evaluation/snapshot/namespace/default "$evidence/legacy-snapshot.json"

git -C "$root" rev-parse HEAD >"$evidence/revision.txt"
git -C "$root" diff --binary -- .dev/tadoku-api.yaml docs/docs/services/s2s-auth.md infra/dev/ory/access_rules.yaml k8s/dev/base/oathkeeper k8s/dev/base/services/tadoku-api.yaml services/tadoku-api/deployments/api.yaml services/token-reflector >"$evidence/source.diff"
sha256sum "$evidence/source.diff" >"$evidence/source.sha256"
sha256sum "$root/k8s/dev/base/oathkeeper/e2e.sh" >>"$evidence/source.sha256"
for image in "$oathkeeper_image" docker.io/node:22-alpine docker.flipt.io/flipt/flipt:v2.11.0@sha256:d20384874048ef6ac326f4937cee64f1db175a1878a87db32916cc8db46c740e bazel/services/token-reflector:latest; do docker image inspect "$image" --format '{{json .RepoDigests}} {{.Id}}'; done >"$evidence/images.txt"
printf 'passed\n' >"$evidence/result.txt"
echo "passed: $evidence"
