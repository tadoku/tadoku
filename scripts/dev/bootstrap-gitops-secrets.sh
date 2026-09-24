#!/usr/bin/env bash
# DEVELOPMENT ONLY. Run after explicit credential/bootstrap approval.
# Never copy credentials from the old Tilt stack or any production namespace.
set -euo pipefail
set +x

kube() { kubectl --context homelab-dev "$@"; }
test "$(kube config view --minify -o jsonpath='{.clusters[0].cluster.server}')" = https://192.168.1.190:6443

for namespace in tdk-dev-data tdk-dev-kratos tdk-dev-keto tdk-dev-oathkeeper tdk-dev-tadoku-api; do
  test "$(kube get namespace "$namespace" -o jsonpath='{.metadata.labels.app\.kubernetes\.io/part-of}')" = tadoku-dev
done

check_secret() {
  local namespace="$1" name="$2" ownership="$3"
  shift 3
  local existing
  existing=$(kube -n "$namespace" get secret "$name" --ignore-not-found -o json)
  if [ -z "$existing" ]; then
    if [ "$ownership" = source ]; then
      echo "Missing source Secret: $namespace/$name" >&2
      return 1
    fi
    return
  fi
  if ! printf '%s' "$existing" | jq -e --arg ownership "$ownership" --args '
    . as $secret |
    .type == "Opaque" and
    ($ownership == "source" or .metadata.labels["app.kubernetes.io/part-of"] == "tadoku-dev") and
    all($ARGS.positional[]; $secret.data[.] | type == "string" and length > 0 and (@base64d | length > 0))
  ' "$@" >/dev/null 2>&1; then
    echo "Refusing bootstrap: $namespace/$name has unexpected ownership or missing/invalid required data keys." >&2
    return 1
  fi
}

# Validate all existing destinations before the first write, including runtime
# keys we preserve. Namespace ownership alone does not authorize replacing a Secret.
check_secret tdk-dev-tadoku-api tadoku.tadoku-dev-db.credentials.postgresql.acid.zalan.do destination username password
check_secret tdk-dev-data tadoku.tadoku-dev-db.credentials.postgresql.acid.zalan.do source username password
for provider in kratos keto; do
  check_secret tdk-dev-data "$provider.tadoku-dev-db.credentials.postgresql.acid.zalan.do" source username password
  check_secret "tdk-dev-$provider" "$provider.tadoku-dev-db.credentials.postgresql.acid.zalan.do" destination username password
done
check_secret tdk-dev-kratos kratos-runtime destination secret
check_secret tdk-dev-oathkeeper dev-oathkeeper-jwks destination jwks.json
check_secret tdk-dev-oathkeeper dev-oathkeeper-authz destination token
check_secret tdk-dev-tadoku-api dev-oathkeeper-authz destination token

copy_secret() {
  local source_namespace="$1" destination="$2" name="$3"
  # Pipeline stays in memory; neither credentials nor kubectl annotations enter Git.
  kube -n "$source_namespace" get secret "$name" -o json |
    jq --arg namespace "$destination" '{apiVersion:"v1",kind:"Secret",type:.type,data:.data,metadata:{name:.metadata.name,namespace:$namespace,labels:{"app.kubernetes.io/part-of":"tadoku-dev"}}}' |
    kube apply --server-side --field-manager=tadoku-dev-bootstrap -f -
}

copy_secret tdk-dev-data tdk-dev-tadoku-api tadoku.tadoku-dev-db.credentials.postgresql.acid.zalan.do
for provider in kratos keto; do
  copy_secret tdk-dev-data "tdk-dev-$provider" "$provider.tadoku-dev-db.credentials.postgresql.acid.zalan.do"
done

create_runtime_secret() {
  local namespace="$1" name="$2" purpose="$3"
  # A missing Secret is distinct from a failed API request; never rotate on error.
  local existing
  existing=$(kube -n "$namespace" get secret "$name" --ignore-not-found -o name)
  if [ -n "$existing" ]; then return; fi
  node - "$namespace" "$name" "$purpose" <<'JS' | kube create -f -
const crypto = require('node:crypto')
const [namespace, name, purpose] = process.argv.slice(2)
let data
if (purpose === 'jwks') {
  const { privateKey } = crypto.generateKeyPairSync('rsa', { modulusLength: 2048 })
  const key = { ...privateKey.export({ format: 'jwk' }), alg: 'RS256', use: 'sig', kid: crypto.randomUUID() }
  data = { 'jwks.json': JSON.stringify({ keys: [key] }) }
} else {
  data = { [purpose]: crypto.randomBytes(32).toString('hex') }
}
process.stdout.write(JSON.stringify({ apiVersion: 'v1', kind: 'Secret', type: 'Opaque', metadata: { namespace, name, labels: { 'app.kubernetes.io/part-of': 'tadoku-dev' } }, stringData: data }))
JS
}

create_runtime_secret tdk-dev-kratos kratos-runtime secret
create_runtime_secret tdk-dev-oathkeeper dev-oathkeeper-jwks jwks
create_runtime_secret tdk-dev-oathkeeper dev-oathkeeper-authz token
copy_secret tdk-dev-oathkeeper tdk-dev-tadoku-api dev-oathkeeper-authz
echo 'Development credentials are present; existing signing/session keys were preserved.'
