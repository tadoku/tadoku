import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";
import {
  legacyStyleguideRedirects,
  phaseThreeContentFixtures,
  phaseThreeExperimentDocuments,
  phaseThreeFoundationDocuments,
  phaseThreeGovernanceDocuments,
  phaseThreePatternDocuments,
} from "../src/catalog/phase-three-content";

describe("Phase 3 foundation and product content", () => {
  it("covers every canonical foundation from its Paper source", () => {
    expect(phaseThreeFoundationDocuments).toHaveLength(10);
    expect(phaseThreeFoundationDocuments.map((document) => document.route)).toEqual(
      expect.arrayContaining([
        "/foundations/color",
        "/foundations/iconography",
        "/foundations/brand",
      ]),
    );
    expect(phaseThreeFoundationDocuments.every((document) =>
      document.guidance.content.every((paragraph) => !paragraph.includes("initial catalogue stub")),
    )).toBe(true);
  });

  it("keeps the logging product pattern separate from its experiment", () => {
    expect(phaseThreePatternDocuments[0]).toMatchObject({
      kind: "pattern",
      lifecycle: "Stable",
      route: "/patterns/logging",
    });
    expect(phaseThreeExperimentDocuments[0]).toMatchObject({
      kind: "experiment",
      lifecycle: "Experimental",
      route: "/experiments/logging-v2",
    });
  });

  it("renders both deterministic product fixtures without a router", () => {
    for (const fixture of phaseThreeContentFixtures) {
      expect(renderToString(<>{fixture.render()}</>)).toContain("paper-");
    }
  });

  it("records lifecycle, deprecation, review, and design history guidance", () => {
    const content = phaseThreeGovernanceDocuments
      .flatMap((document) => document.guidance.content)
      .join(" ");
    expect(content).toMatch(/lifecycle|Experimental/u);
    expect(content).toContain("Deprecated");
    expect(content).toContain("review date");
    expect(content).toContain("Design history");
  });

  it("plans every non-root legacy styleguide route as an exact redirect", () => {
    expect(legacyStyleguideRedirects).toHaveLength(17);
    expect(legacyStyleguideRedirects).toEqual(
      expect.arrayContaining([
        { from: "/forms", to: "/components/forms/input" },
        { from: "/navigation", to: "/components/navigation/navbar" },
        { from: "/logging-v2", to: "/experiments/logging-v2" },
      ]),
    );
  });
});
