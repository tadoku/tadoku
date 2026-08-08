# Migration marker and repository-check proposal

Status: Phase 0 proposal; no guard implementation in this inventory lane.

## Marker design

Add a committed repository data file owned by the Paper integrator, for example `frontend/paper-migrations.json`:

```json
{
  "schemaVersion": 1,
  "applications": {
    "paper-styleguide": "paper",
    "styleguide": "legacy",
    "admin": "legacy",
    "auth": "legacy",
    "webv2": "legacy"
  }
}
```

Allowed values are exactly `legacy`, `paper`, and later `removed`. The marker changes in the same deployable application-cutover PR only after its zero-mixing and build gates pass. It is not inferred from dependencies, which makes intended state reviewable and prevents partial migrations from appearing legitimate.

The final sequence is:

1. `paper-styleguide=paper`, legacy styleguide/admin/auth/webv2 `legacy`.
2. Admin, then auth, then webv2 change individually to `paper` after their cutover gates.
3. At final hostname/source cleanup, legacy styleguide becomes `removed`; package-wide legacy allowances disappear.

## Check behavior

A single deterministic script (suggested path `frontend/scripts/check-paper-boundaries.mjs`) should read the marker and fail with file/line evidence. It should use repository-relative paths and no network.

### Rules for every Paper package/application

- Reject module imports/requires matching `ui`, `ui/*`, or workspace dependency `"ui": "workspace:*"`.
- Reject `ui/styles/globals.css`, legacy Tailwind inheritance/mutation, and `transpilePackages` entries for `ui`.
- Reject `@headlessui/react` source imports and dependency/devDependency/peerDependency entries.
- Reject `next`, `next/*`, and Next configuration inside `paper-ui` and `paper-styleguide`; product apps may retain Next.
- Reject `paper-ui/src/*` and other private paths; allow only package export-map paths.
- Require exactly one application entry import of `paper-ui/styles.css`; reject zero and duplicates. Exempt `paper-ui` package tests/build internals.
- Parse static and composed JSX classes and reject the versioned legacy-only class set. Because tokens such as `title`, `error`, `default`, and `link` are generic, compare only actual class tokens and maintain explicit Paper recipe allowlisting.

### Rules for every legacy application

- Allow `ui` imports/dependency/configuration and reject all `paper-ui` imports/dependencies/styles.
- The legacy styleguide remains legacy until final domain cutover.
- Do not globally reject Headless UI while `ui` remains, but reject new direct app source imports; the existing app manifests may stay allowlisted until each cutover.

### Package-global rules

- `paper-ui` may depend on Base UI but must not expose Base UI types from public declarations where avoidable.
- No app may load both stylesheet roots.
- Headless UI imports are allowlisted only under `frontend/packages/ui/**` during coexistence.
- After every marker is `paper`/`removed`, repository search for `ui`, legacy styles/config, and Headless UI must be zero before deletion is allowed.

## Files and syntax to inspect

| Concern | Files |
| --- | --- |
| imports/requires/stylesheet entries | `.ts`, `.tsx`, `.js`, `.jsx`, `.mjs`, `.cjs` |
| workspace and forbidden dependencies | all relevant `package.json` sections |
| Next transpilation | `next.config.*` |
| Tailwind inheritance/content paths | `tailwind.config.*` |
| duplicate stylesheet entry | source import graph rooted at app entry (`_app.tsx` today; Vite entry for Paper) |
| raw legacy classes | JSX/TSX class attributes and recognized composition helpers/template literals |
| lockfile cleanup | importer block for the migrated package, not global package snapshots during coexistence |

Use a parser where available. Regex remains a defense-in-depth repository search, not the sole class/import proof.

## CI and local integration

- Add `pnpm paper:check-boundaries` at the frontend root and run it in every frontend workflow.
- Path-filter the check on all app/package manifests, `paper-ui/**`, both styleguides, app source/config, lockfile, marker, and the check itself.
- Run the check before expensive production/image builds.
- A marker transition must require the target app lint, typecheck, production build, image build, app smoke evidence, and rollback-image record.
- Add small fixture directories for positive/negative guard tests: legacy app importing Paper, Paper app importing legacy UI, private deep import, duplicate CSS, forbidden manifest dependency, and a false-positive generic `error` class that is allowed by Paper.

## Zero-mixing searches for human review

Target-app review (replace `$APP` with an explicit validated directory in automation):

```sh
rg -n "from ['\"]ui|import ['\"]ui/|ui/styles|ui/tailwind|@headlessui/react|paper-ui/src/" frontend/apps/<app>
rg -n '"ui"|"@headlessui/react"' frontend/apps/<app>/package.json
rg -n "paper-ui/styles\.css" frontend/apps/<app>
```

Final cleanup proof:

```sh
rg -n "from ['\"]ui|ui/styles|ui/components|\"ui\": \"workspace|@headlessui/react" frontend
```

## Why a marker is necessary

Repository-level coexistence is permitted while application-level mixing is forbidden. A global import ban cannot express that transition, and inference from current imports would bless accidental partial work. The explicit marker gives CI one stable policy input and gives reviewers a visible cutover event.

