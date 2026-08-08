import { createElement } from "react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { Button, buttonClassName, chartPalette, chartSeries } from "../src";
import {
  catalogRegistry,
  defineCatalogFixture,
  validateCatalogRegistry,
} from "../src/catalog";

describe("public source entries", () => {
  it("import without a browser and expose foundation values", () => {
    expect(chartPalette).toBeDefined();
    expect(chartSeries.length).toBeGreaterThan(0);
    expect(validateCatalogRegistry(catalogRegistry).valid).toBe(true);
  });

  it("server-renders a fixture imported through the catalog entry", () => {
    const fixture = defineCatalogFixture({
      id: "consumer.server-render",
      name: "Server render",
      description: "A fixed server-render consumer fixture.",
      tags: ["consumer"],
      themes: ["light"],
      densities: ["comfortable"],
      viewports: [{ id: "desktop", label: "Desktop", width: 1280, height: 800 }],
      deterministic: true,
      render: () => createElement("p", null, "Paper consumer"),
    });

    expect(renderToString(createElement("div", null, fixture.render()))).toContain(
      "Paper consumer",
    );
  });

  it("server-renders the public component entry without browser globals", () => {
    expect(renderToString(createElement(Button, null, "Save log"))).toContain(
      "Save log",
    );
    expect(buttonClassName({ variant: "destructive" })).toContain(
      "paper-button--destructive",
    );
  });
});
