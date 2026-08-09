import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const recipesCss = readFileSync(resolve("styles/recipes.css"), "utf8");
const feedbackCss = readFileSync(
  resolve("src/components/feedback/feedback.css"),
  "utf8",
);
const tokensCss = readFileSync(
  resolve("src/foundations/tokens.css"),
  "utf8",
);

const RULE_TOKEN = /var\((--paper-color-rule-[^)]+)\)/;

function withoutLayer(css: string): string {
  return css
    .replace(/^\s*@layer\s+[^{}]+\{/, "")
    .replace(/\}\s*$/, "");
}

function effectiveBorderToken(classNames: readonly string[]): string | undefined {
  let winner:
    | { readonly specificity: number; readonly order: number; readonly token: string }
    | undefined;
  let order = 0;

  for (const css of [recipesCss, feedbackCss]) {
    for (const match of withoutLayer(css).matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
      const selectors = match[1].split(",");
      const declarations = match[2];

      for (const selector of selectors) {
        const selectorClasses = [...selector.matchAll(/\.([\w-]+)/g)].map(
          (classMatch) => classMatch[1],
        );
        if (
          selectorClasses.length === 0 ||
          !selectorClasses.every((className) => classNames.includes(className))
        ) {
          continue;
        }

        for (const declaration of declarations.matchAll(
          /(?:^|;)\s*(border(?:-color)?)\s*:\s*([^;]+)/g,
        )) {
          const token = declaration[2].match(RULE_TOKEN)?.[1];
          if (!token) continue;

          const candidate = {
            specificity: selectorClasses.length,
            order: order++,
            token,
          };
          if (
            !winner ||
            candidate.specificity > winner.specificity ||
            (candidate.specificity === winner.specificity &&
              candidate.order > winner.order)
          ) {
            winner = candidate;
          }
        }
      }
    }
  }

  return winner?.token;
}

describe("floating Surface visual contract", () => {
  it("provides a theme-aware strong structural rule token", () => {
    expect(tokensCss.match(/:root,[\s\S]*?\[data-theme="light"\]\s*\{([\s\S]*?)\n\}/)?.[1])
      .toContain("--paper-color-rule-strong: var(--paper-private-rule-strong);");
    expect(tokensCss.match(/\[data-theme="dark"\]\s*\{([\s\S]*?)\n\}/)?.[1])
      .toContain("--paper-color-rule-strong: var(--paper-private-rule-strong);");
  });

  it.each(["floating", "showcase"] as const)(
    "keeps the strong structural outline on %s Surface compositions after the component styles load",
    (elevation) => {
      expect(
        effectiveBorderToken([
          "paper-surface-card",
          `paper-elevation-${elevation}`,
        ]),
      ).toBe("--paper-color-rule-strong");
    },
  );
});
