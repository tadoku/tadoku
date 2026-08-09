# Tadoku Paper brand assets

The Cut Meter source follows the approved square-cut geometry exactly:

- view box `0 0 64 64`;
- bars `M7 38h13v18H7z M25 24h14v32H25z M44 8h13v48H44z`;
- negative-space cut `M2 43L62 17`, square-ended, stroke width 8.

`cut-meter.svg` is the canonical monochrome mark and uses `currentColor` for
inline use. The reversed and optional-accent exports are secondary variants;
the violet bar never carries information the monochrome silhouette lacks.

The accent and reversed wordmarks preserve the original exported artwork from
`frontend/packages/ui/components/logo.svg` and `logo-light.svg`, respectively,
including the 158 by 29 view box, fixed sizing, colors, paths, and metadata.
Their only byte-level difference is the package source's trailing newline.
These assets are covered by the repository's MIT license.

`favicon.svg` is the scalable source. The 16, 32, and 48 pixel SVGs declare
their intended raster dimensions; `apple-touch-icon.svg` declares 180 pixels
and includes the warm paper background required for opaque platform icons.
The favicon geometry uses the approved small-size adjustment (15/16/15-pixel
bars and a 10-unit cut) from the archived v8 study. Applications may rasterize
these sources without altering their view boxes.
