import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { chartPalette, chartSeries } from "./chart-palette";

const foundationsRoot = dirname(fileURLToPath(import.meta.url));
const assetsRoot = resolve(foundationsRoot, "../assets");
const readFoundation = (name: string) =>
  readFileSync(resolve(foundationsRoot, name), "utf8");

describe("Paper visual foundations", () => {
  it("defines both themes and both density modes through semantic roles", () => {
    const tokens = readFoundation("tokens.css");

    expect(tokens).toContain('[data-theme="light"]');
    expect(tokens).toContain('[data-theme="dark"]');
    expect(tokens).toContain('[data-density="comfortable"]');
    expect(tokens).toContain('[data-density="compact"]');
    expect(tokens).toContain("--paper-control-height: 2.75rem");
    expect(tokens).toContain("--paper-control-height: 2.25rem");
    expect(tokens).toContain("--paper-color-focus-ring");
    expect(tokens).toContain("--paper-elevation-floating");
  });

  it("ships local declarations for every approved real font weight", () => {
    const fonts = readFoundation("fonts.css");

    expect(fonts).not.toMatch(/https?:\/\//);
    expect(fonts.match(/@font-face/g)).toHaveLength(4);
    expect(fonts).toContain('font-weight: 400');
    expect(fonts).toContain('font-weight: 600');
    expect(fonts.match(/font-weight: 700/g)).toHaveLength(2);

    for (const filename of [
      "merriweather-700.woff2",
      "open-sans-400.woff2",
      "open-sans-600.woff2",
      "open-sans-700.woff2",
    ]) {
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
