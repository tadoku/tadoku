import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const recipes = readFileSync("styles/recipes.css", "utf8");
const utilities = readFileSync("styles/utilities.css", "utf8");

describe("foundation composition contracts", () => {
  it("aligns accent rails with the outer perimeter only on bordered surfaces", () => {
    expect(recipes).toMatch(/\.paper-accent-rail\s*\{[^}]*--paper-accent-rail-outset:\s*0px;/s);
    expect(recipes).toMatch(/\.paper-surface-card\.paper-accent-rail\s*\{[^}]*--paper-accent-rail-outset:\s*var\(--paper-border-static-width\);/s);
    expect(recipes).toMatch(/inset-inline-start:\s*calc\(-1 \* var\(--paper-accent-rail-outset\)\);/);
    expect(recipes).not.toMatch(/border-left:\s*0;/);
  });
  it("ships the documented wrapping and readable layout utilities", () => {
    expect(utilities).toMatch(/\.paper-cluster\s*\{[^}]*flex-wrap:\s*wrap;/s);
    expect(utilities).toMatch(/\.paper-stack\s*\{[^}]*display:\s*flex;[^}]*flex-direction:\s*column;/s);
    expect(utilities).toMatch(/\.paper-measure\s*\{[^}]*max-inline-size:\s*65ch;/s);
  });
});
