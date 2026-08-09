import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const baseCss = readFileSync(resolve("src/foundations/base.css"), "utf8");
const inputCss = readFileSync(
  resolve("src/components/forms/Input/input.css"),
  "utf8",
);

describe("Form control visual contract", () => {
  it("uses the same neutral perimeter and field edge for Input, TextArea, and Select", () => {
    expect(inputCss).toMatch(
      /\.paper-input\s*\{[^}]*border: var\(--paper-border-static-width\) solid var\(--paper-color-rule-default\);[^}]*border-block-end: var\(--paper-border-field-edge-width\) solid\s*var\(--paper-color-rule-field-edge\);/s,
    );

    const elementSpecificRule = inputCss.match(
      /\.paper-input:is\(input, textarea\)\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(elementSpecificRule ?? "").not.toMatch(/border(?:-[\w-]+)?\s*:/u);
  });

  it("strengthens only interactive controls on hover and preserves focus visibility", () => {
    expect(inputCss).toMatch(
      /\.paper-input:where\(input, textarea\):where\(:hover:not\(:disabled\):not\(\[readonly\]\)\)\s*\{[^}]*border-color: var\(--paper-color-rule-field-edge\);/s,
    );
    expect(inputCss).toMatch(
      /\.paper-input:where\(select\):where\(:hover:not\(:disabled\)\)\s*\{[^}]*border-color: var\(--paper-color-rule-field-edge\);/s,
    );
    expect(baseCss).toMatch(
      /:focus-visible\s*\{[^}]*outline: var\(--paper-focus-ring-width\) solid var\(--paper-color-focus-ring\);[^}]*outline-offset: var\(--paper-focus-ring-offset\);/s,
    );
  });

  it("does not let the CSS read-only pseudo-class restyle native Select", () => {
    expect(inputCss).toMatch(
      /\.paper-input:(?:is|where)\(input, textarea\)\[readonly\]\s*\{[^}]*border-block-end-color: var\(--paper-color-rule-default\);[^}]*background: var\(--paper-color-surface-raised\);/s,
    );
    expect(inputCss).not.toMatch(/\.paper-input:read-only\s*\{/u);
  });

  it("gives Paper inputs a zero-offset, low-opacity action hug on keyboard focus", () => {
    expect(inputCss).toMatch(
      /\.paper-input:focus-visible\s*\{[^}]*--paper-input-focus-color: var\(--paper-color-focus-ring\);[^}]*position: relative;[^}]*z-index: 1;[^}]*border-block-end-color: var\(--paper-input-focus-color\);[^}]*outline: 3px solid color-mix\(in srgb, var\(--paper-input-focus-color\) 10%, transparent\);[^}]*outline-offset: 0;/s,
    );

    const focusedRule = inputCss.match(
      /\.paper-input:focus-visible\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(focusedRule).not.toMatch(/border[^:]*-width:/u);
  });

  it("keeps invalid focus entirely in the danger state", () => {
    expect(inputCss).toMatch(
      /\.paper-field--invalid \.paper-input:focus-visible,\s*\.paper-input\[aria-invalid="true"\]:focus-visible\s*\{[^}]*--paper-input-focus-color: var\(--paper-color-status-danger\);/s,
    );

    const invalidFocusRule = inputCss.match(
      /\.paper-field--invalid \.paper-input:focus-visible,\s*\.paper-input\[aria-invalid="true"\]:focus-visible\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(invalidFocusRule).not.toMatch(/(?:border|outline|box-shadow)\s*:/u);
  });

  it("uses a system Highlight action hug in forced-colors mode", () => {
    expect(inputCss).toMatch(
      /@media \(forced-colors: active\)\s*\{[^}]*\.paper-input:focus-visible\s*\{[^}]*border-block-end-color: Highlight;[^}]*outline: 3px solid Highlight;[^}]*outline-offset: 0;[^}]*box-shadow: none;/s,
    );
  });
});
