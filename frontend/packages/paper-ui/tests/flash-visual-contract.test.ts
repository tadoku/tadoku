import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const feedbackCss = readFileSync(
  resolve("src/components/feedback/feedback.css"),
  "utf8",
);

describe("Flash visual contract", () => {
  it("draws a rectangular status rail independently of the perimeter", () => {
    const flashRule = feedbackCss.match(
      /\.paper-flash\s*\{(?<declarations>[^}]*)\}/s,
    )?.groups?.declarations;

    expect(flashRule).toMatch(/position:\s*relative;/);
    expect(flashRule).toMatch(
      /border:\s*1px solid var\(--paper-color-rule-default\);/,
    );
    expect(flashRule).toMatch(
      /--paper-flash-rail-color:\s*var\(--paper-color-status-information\);/,
    );
    expect(flashRule).not.toMatch(/border-inline-start(?:-color)?:/);

    expect(feedbackCss).toMatch(
      /\.paper-flash::before\s*\{[^}]*position:\s*absolute;[^}]*inset-block:\s*-1px;[^}]*inset-inline-start:\s*-1px;[^}]*inline-size:\s*4px;[^}]*background:\s*var\(--paper-flash-rail-color\);[^}]*content:\s*"";/s,
    );
  });

  it.each([
    ["success", "success"],
    ["warning", "warning"],
    ["danger", "danger"],
  ])("maps the %s variant to its semantic rail color", (variant, status) => {
    expect(feedbackCss).toMatch(
      new RegExp(
        `\\.paper-flash--${variant}\\s*\\{[^}]*--paper-flash-rail-color:\\s*var\\(--paper-color-status-${status}\\);`,
        "s",
      ),
    );
  });

  it("keeps the rectangular rail visible in forced-colors mode", () => {
    expect(feedbackCss).toMatch(
      /@media \(forced-colors: active\)\s*\{[^}]*\.paper-flash::before\s*\{[^}]*background:\s*CanvasText;/s,
    );
  });
});
