#!/usr/bin/env bash
# Fails when a handwritten text file is over 20,000 bytes (about 5,000 tokens at 4 bytes per token).
# Checks tracked and untracked, unignored files. Skips binary files and files marked
# linguist-generated=true in .gitattributes.
set -euo pipefail

limit=20000
cd "$(git rev-parse --show-toplevel)"

oversized=$(
  git ls-files -z --cached --others --exclude-standard |
    xargs -0 stat -c '%s %n' 2>/dev/null |
    while read -r size path; do
      ((size > limit)) || continue
      grep -Iq . "$path" || continue
      [[ $(git check-attr linguist-generated -- "$path") == *": true" ]] && continue
      echo "$path"
    done |
    sort -u
)
[[ -z $oversized ]] && exit 0

while read -r path; do
  size=$(stat -c %s "$path")
  echo "::error file=$path::$path is $size bytes (about $((size / 4)) tokens); the limit is $limit bytes."
done <<<"$oversized"
cat <<'EOF'

Split each file along its existing responsibilities so a reader can work on one part without the rest:
- Go: move a cohesive group of declarations, or a group of tests and their fixtures, into another file
  in the same package. Run bazel run //:gazelle afterwards.
- TypeScript and React: move a component, hook or domain group with its types into its own module.
  Keep a package's public exports; inside an app, import from the new module instead of re-exporting.
- Markdown: move a self-contained topic to its own page, link to it and keep the sidebar current.
- YAML and scripts: move an independent job, step script or schema into its own file while keeping
  the entry point that tools and workflows already read.
Do not split by line count, add pass-through modules or move one operation across files.
See "Splitting a long file" in docs/docs/architecture/module-design.md.
Mark a generated file linguist-generated=true in .gitattributes instead of splitting it.
EOF
exit 1
