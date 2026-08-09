import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const formsCss = readFileSync(
  resolve("src/components/forms/forms.css"),
  "utf8",
);

describe("AmountWithUnit focus visual contract", () => {
  it("uses one parent-owned frame instead of two standalone control frames", () => {
    expect(formsCss).toMatch(
      /\.paper-compound-field\s*\{[^}]*border: var\(--paper-border-static-width\) solid var\(--paper-color-rule-default\);[^}]*border-block-end: 0;[^}]*background: var\(--paper-color-surface-paper\);/s,
    );
    expect(formsCss).toMatch(
      /\.paper-compound-field > \.paper-input\s*\{[^}]*border: 0;[^}]*border-block-end: var\(--paper-border-field-edge-width\) solid var\(--paper-color-rule-field-edge\);/s,
    );
    expect(formsCss).toMatch(
      /\.paper-compound-field > \.paper-compound-field__unit\s*\{[^}]*border-inline-start: var\(--paper-border-static-width\) solid var\(--paper-color-rule-default\);/s,
    );
    expect(formsCss).not.toMatch(
      /\.paper-field--invalid \.paper-compound-field__unit\s*\{/s,
    );
  });

  it("puts one low-opacity action hug around the whole compound control", () => {
    expect(formsCss).toMatch(
      /\.paper-compound-field:has\(> \.paper-input:focus-visible\)\s*\{[^}]*--paper-compound-focus-color: var\(--paper-color-focus-ring\);[^}]*outline: 3px solid color-mix\(in srgb, var\(--paper-compound-focus-color\) 10%, transparent\);[^}]*outline-offset: 0;/s,
    );

    const compoundFocusRule = formsCss.match(
      /\.paper-compound-field:has\(> \.paper-input:focus-visible\)\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(compoundFocusRule).not.toMatch(
      /(?:inline-size|block-size|width|height|margin|padding|border[^:]*-width)\s*:/u,
    );
  });

  it("suppresses child outlines while marking only the focused child edge", () => {
    expect(formsCss).toMatch(
      /\.paper-compound-field > \.paper-input:focus-visible\s*\{[^}]*border-block-end-color: var\(--paper-compound-focus-color\);[^}]*outline: 0;/s,
    );

    const childFocusRule = formsCss.match(
      /\.paper-compound-field > \.paper-input:focus-visible\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(childFocusRule).not.toMatch(
      /(?:inline-size|block-size|width|height|margin|padding|border[^:]*-width)\s*:/u,
    );
  });

  it("keeps the compound hug and focused edge in the invalid color", () => {
    expect(formsCss).toMatch(
      /\.paper-field--invalid \.paper-compound-field:has\(> \.paper-input:focus-visible\)\s*\{[^}]*--paper-compound-focus-color: var\(--paper-color-status-danger\);/s,
    );
  });

  it("uses a system focus color for both indicators in forced-colors mode", () => {
    expect(formsCss).toMatch(
      /@media \(forced-colors: active\)\s*\{[\s\S]*?\.paper-compound-field:has\(> \.paper-input:focus-visible\)\s*\{[^}]*--paper-compound-focus-color: Highlight;[^}]*outline: 3px solid Highlight;[^}]*outline-offset: 0;/s,
    );
    expect(formsCss).toMatch(
      /@media \(forced-colors: active\)\s*\{[\s\S]*?\.paper-compound-field > \.paper-input:focus-visible\s*\{[^}]*border-block-end-color: Highlight;[^}]*outline: 0;/s,
    );
  });
});
