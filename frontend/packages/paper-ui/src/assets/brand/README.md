# Tadoku Paper brand assets

The Cut Meter source follows the approved square-cut geometry exactly:

- view box `0 0 64 64`;
- bars `M7 38h13v18H7z M25 24h14v32H25z M44 8h13v48H44z`;
- negative-space cut `M2 43L62 17`, square-ended, stroke width 8.

`cut-meter.svg` is the canonical monochrome mark and uses `currentColor` for
inline use. The reversed and optional-accent exports are secondary variants;
the violet bar never carries information the monochrome silhouette lacks.

The cleaned wordmark preserves the repository-owned geometry and even-odd fill
behavior from `frontend/packages/ui/components/logo.svg`, traced to repository
commit `f13769b1437bf2d239c9f30cfe2a2d33174512e9`. Sketch metadata, fixed sizing,
and the legacy bright violet were removed. These assets are covered by the
repository's MIT license.

`favicon.svg` is the scalable source. The 16, 32, and 48 pixel SVGs declare
their intended raster dimensions; `apple-touch-icon.svg` declares 180 pixels
and includes the warm paper background required for opaque platform icons.
The favicon geometry uses the approved small-size adjustment (15/16/15-pixel
bars and a 10-unit cut) from the archived v8 study. Applications may rasterize
these sources without altering their view boxes.
