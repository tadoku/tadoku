import { createElement } from "react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";

import {
  Breadcrumb,
  Button,
  HeatmapChart,
  Navbar,
  Pagination,
  Sidebar,
  Table,
  Tabbar,
  VerticalTabbar,
  buttonClassName,
  chartPalette,
  chartSeries,
} from "../src";
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

  it("server-renders the Phase 3 public entries without browser globals", () => {
    const links = [{ id: "logs", label: "Logs", href: "/logs", current: true }];

    expect(renderToString(createElement(Navbar, { brand: "Tadoku", brandHref: "/", navigation: [] }))).toContain("Tadoku");
    expect(renderToString(createElement(Sidebar, { label: "Sections", sections: [{ id: "main", title: "Main", links }] }))).toContain("Logs");
    expect(renderToString(createElement(Breadcrumb, { items: [{ id: "home", label: "Home" }] }))).toContain("Home");
    expect(renderToString(createElement(Tabbar, { label: "Views", links }))).toContain("Logs");
    expect(renderToString(createElement(VerticalTabbar, { label: "Views", links }))).toContain("Logs");
    expect(renderToString(createElement(Pagination, {
      currentPage: 1,
      totalPages: 2,
      getHref: (page) => `/logs?page=${page}`,
    }))).toContain("Page 1");
    expect(renderToString(
      <Table
        caption="Entries"
        columns={[{ id: "title", header: "Title", cell: (row) => row.title }]}
        rows={[{ title: "Convenience Store Woman" }]}
      />,
    )).toContain("Convenience Store Woman");
    expect(renderToString(createElement(HeatmapChart, {
      title: "Reading",
      columns: [{ id: "mon", label: "Monday" }],
      rows: [{ id: "week", label: "This week", cells: [{ value: 12 }] }],
    }))).toContain("12");
  });
});
