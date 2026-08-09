import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const navigationCss = readFileSync(
  resolve("src/components/navigation/navigation.css"),
  "utf8",
);

describe("Navbar visual contract", () => {
  it("centers custom brand content without inline baseline space", () => {
    expect(navigationCss).toMatch(
      /\.paper-navbar__brand\s*\{(?=[^}]*display:\s*flex;)(?=[^}]*align-items:\s*center;)[^}]*\}/su,
    );
  });
});
