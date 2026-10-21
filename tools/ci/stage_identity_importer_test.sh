#!/usr/bin/env bash
#
# Tests stage_identity_importer.sh against a disposable Git repository, a fake
# bazel and inert fixture files. It never builds or executes the importer:
# a canary script stands in for a wrong output and must never run.
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/stage_identity_importer.sh"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
failures=0

fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

mkdir -p "${work}/bin"
cat > "${work}/bin/bazel" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "$*" >> "${FAKE_BAZEL_LOG}"
case "$1" in
  version) printf 'Bazelisk v1.29.0\nbazel 9.2.0\n' ;;
  mod) echo '{"key":"<root>","dependencies":[{"key":"rules_go@0.63.0","name":"rules_go","version":"0.63.0"}]}' ;;
  build)
    if [[ -n "${FAKE_BUILD_TOUCH:-}" ]]; then touch "${FAKE_BUILD_TOUCH}"; fi
    exit "${FAKE_BUILD_EXIT:-0}"
    ;;
  cquery) printf '%s\n' "${FAKE_CQUERY}" ;;
  info) echo "${FAKE_EXECROOT}" ;;
  run) echo "${*: -1}: go1.26.9" ;;
  *) echo "fake bazel: unexpected $*" >&2; exit 2 ;;
esac
EOF
chmod +x "${work}/bin/bazel"

repo="${work}/repo"
git init -q "${repo}"
printf 'module fixture\n' > "${repo}/MODULE.bazel.lock"
git -C "${repo}" add MODULE.bazel.lock
git -C "${repo}" -c user.name=fixture -c user.email=fixture@example.invalid commit -qm fixture
head="$(git -C "${repo}" rev-parse HEAD)"
tree="$(git -C "${repo}" rev-parse 'HEAD^{tree}')"
lock_sha="$(sha256sum "${repo}/MODULE.bazel.lock" | cut -d' ' -f1)"

execroot="${work}/execroot"
importer_rel="bazel-out/k8-fastbuild/bin/services/tadoku-api/cmd/tadoku-identity-import/tadoku-identity-import_/tadoku-identity-import"
canary_rel="bazel-out/k8-fastbuild/bin/canary"
mkdir -p "$(dirname "${execroot}/${importer_rel}")"
# Inert fixture: a 64-byte ELF64 x86-64 EXEC header with no program headers.
{
  printf '\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00\x01\x00\x00\x00'
  printf '\x00%.0s' {1..28}
  printf '\x40\x00\x38\x00\x00\x00\x40\x00\x00\x00\x00\x00'
} > "${execroot}/${importer_rel}"
chmod +x "${execroot}/${importer_rel}"
fixture_sha="$(sha256sum "${execroot}/${importer_rel}" | cut -d' ' -f1)"
fixture_bytes="$(stat -c %s "${execroot}/${importer_rel}")"
canary_marker="${work}/canary-ran"
printf '#!/bin/sh\ntouch %s\n' "${canary_marker}" > "${execroot}/${canary_rel}"
chmod +x "${execroot}/${canary_rel}"

good_cquery="{\"files\":[\"${importer_rel}\"],\"platforms\":[\"@@rules_go+//go/toolchain:linux_amd64\"],\"pure\":true}"

# run_stage NAME [VAR=VALUE...] runs the script in a clean environment and
# leaves its output in ${work}/NAME.
run_stage() {
  local name="$1"
  shift
  (
    cd "${repo}"
    env -i \
      PATH="${work}/bin:/usr/bin:/bin" \
      FAKE_BAZEL_LOG="${work}/${name}.bazel.log" \
      FAKE_EXECROOT="${execroot}" \
      FAKE_CQUERY="${good_cquery}" \
      GITHUB_SHA="${head}" \
      GITHUB_REPOSITORY=tadoku/tadoku \
      GITHUB_WORKFLOW=build-bazel \
      GITHUB_RUN_ID=7 \
      GITHUB_RUN_ATTEMPT=1 \
      GITHUB_EVENT_NAME=push \
      GITHUB_REF=refs/heads/main \
      GITHUB_SERVER_URL=https://github.com \
      "$@" \
      "${script}" "${work}/${name}"
  ) > "${work}/${name}.log" 2>&1
}

expect_refusal() {
  local name="$1"
  shift
  if run_stage "${name}" "$@"; then
    fail "${name}: expected the script to fail"
  fi
  if [[ -e "${work}/${name}/tadoku-identity-import" ]]; then
    fail "${name}: a binary was staged"
  fi
}

# Success: exactly the binary, its checksum and the receipt, bound to source.
if run_stage ok; then
  out="${work}/ok"
  listing="$(cd "${out}" 2>/dev/null && ls -A | tr '\n' ' ' || true)"
  [[ "${listing}" == "receipt.json tadoku-identity-import tadoku-identity-import.sha256 " ]] ||
    fail "ok: staged files are '${listing}'"
  (cd "${out}" 2>/dev/null && sha256sum --quiet -c tadoku-identity-import.sha256) || fail "ok: checksum file does not verify"
  [[ ! -x "${out}/tadoku-identity-import" ]] || fail "ok: staged binary is executable"
  cmp -s "${out}/tadoku-identity-import" "${execroot}/${importer_rel}" || fail "ok: staged bytes differ from the output"
  expected="$(jq -n \
    --arg sha "${fixture_sha}" \
    --argjson bytes "${fixture_bytes}" \
    --arg commit "${head}" \
    --arg tree "${tree}" \
    --arg lock "${lock_sha}" \
    --arg output "${importer_rel}" \
    '{
      schema: "tadoku-identity-import-ci-receipt/v1",
      target: "//services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import",
      file: "tadoku-identity-import",
      sha256: $sha,
      bytes: $bytes,
      source: {commit: $commit, tree: $tree},
      toolchain: {bazel: "9.2.0", rules_go: "0.63.0", go: "go1.26.9"},
      config: {
        name: "release",
        platforms: ["@@rules_go+//go/toolchain:linux_amd64"],
        pure: true,
        lockfile_mode: "error",
        module_lock_sha256: $lock
      },
      elf: {class: "ELF64", machine: "Advanced Micro Devices X86-64", type: "EXEC (Executable file)", interpreter: false},
      build: {
        command: "bazel build --lockfile_mode=error --config=release //services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import",
        output: $output
      },
      ci: {
        repository: "tadoku/tadoku",
        workflow: "build-bazel",
        event: "push",
        ref: "refs/heads/main",
        sha: $commit,
        run_id: "7",
        run_attempt: "1",
        run_url: "https://github.com/tadoku/tadoku/actions/runs/7/attempts/1"
      },
      executed: false
    }')"
  actual="$(jq -S . "${out}/receipt.json" 2>/dev/null || echo invalid)"
  [[ "${actual}" == "$(jq -S . <<< "${expected}")" ]] ||
    fail "ok: receipt differs from expected:"$'\n'"$(diff <(jq -S . <<< "${expected}") <(echo "${actual}") || true)"
  grep -qsx 'build --lockfile_mode=error --config=release //services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import' \
    "${work}/ok.bazel.log" || fail "ok: release build with frozen lockfile was not requested"
else
  fail "ok: script failed: $(cat "${work}/ok.log")"
fi

touch "${repo}/untracked"
expect_refusal dirty-source
rm "${repo}/untracked"

expect_refusal source-changed-during-build FAKE_BUILD_TOUCH="${repo}/generated"
rm -f "${repo}/generated"

expect_refusal wrong-commit GITHUB_SHA=0000000000000000000000000000000000000000
expect_refusal build-failed FAKE_BUILD_EXIT=1
expect_refusal no-output FAKE_CQUERY='{"files":[],"platforms":["@@rules_go+//go/toolchain:linux_amd64"],"pure":true}'
expect_refusal two-outputs \
  FAKE_CQUERY="{\"files\":[\"${importer_rel}\",\"${importer_rel}\"],\"platforms\":[\"@@rules_go+//go/toolchain:linux_amd64\"],\"pure\":true}"
expect_refusal not-pure FAKE_CQUERY="{\"files\":[\"${importer_rel}\"],\"platforms\":[\"@@rules_go+//go/toolchain:linux_amd64\"],\"pure\":false}"
expect_refusal wrong-platform FAKE_CQUERY="{\"files\":[\"${importer_rel}\"],\"platforms\":[\"@@rules_go+//go/toolchain:darwin_arm64\"],\"pure\":true}"
expect_refusal not-elf FAKE_CQUERY="{\"files\":[\"${canary_rel}\"],\"platforms\":[\"@@rules_go+//go/toolchain:linux_amd64\"],\"pure\":true}"

mkdir "${work}/existing-output"
expect_refusal existing-output

[[ ! -e "${canary_marker}" ]] || fail "the canary output was executed"

if ((failures)); then
  echo "${failures} failure(s)" >&2
  exit 1
fi
echo "stage_identity_importer: all cases passed"
