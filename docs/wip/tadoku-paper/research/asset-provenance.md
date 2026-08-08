# Asset provenance and export record

## Fonts

The legacy stylesheet loads Merriweather 700 and Open Sans 400/400 italic/500/600 at runtime from Google Fonts. It references Open Sans 700 without loading a real 700 face. Paper replaces the runtime import with repository-owned WOFF2 assets.

Merriweather and Open Sans are distributed by Google Fonts under open-source licenses. Production assets must be copied from a versioned upstream source together with the applicable `OFL.txt`; the commit or release identifier and SHA-256 for each shipped WOFF2 file must be recorded beside the assets. Required initial faces are Merriweather 700 and Open Sans 400, 600, and 700 in normal style. Italic is added only if a component/content inventory proves current use.

Authoritative references:

- Google Fonts states that its fonts are released under open-source licenses: <https://developers.google.com/fonts>
- Open Sans upstream and license: <https://github.com/googlefonts/opensans>
- Google Fonts repository: <https://github.com/google/fonts>

## Existing Tadoku wordmark

`frontend/packages/ui/components/logo.svg` and `logo-light.svg` are repository assets under the repository's MIT license. Git history traces the current source to commit `f13769b1437bf2d239c9f30cfe2a2d33174512e9` (2022-12-04), with subsequent corrections. Paper may preserve the path geometry but must remove Sketch metadata and fixed palette values, use `currentColor`/semantic accent values, preserve even-odd fill behavior, and render through ordinary SVG rather than `next/image`.

## Cut Meter

The canonical source is the approved square-diagonal construction archived in `design-system-cut-meter-variations-v8.html`:

- view box `0 0 64 64`;
- bars `M7 38h13v18H7z M25 24h14v32H25z M44 8h13v48H44z`;
- negative-space cut `M2 43L62 17`, square-ended, stroke width 8.

The production family requires monochrome, reversed, and optional-accent SVGs; a wordmark lockup; and favicon sources tested at 16, 32, 48, and 180 pixels. Color must remain optional: the monochrome silhouette is canonical.

## Required production record

Before the Phase 1 gate, add the actual asset paths, upstream commit/release identifiers, SHA-256 hashes, generated favicon sizes, and a reproduction command. No remote font request or Next-specific asset loader is permitted in `paper-ui` or `paper-styleguide`.

