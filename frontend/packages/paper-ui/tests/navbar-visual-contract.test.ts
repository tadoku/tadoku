import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const navigationCss = readFileSync(
  resolve("src/components/navigation/navigation.css"),
  "utf8",
);
const tokensCss = readFileSync(resolve("src/foundations/tokens.css"), "utf8");

function declaration(selector: string, property: string) {
  const rule = [...navigationCss.matchAll(/([^{}]+)\{([^{}]*)\}/gu)]
    .find((match) => match[1].split(",").some((value) => value.trim() === selector));
  const value = rule?.[2].match(new RegExp(`(?:^|;)\\s*${property}:\\s*([^;]+)`))?.[1];
  if (!value) throw new Error(`Missing ${property} on ${selector}`);
  return value.match(/var\((--[^)]+)\)/u)?.[1] ?? value;
}

function luminance(hex: string) {
  const rgb = hex.slice(1).match(/../gu)!.map((channel) => {
    const value = parseInt(channel, 16) / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  return rgb[0] * 0.2126 + rgb[1] * 0.7152 + rgb[2] * 0.0722;
}

describe("Navbar visual contract", () => {
  it("centers custom brand content without inline baseline space", () => {
    expect(navigationCss).toMatch(
      /\.paper-navbar__brand\s*\{(?=[^}]*display:\s*flex;)(?=[^}]*align-items:\s*center;)[^}]*\}/su,
    );
  });

  it.each(["light", "dark"])("keeps navigation state marks distinguishable in %s mode", (theme) => {
    const themeCss = tokensCss.match(new RegExp(`\\[data-theme="${theme}"\\]\\s*\\{([^}]+)\\}`))![1];
    const tokens = Object.fromEntries([...themeCss.matchAll(/(--[\w-]+):\s*([^;]+);/gu)].map((match) => [match[1], match[2]]));
    const color = (token: string): string => {
      const value = tokens[token];
      const reference = value.match(/^var\((--[^)]+)\)$/u)?.[1];
      return reference ? color(reference) : value;
    };
    const marks = [
      ['.paper-navbar__link[aria-current="page"]', "box-shadow", ["surface-paper", "action-neutral-hover"]],
      ['.paper-navbar__mobile-link[aria-current="page"]', "box-shadow", ["surface-paper", "action-neutral-hover"]],
      ['.paper-sidebar__link[aria-current="page"]::before', "background", ["surface-paper"]],
      ['.paper-tabbar__list--horizontal .paper-tabbar__link[aria-current="page"]', "box-shadow", ["surface-canvas", "surface-raised"]],
      ['.paper-tabbar__list--vertical .paper-tabbar__link[aria-current="page"]', "box-shadow", ["surface-canvas", "surface-raised"]],
      ['.paper-pagination__page[aria-current="page"]', "box-shadow", ["surface-canvas", "surface-paper", "action-neutral-hover"]],
    ] as const;
    for (const [selector, property, surfaces] of marks) {
      const foreground = luminance(color(declaration(selector, property)));
      for (const surface of surfaces) {
        const background = luminance(color(`--paper-color-${surface}`));
        const contrast = (Math.max(foreground, background) + 0.05) / (Math.min(foreground, background) + 0.05);
        expect.soft(contrast, `${selector} against ${surface}`).toBeGreaterThanOrEqual(3);
      }
    }
  });
});
