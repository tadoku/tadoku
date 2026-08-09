import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const navigationCss = readFileSync(
  resolve("src/components/navigation/navigation.css"),
  "utf8",
);

describe("Sidebar visual contract", () => {
  it("uses density tokens for link height and tightens section rhythm in compact contexts", () => {
    expect(navigationCss).toMatch(
      /\.paper-sidebar__link\s*\{[^}]*min-block-size:\s*var\(--paper-control-height\);/su,
    );
    expect(navigationCss).toMatch(
      /\[data-density="compact"\]\s+\.paper-sidebar\s*\{[^}]*gap:\s*1rem;/su,
    );
  });
});
