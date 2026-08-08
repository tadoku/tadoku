# ADR 003: Catalogue, fixtures, lifecycle, and routes

Status: accepted

## Context

Navigation, search, routes, status, documentation completeness, fixtures, and tests need one validated source of truth. The catalogue must be router-neutral even though the Vite application owns route resolution.

## Decision

Every catalogue document has a stable ID and this required core metadata: route, name, kind, category, aliases, summary, keywords, lifecycle, owner, review date, source path, package version, guidance, accessibility notes, API/class contract, fixture IDs, dependencies, migration notes, and changelog notes.

Kinds are `foundation`, `component`, `pattern`, `experiment`, and `governance`. Lifecycles are exactly `Experimental`, `Stable`, and `Deprecated`. Canonical routes are:

- `/foundations/:slug`
- `/components/:category/:slug`
- `/patterns/:slug`
- `/experiments/:slug`
- `/contributing` and `/changelog`

Aliases and legacy paths live in one redirect manifest but do not create duplicate canonical documents.

A named fixture contains a stable ID, name, description, tags, supported themes/densities/viewports, and a deterministic render function. It may not read time, randomness, network state, or an application router. Documentation imports fixtures through `paper-ui/catalog`; tests import the same definitions and own assertions separately.

Stable component documents must provide the ordered 16-section instructional contract from the implementation plan. The schema stores each required section separately so validation can identify omissions. Optional component-specific sections follow the required sequence rather than replacing it.

Registry validation fails for duplicate IDs/routes, invalid categories/lifecycles, unknown fixture references, missing Stable metadata/sections/tests/accessibility notes, non-deterministic fixture markers, or Deprecated entries without replacement and removal guidance.

The Vite application derives route resolution, grouped navigation, search records, component indexes, lifecycle filters, and source/status metadata from the validated registry. No parallel hard-coded page list is allowed.

## Consequences

Promoting a document to Stable is an evidence-based operation. Old bookmarks can be cut over later from one redirect manifest while the legacy catalogue remains unchanged during Phases 0–4.

