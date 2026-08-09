import { createElement } from "react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";

import {
  Breadcrumb,
  Button,
  DRAWER_PLACEMENTS,
  Drawer,
  HeatmapChart,
  Navbar,
  Pagination,
  Sidebar,
  Table,
  Tabbar,
  Tabs,
  TabsList,
  TabsPanel,
  TabsRoot,
  TabsTab,
  VerticalTabbar,
  buttonClassName,
  chartPalette,
  chartSeries,
  type DrawerPlacement,
  type DrawerProps,
  type TabsListProps,
  type TabsOrientation,
  type TabsPanelProps,
  type TabsRootProps,
  type TabsTabProps,
  type TabsValue,
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

  it("publishes Tabs compound and named APIs with consumer-owned props", () => {
    const orientation: TabsOrientation = "horizontal";
    const value: TabsValue = "reading";
    const rootProps: TabsRootProps = { defaultValue: value, orientation };
    const listProps: TabsListProps = { "aria-label": "Log views" };
    const tabProps: TabsTabProps = { value, children: "Reading" };
    const panelProps: TabsPanelProps = { value, children: "Reading log" };

    expect(Tabs.Root).toBe(TabsRoot);
    expect(Tabs.List).toBe(TabsList);
    expect(Tabs.Tab).toBe(TabsTab);
    expect(Tabs.Panel).toBe(TabsPanel);
    expect(renderToString(
      <TabsRoot {...rootProps}>
        <TabsList {...listProps}>
          <TabsTab {...tabProps} />
        </TabsList>
        <TabsPanel {...panelProps} />
      </TabsRoot>,
    )).toContain("Reading log");
  });

  it("publishes Drawer and its placement contract", () => {
    const placement: DrawerPlacement = "end";
    const props: DrawerProps = {
      trigger: <button type="button">Review filters</button>,
      title: "Filters",
      placement,
      children: <p>Language and contest filters</p>,
    };

    expect(DRAWER_PLACEMENTS).toEqual(["start", "end"]);
    expect(renderToString(<Drawer {...props} />)).toContain("Review filters");
  });
});
