import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const formsCss = readFileSync(resolve("src/components/forms/forms.css"), "utf8");

describe("segmented RadioSelect visual contract", () => {
  it("uses a connected, responsive, density-aware Paper treatment", () => {
    expect(formsCss).toMatch(/\.paper-radio-select--segmented\s+\.paper-choice-list\s*\{/s);
    expect(formsCss).toContain("flex-wrap: wrap");
    expect(formsCss).toMatch(/\.paper-radio-select__segment\s*\{[^}]*min-block-size:\s*var\(--paper-control-height\)/s);
    expect(formsCss).toMatch(/\.paper-radio-select__segment:has\(input:checked\)\s*\{/s);
    expect(formsCss).toContain("var(--paper-color-surface-paper)");
    expect(formsCss).toContain("var(--paper-color-text-ink)");
    expect(formsCss).toContain("var(--paper-color-action-soft)");
    expect(formsCss).toContain("var(--paper-color-rule-action-edge)");
    expect(formsCss).toContain('[data-density="compact"] .paper-radio-select__segment');
  });

  it("covers focus, disabled, invalid, logical direction, and forced colors", () => {
    expect(formsCss).toMatch(/\.paper-radio-select__segment:has\(input:focus-visible\)\s*\{/s);
    expect(formsCss).toMatch(/\.paper-radio-select__segment:has\(input:disabled\)\s*\{/s);
    expect(formsCss).toMatch(/\.paper-field--invalid[^\n{]*\.paper-radio-select__segment/s);
    expect(formsCss).toContain("border-inline-start");
    expect(formsCss).toContain("border-block-end");
    expect(formsCss).toContain("@media (forced-colors: active)");
    expect(formsCss).toContain("CanvasText");
    expect(formsCss).not.toMatch(/#[0-9a-f]{3,8}\b/iu);
    expect(formsCss).not.toMatch(/rgb\(/u);
  });
});
