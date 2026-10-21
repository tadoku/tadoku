#!/usr/bin/env bash
#
# Builds the identity importer with the release configuration from the clean,
# checked-out commit and stages the binary, its SHA-256 file and receipt.json
# in OUT_DIR for retention as a CI artifact. The binary is inspected
# statically and never executed.
#
# Usage, from the repository root: tools/ci/stage_identity_importer.sh OUT_DIR
set -euo pipefail

target=//services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import
platform=@@rules_go+//go/toolchain:linux_amd64
flags=(--lockfile_mode=error --config=release)
out="${1:?usage: stage_identity_importer.sh OUT_DIR}"

die() {
  echo "stage_identity_importer: $*" >&2
  exit 1
}

require_clean_source() {
  [[ -z "$(git status --porcelain --untracked-files=all)" ]] ||
    die "the checkout has uncommitted or untracked changes"
}

[[ ! -e "${out}" ]] || die "${out} already exists"

require_clean_source
commit="$(git rev-parse HEAD)"
tree="$(git rev-parse 'HEAD^{tree}')"
if [[ -n "${GITHUB_SHA:-}" && "${GITHUB_SHA}" != "${commit}" ]]; then
  die "HEAD ${commit} is not the run commit ${GITHUB_SHA}"
fi

bazel build "${flags[@]}" "${target}"
require_clean_source
[[ "$(git rev-parse HEAD)" == "${commit}" ]] || die "HEAD moved during the build"

configured="$(bazel cquery "${flags[@]}" --output=starlark \
  --starlark:expr='json.encode({
    "files": [f.path for f in target.files.to_list()],
    "platforms": [str(p) for p in build_options(target)["//command_line_option:platforms"]],
    "pure": build_options(target)["@@rules_go+//go/config:pure"],
  })' \
  "${target}")"
jq -e --arg platform "${platform}" \
  '(.files | length) == 1 and .platforms == [$platform] and .pure == true' \
  <<< "${configured}" > /dev/null ||
  die "expected one pure ${platform} output, got ${configured}"
output="$(jq -r '.files[0]' <<< "${configured}")"
binary="$(bazel info "${flags[@]}" execution_root)/${output}"
[[ -f "${binary}" && ! -L "${binary}" ]] || die "${output} is not a regular file"

header="$(LC_ALL=C readelf -h "${binary}" 2> /dev/null)" || die "${output} is not an ELF file"
elf_class="$(sed -n 's/^ *Class: *//p' <<< "${header}")"
elf_machine="$(sed -n 's/^ *Machine: *//p' <<< "${header}")"
elf_type="$(sed -n 's/^ *Type: *//p' <<< "${header}")"
[[ "${elf_class}" == ELF64 && "${elf_machine}" == "Advanced Micro Devices X86-64" && "${elf_type}" == EXEC* ]] ||
  die "${output} is ${elf_class} ${elf_machine} ${elf_type}, not an ELF64 x86-64 executable"
program_headers="$(LC_ALL=C readelf -lW "${binary}")"
[[ "${program_headers}" != *INTERP* ]] || die "${output} requests a dynamic loader"

bazel_version="$(bazel version --gnu_format | awk '$1 == "bazel" { print $2 }')"
rules_go_version="$(bazel mod graph --lockfile_mode=error --depth=1 --output=json |
  jq -r '.dependencies[] | select(.name == "rules_go") | .version')"
go_version="$(bazel run "${flags[@]}" @rules_go//go -- version "${binary}")"
go_version="${go_version##*: }"
[[ -n "${bazel_version}" && -n "${rules_go_version}" && "${go_version}" =~ ^go[0-9.]+$ ]] ||
  die "could not resolve the toolchain: bazel '${bazel_version}', rules_go '${rules_go_version}', go '${go_version}'"

mkdir -p "${out}"
install -m 0644 "${binary}" "${out}/tadoku-identity-import"
cmp -s "${binary}" "${out}/tadoku-identity-import" || die "the staged copy differs from ${output}"
(cd "${out}" && sha256sum tadoku-identity-import > tadoku-identity-import.sha256)

ci=null
if [[ -n "${GITHUB_RUN_ID:-}" ]]; then
  ci="$(jq -n \
    --arg repository "${GITHUB_REPOSITORY}" \
    --arg workflow "${GITHUB_WORKFLOW}" \
    --arg event "${GITHUB_EVENT_NAME}" \
    --arg ref "${GITHUB_REF}" \
    --arg sha "${GITHUB_SHA}" \
    --arg run_id "${GITHUB_RUN_ID}" \
    --arg run_attempt "${GITHUB_RUN_ATTEMPT}" \
    --arg run_url "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}/attempts/${GITHUB_RUN_ATTEMPT}" \
    '$ARGS.named')"
fi

jq -n \
  --arg target "${target}" \
  --arg sha256 "$(cut -d' ' -f1 "${out}/tadoku-identity-import.sha256")" \
  --argjson bytes "$(stat -c %s "${out}/tadoku-identity-import")" \
  --arg commit "${commit}" \
  --arg tree "${tree}" \
  --arg bazel "${bazel_version}" \
  --arg rules_go "${rules_go_version}" \
  --arg go "${go_version}" \
  --argjson configured "${configured}" \
  --arg lock "$(sha256sum MODULE.bazel.lock | cut -d' ' -f1)" \
  --arg elf_class "${elf_class}" \
  --arg elf_machine "${elf_machine}" \
  --arg elf_type "${elf_type}" \
  --arg command "bazel build ${flags[*]} ${target}" \
  --arg output "${output}" \
  --argjson ci "${ci}" \
  '{
    schema: "tadoku-identity-import-ci-receipt/v1",
    target: $target,
    file: "tadoku-identity-import",
    sha256: $sha256,
    bytes: $bytes,
    source: {commit: $commit, tree: $tree},
    toolchain: {bazel: $bazel, rules_go: $rules_go, go: $go},
    config: {
      name: "release",
      platforms: $configured.platforms,
      pure: $configured.pure,
      lockfile_mode: "error",
      module_lock_sha256: $lock
    },
    elf: {class: $elf_class, machine: $elf_machine, type: $elf_type, interpreter: false},
    build: {command: $command, output: $output},
    ci: $ci,
    executed: false
  }' > "${out}/receipt.json"
