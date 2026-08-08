import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const modalCss = readFileSync(
  resolve("src/components/overlays/Modal/modal.css"),
  "utf8",
);
const tokensCss = readFileSync(
  resolve("src/foundations/tokens.css"),
  "utf8",
);
const actionMenuCss = readFileSync(
  resolve("src/components/overlays/ActionMenu/action-menu.css"),
  "utf8",
);

describe("Modal visual contract", () => {
  it("uses the theme-aware translucent scrim and blurs only the backdrop", () => {
    expect(tokensCss).toContain(
      "--paper-color-surface-scrim: var(--paper-private-scrim);",
    );
    expect(tokensCss).toContain("--paper-private-scrim: rgb(33 26 29 / 28%);");
    expect(tokensCss).toContain("--paper-private-scrim: rgb(0 0 0 / 38%);");
    expect(modalCss).toMatch(
      /\.paper-modal__backdrop\s*\{[^}]*background: var\(--paper-color-surface-scrim\);[^}]*-webkit-backdrop-filter: blur\(0\.375rem\);[^}]*backdrop-filter: blur\(0\.375rem\);/s,
    );
    expect(modalCss).toMatch(
      /@media \(forced-colors: active\)\s*\{[^}]*\.paper-modal__backdrop\s*\{[^}]*backdrop-filter: none;/s,
    );
    expect(actionMenuCss).not.toContain("backdrop-filter");
  });
});
