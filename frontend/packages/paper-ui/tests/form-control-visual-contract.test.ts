import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const baseCss = readFileSync(resolve("src/foundations/base.css"), "utf8");
const inputCss = readFileSync(
  resolve("src/components/forms/Input/input.css"),
  "utf8",
);

describe("Form control visual contract", () => {
  it("gives text-entry controls the outline button's neutral perimeter treatment", () => {
    expect(inputCss).toMatch(
      /\.paper-input:is\(input, textarea\)\s*\{[^}]*border-color: var\(--paper-color-rule-default\);[^}]*border-block-end-color: var\(--paper-color-rule-field-edge\);/s,
    );
  });

  it("strengthens only interactive controls on hover and preserves focus visibility", () => {
    expect(inputCss).toMatch(
      /\.paper-input:where\(input, textarea\):where\(:hover:not\(:disabled\):not\(:read-only\)\)\s*\{[^}]*border-color: var\(--paper-color-rule-field-edge\);/s,
    );
    expect(baseCss).toMatch(
      /:focus-visible\s*\{[^}]*outline: var\(--paper-focus-ring-width\) solid var\(--paper-color-focus-ring\);[^}]*outline-offset: var\(--paper-focus-ring-offset\);/s,
    );
  });

  it("preserves the approved native Select edge treatment", () => {
    expect(inputCss).toMatch(
      /\.paper-input\s*\{[^}]*border: var\(--paper-border-static-width\) solid var\(--paper-color-rule-field-edge\);[^}]*border-block-end: var\(--paper-border-field-edge-width\) solid\s*var\(--paper-color-rule-action-edge\);/s,
    );
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
