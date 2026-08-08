import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const tabsCss = readFileSync(resolve("src/components/navigation/Tabs/tabs.css"), "utf8");

describe("Tabs visual contract", () => {
  it("uses a scroll-safe underline treatment for horizontal tabs", () => {
    expect(tabsCss).toMatch(
      /\.paper-tabs__list\s*\{[^}]*inline-size: 100%;[^}]*max-inline-size: 100%;[^}]*overflow-x: auto;[^}]*border-block-end: var\(--paper-border-static-width\) solid var\(--paper-color-rule-default\);/s,
    );
    expect(tabsCss).toMatch(
      /\.paper-tabs__tab\s*\{[^}]*flex: 0 0 auto;[^}]*border: 0;[^}]*border-block-end: var\(--paper-border-action-edge-width\) solid transparent;/s,
    );

    const activeRule = tabsCss.match(/\.paper-tabs__tab\[data-active\]\s*\{([^}]*)\}/s)?.[1];
    expect(activeRule).toBeDefined();
    expect(activeRule).toContain(
      "border-block-end-color: var(--paper-color-action-default);",
    );
    expect(activeRule).toContain("background: transparent;");
    expect(activeRule).not.toMatch(/(?:^|[;\s])border(?:-block-start|-inline(?:-start|-end)?)?(?:-color)?:/);
  });

  it("styles the complete interaction state set with Paper tokens", () => {
    expect(tabsCss).toMatch(/\.paper-tabs__tab\s*\{[^}]*var\(--paper-control-height\)/s);
    expect(tabsCss).toMatch(/\.paper-tabs__tab:hover:not\(:disabled\):not\(\[data-active\]\)/s);
    expect(tabsCss).toMatch(/\.paper-tabs__tab\[data-active\]\s*\{/s);
    expect(tabsCss).toMatch(/\.paper-tabs__tab:focus-visible\s*\{/s);
    expect(tabsCss).toMatch(/\.paper-tabs__tab\[data-disabled\]\s*\{/s);
    expect(tabsCss).toContain('[data-density="compact"] .paper-tabs__tab');
    expect(tabsCss).toContain("var(--paper-color-surface-paper)");
    expect(tabsCss).toContain("var(--paper-color-text-ink)");
    expect(tabsCss).toContain("var(--paper-color-action-default)");
    expect(tabsCss).toContain("var(--paper-color-focus-ring)");
  });

  it("supports orientation and forced-colors without relying on literal palette colors", () => {
    expect(tabsCss).toContain('.paper-tabs[data-orientation="vertical"]');
    expect(tabsCss).toContain("@media (forced-colors: active)");
    expect(tabsCss).toContain("CanvasText");
    expect(tabsCss).not.toMatch(/#[0-9a-f]{3,8}\b/iu);
    expect(tabsCss).not.toMatch(/rgb\(/u);
  });
});
