import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const buttonCss = readFileSync(
  resolve("src/components/actions/Button/button.css"),
  "utf8",
);
const tokensCss = readFileSync(
  resolve("src/foundations/tokens.css"),
  "utf8",
);

describe("Button visual contract", () => {
  it("maps current variants to the Button Grammar perimeter treatments", () => {
    expect(buttonCss).toMatch(
      /\.paper-button\s*\{[^}]*border: var\(--paper-border-static-width\) solid var\(--paper-color-action-default\);[^}]*border-block-end-color: var\(--paper-color-rule-action-edge\);[^}]*border-block-end-width: var\(--paper-border-action-edge-width\);/s,
    );
    expect(buttonCss).toMatch(
      /\.paper-button--outline\s*\{[^}]*border-color: var\(--paper-color-rule-default\);[^}]*border-block-end-color: var\(--paper-color-rule-field-edge\);/s,
    );
    expect(buttonCss).toMatch(
      /\.paper-button--ghost\s*\{[^}]*border-color: transparent;[^}]*color: var\(--paper-color-text-ink\);/s,
    );
    expect(buttonCss).toMatch(
      /\.paper-button--link\s*\{[^}]*border: 0;[^}]*color: var\(--paper-color-text-link\);[^}]*background: transparent;/s,
    );
    expect(buttonCss).toMatch(
      /\.paper-button--destructive\s*\{[^}]*border-color: var\(--paper-color-action-destructive\);[^}]*border-block-end-color: var\(--paper-color-rule-destructive-action-edge\);/s,
    );
    expect(tokensCss).toContain(
      "--paper-color-rule-destructive-action-edge: var(--paper-private-danger-edge);",
    );
  });

  it("uses neutral hover and compact geometry from the visual specimen", () => {
    expect(buttonCss).toMatch(
      /\.paper-button--outline:hover:not\(:disabled\)[^{]*\{[^}]*background: var\(--paper-color-action-neutral-hover\);/s,
    );
    expect(buttonCss).toMatch(
      /\[data-density="compact"\] \.paper-button\s*\{[^}]*min-block-size: 2\.25rem;[^}]*padding: 0\.375rem 0\.6875rem 0\.3125rem;[^}]*font-size: 0\.875rem;/s,
    );
  });
});
