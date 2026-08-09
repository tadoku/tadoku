# Publish Tadoku OpenAPI documentation on GitHub Pages

Status: approved for implementation  
Date: 9 August 2026  
Target: <https://tadoku.github.io/tadoku/api/>

## Overview

Extend the Docusaurus documentation site already deployed at
`https://tadoku.github.io/tadoku/` with generated, navigable API reference
pages. The OpenAPI YAML files remain the source of truth, generated pages are
checked into Git for review, and CI rejects stale output or accidental
publication of internal service contracts.

The initial release is documentation-only. It shows request and response
schemas and code samples, but does not expose browser-side request execution.
The APIs use cookie authentication and the repository does not define a
cross-origin security policy for requests from GitHub Pages.

### Public allowlist

| Service | Source | Operations |
| --- | --- | ---: |
| Authorization API | `services/authz-api/http/rest/openapi/api.yaml` | 4 |
| Content API | `services/content-api/http/rest/openapi/api.yaml` | 21 |
| Immersion API | `services/immersion-api/http/rest/openapi/api.yaml` | 34 |
| Profile API | `services/profile-api/http/rest/openapi/api.yaml` | 2 |

The following contracts are explicitly excluded:

- `services/authz-api/http/rest/openapi/internal-api.yaml`
- `services/profile-api/http/rest/openapi/internal-api.yaml`

Use exact paths throughout the configuration and workflow. Do not glob the
OpenAPI directories.

## Benefits and trade-offs

Benefits:

- Searchable, deep-linked API documentation appears alongside the existing
  architecture and service documentation.
- The specifications remain the source of truth and CI detects stale pages.
- The existing GitHub Pages project, permissions, and URL can be reused.
- Public and internal contracts have an explicit, testable boundary.

Risks and mitigations:

| Risk | Impact | Mitigation |
| --- | --- | --- |
| An internal spec is published | High | Four exact config entries, exact workflow paths, and negative scans of the built output |
| Cookie-authenticated browser requests introduce CORS or CSRF problems | High | Hide the request-sending control and disable credential persistence |
| Docusaurus and renderer versions are incompatible | Medium | Upgrade all Docusaurus packages together and pin a compatible renderer pair |
| Generated pages drift from the YAML | Medium | Regenerate in pull-request CI and fail on a generated-directory diff |
| Spec-only changes do not deploy | Medium | Add all four public spec paths to the Pages workflow triggers |

## Phase 0: establish the baseline and safety contract

- [ ] Record the existing Pages URL, `/tadoku/` base path, and representative
      documentation routes.
- [ ] Confirm that only the four public `api.yaml` files are publishable.
- [ ] Record the expected total of 61 public operations.
- [ ] Add negative assertions for `/internal/v1/`, `authz-api:8080`, and
      `profile-api:8080`.
- [ ] Build the current site before adding generated reference pages.

Gate: maintainers approve the public allowlist and the existing site has a
known-good baseline.

## Phase 1: make the docs toolchain reproducible

- [ ] Add a pinned pnpm version to `docs/package.json`.
- [ ] Generate `docs/pnpm-lock.yaml` and remove `docs/package-lock.json`.
- [ ] Upgrade all Docusaurus packages from 3.9.2 to a mutually compatible
      release.
- [ ] Add pinned `docusaurus-plugin-openapi-docs` and
      `docusaurus-theme-openapi-docs` packages and the required Sass plugin.
- [ ] Change the Pages workflow to use pnpm with a frozen lockfile.
- [ ] Typecheck and build the unchanged site before API integration.

Gate: the existing site typechecks and builds from a frozen pnpm install while
preserving the `/tadoku/` base path and current routes.

## Phase 2: generate the public API reference

This phase can run in parallel with the content and workflow preparation after
Phase 1 passes.

- [ ] Register four named OpenAPI renderer configurations using exact spec
      paths.
- [ ] Generate one stable directory per service under `docs/docs/api/`.
- [ ] Group operations by OpenAPI tag and create service overview pages.
- [ ] Hide browser-side request execution and keep credential persistence off.
- [ ] Add `api:generate` and `api:check` scripts.
- [ ] Commit all generated MDX and sidebar output; never hand-edit it.
- [ ] Ensure repeated operation IDs such as `ping` have unique service-scoped
      routes.

Gate: clean regeneration creates no diff, all 61 public operations render, and
no internal path or cluster-local server hostname appears in the build.

## Phase 3: integrate the reference into the documentation

- [ ] Add an `API Reference` navigation entry.
- [ ] Compose the four generated sidebar slices in this order: Immersion,
      Content, Profile, Authorization.
- [ ] Add an API landing page containing the service map, base URLs, versions,
      authentication guidance, and source links.
- [ ] Explain that request execution is intentionally disabled in the first
      release.
- [ ] Update existing service pages to link to the rendered reference while
      retaining links to the source specifications.
- [ ] Check desktop and mobile navigation, light and dark themes, endpoint
      anchors, schema expansion, and code samples.

Gate: every service reference is reachable in two clicks and readers can
understand its base URL and authentication model.

## Phase 4: verify pull requests and deploy from main

- [ ] Extend workflow path filters with the four exact public spec paths.
- [ ] Run frozen install, API regeneration, generated drift detection,
      typecheck, internal-spec checks, and production build on relevant pull
      requests.
- [ ] Configure GitHub Pages before artifact upload and retain the Pages
      permissions, environment, and serialized deployment group.
- [ ] Deploy only for pushes to `main` and manual dispatches.
- [ ] Upload only `docs/build`.

Gate: a spec-only pull request runs documentation verification, pull requests
cannot deploy, and a main-branch run reports the expected Pages URL.

## Phase 5: release verification

- [ ] Verify `https://tadoku.github.io/tadoku/api/` after deployment.
- [ ] Smoke-test one anonymous and one authenticated operation page per
      service.
- [ ] Recheck representative existing documentation routes.
- [ ] Confirm the deployed site does not contain internal spec names, internal
      paths, or cluster-local hostnames.
- [ ] Confirm a subsequent docs-only update can replace the deployment.
- [ ] Record rollback: revert the documentation and workflow commits and
      redeploy the prior static artifact; no service or database rollback is
      required.

Gate: all success criteria pass against the production Pages URL and both an
API owner and docs/frontend owner approve the result.

## Measurable success criteria

- Four service overview pages and all 61 public operations are reachable under
  `/tadoku/api/`.
- The built and deployed output contains no internal spec names,
  `/internal/v1/` routes, or cluster-local API hostnames.
- Changing any allowlisted `api.yaml` runs the documentation checks and stale
  generated MDX fails CI.
- Existing documentation URLs continue to work on desktop and mobile.
- A clean checkout passes frozen pnpm install, API generation, typecheck, and
  production build.
- No browser-side `Send request` control is exposed.

## Atomic delivery

1. Toolchain: pnpm migration, Docusaurus upgrade, renderer dependencies, and
   unchanged-site proof.
2. Reference: allowlisted configs, generated MDX, API landing page, navigation,
   and service links.
3. Delivery: pull-request verification, spec triggers, Pages configuration,
   internal-spec guard, and release checks.

## Follow-up improvements

- Add OpenAPI linting after agreeing on a repository ruleset.
- Improve descriptions, examples, error schemas, and deprecation notes in the
  source specifications.
- Publish version snapshots when API versions diverge from the current
  contracts.
- Add documentation search and API-reference analytics.
- Design a separately reviewed authenticated API explorer.
- Optionally publish internal contracts to an authenticated documentation site.

