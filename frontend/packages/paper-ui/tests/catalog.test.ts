import { createElement } from "react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";

import {
  catalogRegistry,
  CATALOG_LIFECYCLES,
  CatalogValidationError,
  createCatalogRegistry,
  defineCatalogFixture,
  REQUIRED_COMPONENT_SECTION_KEYS,
  validateCatalogFixture,
  validateCatalogRegistry,
  type CatalogDocument,
  type CatalogFixture,
  type ComponentDocumentationSections,
} from "../src/catalog";

function requiredSections(): ComponentDocumentationSections {
  return {
    required: Object.fromEntries(
      REQUIRED_COMPONENT_SECTION_KEYS.map((key) => [
        key,
        { heading: `Heading for ${key}`, content: [`Content for ${key}.`] },
      ]),
    ) as unknown as ComponentDocumentationSections["required"],
  };
}

function fixture(overrides: Partial<CatalogFixture> = {}): CatalogFixture {
  return {
    id: "button.default",
    name: "Default button",
    description: "A deterministic default-action example.",
    tags: ["button", "default"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: [
      { id: "phone", label: "Phone", width: 390, height: 844 },
      { id: "desktop", label: "Desktop", width: 1280, height: 800 },
    ],
    deterministic: true,
    render: () => createElement("button", { type: "button" }, "Save log"),
    ...overrides,
  };
}

function document(overrides: Partial<CatalogDocument> = {}): CatalogDocument {
  return {
    id: "component.button",
    route: "/components/actions/button",
    name: "Button",
    kind: "component",
    category: "actions",
    aliases: [],
    summary: "Triggers an immediate action.",
    keywords: ["button", "action"],
    lifecycle: "Stable",
    reviewDate: "2026-08-08",
    sourcePath: "src/components/actions/Button/Button.tsx",
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: ["Trigger an immediate action."],
      whenNotToUse: ["Do not use it for navigation."],
      content: ["Use a short verb phrase."],
      commonMistakes: ["Do not hide the action behind vague copy."],
    },
    accessibility: {
      requirements: ["The button has an accessible name."],
      keyboard: ["Enter and Space activate the button."],
      knownConstraints: ["Loading labels need an announcement strategy."],
    },
    api: {
      react: ["Button"],
      cssClasses: ["btn"],
      publicTypes: ["ButtonProps"],
      defaults: ["type is button"],
      invalidCombinations: ["Navigation must use an anchor."],
    },
    fixtureIds: ["button.default"],
    dependencies: { documents: [], packages: ["react"] },
    migration: {
      legacy: ["ui/components/Button"],
      notes: ["Rename primary to default."],
    },
    changelog: [{ date: "2026-08-08", note: "Documented the Stable contract." }],
    behaviorTestIds: ["button.activation"],
    sections: requiredSections(),
    ...overrides,
  };
}

function registry(
  documents: readonly CatalogDocument[] = [document()],
  fixtures: readonly CatalogFixture[] = [fixture()],
) {
  return { documents, fixtures, redirects: [] };
}

function codes(result: ReturnType<typeof validateCatalogRegistry>): string[] {
  return result.issues.map((issue) => issue.code);
}

describe("the initial catalogue", () => {
  it("is valid and provides the Phase 1 foundation and governance routes", () => {
    const result = validateCatalogRegistry(catalogRegistry);
    const routes = catalogRegistry.documents.map((entry) => entry.route);

    expect(result).toEqual({ valid: true, issues: [] });
    expect(routes).toContain("/foundations/principles");
    expect(routes).toContain("/foundations/color");
    expect(routes).toContain("/foundations/brand");
    expect(routes).toContain("/contributing");
    expect(routes).toContain("/changelog");
  });

  it("keeps the lifecycle and 16-section sequence exact", () => {
    expect(CATALOG_LIFECYCLES).toEqual(["Experimental", "Stable", "Deprecated"]);
    expect(REQUIRED_COMPONENT_SECTION_KEYS).toHaveLength(16);
    expect(REQUIRED_COMPONENT_SECTION_KEYS[0]).toBe("overview");
    expect(REQUIRED_COMPONENT_SECTION_KEYS[15]).toBe("lifecycle");
  });

  it("does not model a project owner for catalogue documents", () => {
    expect(catalogRegistry.documents.every((entry) => !("owner" in entry))).toBe(true);
  });
});

describe("fixture validation", () => {
  it("accepts router-neutral deterministic render functions without executing them", () => {
    let renderCount = 0;
    const candidate = defineCatalogFixture(
      fixture({
        render: () => {
          renderCount += 1;
          return createElement("span", null, "Fixed content");
        },
      }),
    );

    expect(validateCatalogFixture(candidate)).toEqual({ valid: true, issues: [] });
    expect(renderCount).toBe(0);
    expect(renderToString(createElement("div", null, candidate.render()))).toContain(
      "Fixed content",
    );
  });

  it.each([
    ["randomness", () => String(Math.random()), "non-deterministic-fixture"],
    ["current time", () => String(Date.now()), "non-deterministic-fixture"],
    ["network", () => fetch("/fixture-data") as never, "non-deterministic-fixture"],
    [
      "router location",
      () => String(window.location.pathname),
      "router-dependent-fixture",
    ],
  ])("rejects %s dependencies", (_name, render, expectedCode) => {
    const result = validateCatalogFixture(fixture({ render }));

    expect(result.valid).toBe(false);
    expect(result.issues.map((issue) => issue.code)).toContain(expectedCode);
  });

  it("requires the deterministic review marker and supported preview metadata", () => {
    const result = validateCatalogFixture(
      fixture({
        deterministic: false as true,
        themes: ["sepia" as "light"],
        densities: [],
        viewports: [],
      }),
    );

    expect(result.valid).toBe(false);
    expect(result.issues.map((issue) => issue.code)).toEqual(
      expect.arrayContaining(["invalid-fixture-marker", "invalid-fixture"]),
    );
  });
});

describe("registry validation", () => {
  it("accepts a complete Stable component contract", () => {
    expect(validateCatalogRegistry(registry())).toEqual({ valid: true, issues: [] });
  });

  it("reports every missing required Stable section", () => {
    const candidate = document({ sections: { required: {} } });
    const result = validateCatalogRegistry(registry([candidate]));

    expect(result.valid).toBe(false);
    expect(
      result.issues.filter((issue) => issue.code === "missing-stable-section"),
    ).toHaveLength(16);
  });

  it("requires Stable fixtures, behavior tests, and accessibility evidence", () => {
    const candidate = document({
      fixtureIds: [],
      behaviorTestIds: [],
      accessibility: { requirements: [], keyboard: [], knownConstraints: [] },
    });
    const result = validateCatalogRegistry(registry([candidate], []));

    expect(codes(result)).toEqual(
      expect.arrayContaining([
        "missing-stable-fixture",
        "missing-stable-behavior-test",
        "missing-stable-accessibility",
      ]),
    );
  });

  it("enforces core metadata and real review dates", () => {
    const candidate = document({
      reviewDate: "2026-02-30",
      keywords: [],
    });
    const result = validateCatalogRegistry(registry([candidate]));

    expect(codes(result)).toEqual(
      expect.arrayContaining(["missing-metadata", "invalid-date"]),
    );
    expect(result.issues.map((issue) => issue.path)).toEqual(
      expect.arrayContaining([
        "documents[0].reviewDate",
        "documents[0].keywords",
      ]),
    );
  });

  it("keeps component-specific content outside the required section contract", () => {
    const sections = requiredSections();
    const requiredWithExtra = {
      ...sections.required,
      visualRegressionNotes: {
        heading: "Visual regression notes",
        content: ["Review the baseline."],
      },
    } as ComponentDocumentationSections["required"];
    const result = validateCatalogRegistry(
      registry([document({ sections: { required: requiredWithExtra } })]),
    );

    expect(codes(result)).toContain("invalid-section");
  });

  it("keeps optional sections in required-sequence order", () => {
    const result = validateCatalogRegistry(
      registry([
        document({
          sections: {
            ...requiredSections(),
            optional: [
              { after: "apiReference", heading: "Advanced API", content: ["Details."] },
              { after: "anatomy", heading: "Layout", content: ["Details."] },
            ],
          },
        }),
      ]),
    );

    expect(codes(result)).toContain("invalid-section");
  });

  it("rejects invalid kind, category, lifecycle, and canonical routes", () => {
    const candidate = document({
      category: "widgets" as "actions",
      lifecycle: "stable" as "Stable",
      route: "/buttons",
    });
    const invalidKind = document({
      id: "widget.action",
      route: "/components/actions/widget",
      kind: "widget" as "component",
    });
    const result = validateCatalogRegistry(registry([candidate, invalidKind]));

    expect(codes(result)).toEqual(
      expect.arrayContaining([
        "invalid-kind",
        "invalid-category",
        "invalid-lifecycle",
        "invalid-route",
      ]),
    );
  });

  it("rejects duplicate document IDs, routes, slugs, aliases, and fixture IDs", () => {
    const first = document({ aliases: ["/old-button"] });
    const duplicateIdAndRoute = document();
    const duplicateSlug = document({
      id: "component.forms-button",
      route: "/components/forms/button",
      category: "forms",
      aliases: ["/old-button"],
    });
    const result = validateCatalogRegistry(
      registry(
        [first, duplicateIdAndRoute, duplicateSlug],
        [fixture(), fixture()],
      ),
    );

    expect(codes(result)).toEqual(
      expect.arrayContaining([
        "duplicate-document-id",
        "duplicate-route",
        "duplicate-slug",
        "duplicate-alias",
        "duplicate-fixture-id",
      ]),
    );
  });

  it("resolves fixture, document dependency, redirect, and replacement references", () => {
    const candidate = document({
      fixtureIds: ["button.missing"],
      dependencies: { documents: ["foundation.missing"], packages: [] },
    });
    const result = validateCatalogRegistry({
      documents: [candidate],
      fixtures: [fixture()],
      redirects: [{ from: "/buttons", to: "/components/actions/missing" }],
    });

    expect(codes(result)).toEqual(
      expect.arrayContaining([
        "unknown-fixture-reference",
        "unknown-document-reference",
      ]),
    );
  });

  it("requires a registered replacement and removal guidance for Deprecated entries", () => {
    const deprecated = document({
      lifecycle: "Deprecated",
      migration: { legacy: [], notes: [] },
    });
    const result = validateCatalogRegistry(registry([deprecated]));

    expect(codes(result)).toContain("invalid-deprecation");
    expect(
      result.issues.filter((issue) => issue.code === "invalid-deprecation"),
    ).toHaveLength(2);
  });

  it("accepts deprecation only with a different registered replacement", () => {
    const replacement = document();
    const deprecated = document({
      id: "component.old-button",
      route: "/components/actions/old-button",
      lifecycle: "Deprecated",
      migration: {
        legacy: ["ui/OldButton"],
        notes: ["Use Button."],
        replacementDocumentId: replacement.id,
        removalGuidance: "Remove after all applications migrate in Phase 8.",
      },
    });

    expect(validateCatalogRegistry(registry([replacement, deprecated]))).toEqual({
      valid: true,
      issues: [],
    });
  });

  it("throws structured issues when constructing an invalid registry", () => {
    expect(() => createCatalogRegistry(registry([document({ fixtureIds: [] })], []))).toThrow(
      CatalogValidationError,
    );
  });
});
