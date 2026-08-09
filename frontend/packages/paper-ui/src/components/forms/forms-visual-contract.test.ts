import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const formsCss = readFileSync(resolve("src/components/forms/forms.css"), "utf8");

const actionHugTargets = [
  ".paper-choice input:focus-visible",
  ".paper-radio-select__segment:has(input:focus-visible)",
  ".paper-radio-card:has(input:focus-visible)",
  ".paper-combobox__chips:has(.paper-combobox__chip-input:focus-visible)",
] as const;

function leafRulesContaining(selector: string) {
  return [...formsCss.matchAll(/(?<selectors>[^{}]+)\{(?<declarations>[^{}]*)\}/gs)]
    .filter((match) => match.groups?.selectors.includes(selector))
    .map((match) => match.groups?.declarations ?? "");
}

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

describe("choice and multi-value control action hug", () => {
  it.each(actionHugTargets)(
    "gives %s the approved zero-offset, low-opacity keyboard-focus hug",
    (selector) => {
      const focusedRule = leafRulesContaining(selector).find((declarations) =>
        declarations.includes("color-mix"),
      );

      expect(focusedRule).toBeDefined();
      expect(focusedRule ?? "").toContain(
        "--paper-control-focus-color: var(--paper-color-focus-ring)",
      );
      expect(focusedRule ?? "").toMatch(
        /outline:\s*3px solid color-mix\(in srgb, var\(--paper-control-focus-color\) 10%, transparent\)/s,
      );
      expect(focusedRule ?? "").toMatch(/outline-offset:\s*0/s);
    },
  );

  it.each(actionHugTargets)(
    "keeps %s focus styling free of border-width and layout changes",
    (selector) => {
      const focusedRule = leafRulesContaining(selector).find((declarations) =>
        declarations.includes("color-mix"),
      );

      expect(focusedRule).toBeDefined();
      expect(focusedRule ?? "").not.toMatch(
        /(?:border(?:-[\w-]+)?-width|padding(?:-[\w-]+)?|margin(?:-[\w-]+)?|(?:min-|max-)?(?:block|inline)-size)\s*:/u,
      );
    },
  );

  it.each(actionHugTargets)(
    "switches %s to the danger hug when its field is invalid",
    (selector) => {
      const invalidSelector = `.paper-field--invalid ${selector}`;
      const invalidRule = leafRulesContaining(invalidSelector).find((declarations) =>
        declarations.includes("--paper-control-focus-color"),
      );

      expect(invalidRule).toBeDefined();
      expect(invalidRule ?? "").toMatch(
        /--paper-control-focus-color:\s*var\(--paper-color-status-danger\)/s,
      );
      expect(invalidRule ?? "").not.toMatch(
        /(?:border(?:-[\w-]+)?-width|padding(?:-[\w-]+)?|margin(?:-[\w-]+)?|(?:min-|max-)?(?:block|inline)-size)\s*:/u,
      );
    },
  );

  it.each(actionHugTargets)(
    "uses a system Highlight hug for %s in forced-colors mode",
    (selector) => {
      const forcedColorRule = leafRulesContaining(selector).find((declarations) =>
        /outline:\s*3px solid Highlight/s.test(declarations),
      );

      expect(formsCss).toContain("@media (forced-colors: active)");
      expect(forcedColorRule).toBeDefined();
      expect(forcedColorRule ?? "").toMatch(/outline:\s*3px solid Highlight/s);
      expect(forcedColorRule ?? "").toMatch(/outline-offset:\s*0/s);
      expect(forcedColorRule ?? "").toMatch(/box-shadow:\s*none/s);
    },
  );

  it("shares the multi-value chip surface between MultiAutocomplete and TagsInput", () => {
    const controlsSource = readFileSync(
      resolve("src/components/forms/controls.tsx"),
      "utf8",
    );

    expect(controlsSource).toMatch(
      /<Combobox\.Chips className="paper-combobox__chips">/s,
    );
    expect(controlsSource).toMatch(
      /export function TagsInput\([^)]*\)\s*\{\s*return <AutocompleteMultiInput/s,
    );
  });
});
