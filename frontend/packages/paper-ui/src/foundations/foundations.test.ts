import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { themeCss } from "../tokens/css";
import { chartPalette, chartSeries } from "./chart-palette";
import { documentClassName } from "./document";

const foundationsRoot = dirname(fileURLToPath(import.meta.url));
const assetsRoot = resolve(foundationsRoot, "../assets");

describe("Paper visual foundations", () => {
  it("re-declares themed and density tokens, and what reads them, inside each scope", () => {
    const css = themeCss();
    const scope = (selector: string) => css.slice(css.indexOf(`${selector} {`), css.indexOf("}", css.indexOf(`${selector} {`)));

    expect(scope('[data-theme="dark"]')).toMatch(/--color-ink: #/);
    expect(scope('[data-theme="dark"]')).toContain("--color-focus-halo: color-mix(in srgb, var(--color-focus) 12%, transparent)");
    expect(scope('[data-theme="light"]')).toContain("--color-ink:");
    expect(scope('[data-density="compact"]')).toContain("--spacing-control: 2.25rem");
    expect(scope('[data-density="compact"]')).toContain("--spacing-select-end:");
    expect(scope('[data-density="compact"]')).not.toContain("--color-");
    expect(scope('[data-theme="dark"]')).not.toContain("--spacing-control");
    expect(css).toContain("@media (width < 48rem)");
  });

  it("keeps the global keyboard focus ring and removes it only from script-moved focus", () => {
    const classes = documentClassName.split(" ");

    expect(classes).toEqual(expect.arrayContaining([
      "[:where(&_:focus-visible)]:outline-ring",
      "[:where(&_:focus-visible)]:outline-focus",
    ]));
    for (const candidate of classes.filter((name) => name.endsWith(":outline-none"))) {
      expect(candidate).toMatch(/\[tabindex='-1'\]:focus-visible\)\]:outline-none$/);
    }
  });

  it("ships local declarations for every approved real font weight", () => {
    const css = themeCss();

    expect(css).not.toMatch(/https?:\/\//);
    expect(css.match(/@font-face/g)).toHaveLength(4);
    expect(css).toContain("font-weight: 400");
    expect(css).toContain("font-weight: 600");
    expect(css.match(/font-weight: 700/g)).toHaveLength(2);

    for (const filename of [
      "merriweather-700.woff2",
      "open-sans-400.woff2",
      "open-sans-600.woff2",
      "open-sans-700.woff2",
    ]) {
      expect(css).toContain(`./assets/fonts/${filename}`);
      const font = readFileSync(resolve(assetsRoot, "fonts", filename));
      expect(font.subarray(0, 4).toString("ascii")).toBe("wOF2");
    }
  });

  it("keeps the chart sequence aligned with its non-color cues", () => {
    expect(chartPalette).toHaveLength(8);
    expect(chartSeries).toHaveLength(chartPalette.length);
    expect(new Set(chartSeries.map(({ pattern }) => pattern)).size).toBeGreaterThan(1);
    expect(new Set(chartSeries.map(({ pointStyle }) => pointStyle)).size).toBeGreaterThan(1);
  });

  it("preserves canonical Cut Meter geometry", () => {
    const mark = readFileSync(resolve(assetsRoot, "brand/cut-meter.svg"), "utf8");

    expect(mark).toContain('viewBox="0 0 64 64"');
    expect(mark).toContain("M7 38h13v18H7zM25 24h14v32H25zM44 8h13v48H44z");
    expect(mark).toContain("M2 43L62 17");
    expect(mark).toContain('stroke-width="8"');
  });

  it("matches the checked-in production asset hashes", () => {
    const manifest = readFileSync(resolve(assetsRoot, "SHA256SUMS"), "utf8");

    for (const line of manifest.trim().split("\n")) {
      const [expected, path] = line.split(/\s+/, 2);
      const actual = createHash("sha256")
        .update(readFileSync(resolve(assetsRoot, path)))
        .digest("hex");

      expect(actual, path).toBe(expected);
    }
  });
});
