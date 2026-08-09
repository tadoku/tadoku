import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const drawerCss = readFileSync(
  resolve("src/components/overlays/Drawer/drawer.css"),
  "utf8",
);

describe("Drawer visual contract", () => {
  it("uses Paper tokens for its themed, density-aware fixed sheet", () => {
    expect(drawerCss).toMatch(
      /\.paper-drawer\s*\{[^}]*background:\s*var\(--paper-color-surface-overlay\);[^}]*color:\s*var\(--paper-color-text-ink\);/s,
    );
    expect(drawerCss).toContain("var(--paper-control-height)");
    expect(drawerCss).toContain("var(--paper-inline-gap)");
    expect(drawerCss).toContain("var(--paper-type-body-size)");
  });

  it("keeps the body scrollable and animates both logical placements", () => {
    expect(drawerCss).toMatch(
      /\.paper-drawer\s*\{[^}]*display:\s*flex;[^}]*flex-direction:\s*column;[^}]*overflow:\s*hidden;/s,
    );
    expect(drawerCss).toMatch(
      /\.paper-drawer__body\s*\{[^}]*min-block-size:\s*0;[^}]*overflow-y:\s*auto;/s,
    );
    expect(drawerCss).toMatch(
      /\.paper-drawer\[data-placement="start"\]\[data-starting-style\]/,
    );
    expect(drawerCss).toMatch(
      /\.paper-drawer\[data-placement="end"\]\[data-starting-style\]/,
    );
  });

  it("honors reduced motion and forced colors", () => {
    expect(drawerCss).toMatch(
      /@media \(prefers-reduced-motion:\s*reduce\)\s*\{[^}]*\.paper-drawer,/s,
    );
    expect(drawerCss).toMatch(
      /@media \(forced-colors:\s*active\)\s*\{[^}]*\.paper-drawer__backdrop\s*\{[^}]*forced-color-adjust:\s*none;/s,
    );
  });
});
