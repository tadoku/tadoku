import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { catalogRegistry } from "../src/catalog";

const documentById = (id: string) =>
  catalogRegistry.documents.find((document) => document.id === id);
const fixtureById = (id: string) =>
  catalogRegistry.fixtures.find((fixture) => fixture.id === id);

describe("published Tabs and Drawer catalogue contracts", () => {
  it.each([
    ["component.tabs", "/components/navigation/tabs", "tabs.content"],
    ["component.drawer", "/components/overlays/drawer", "drawer.filters"],
  ])("publishes useful documentation for %s", (id, route, fixtureId) => {
    const document = documentById(id);

    expect(document).toMatchObject({ route, lifecycle: "Stable" });
    expect(document?.fixtureIds).toContain(fixtureId);
    expect(document?.sections?.pageSections.length).toBeLessThanOrEqual(6);
    expect(document?.sections?.pageSections).toEqual([
      "usage",
      "examples",
      "variantsAndStates",
      "behavior",
      "contentGuidance",
      "accessibility",
    ]);
  });

  it("renders real Tabs and Drawer fixtures from the public APIs", () => {
    const tabs = fixtureById("tabs.content");
    const drawer = fixtureById("drawer.filters");

    expect(tabs?.code).toContain('import { Tabs } from "paper-ui"');
    expect(renderToString(<>{tabs?.render()}</>)).toContain("paper-tabs__tab");
    expect(drawer?.code).toContain('import { Button, Drawer } from "paper-ui"');
    expect(renderToString(<>{drawer?.render()}</>)).toContain("Review filters");
  });
});

describe("Modal compositional catalogue extension", () => {
  it("documents a controlled, footerless, initial-focus composition", () => {
    const document = documentById("component.modal");
    const fixture = fixtureById("modal.composable-search");
    const code = fixture?.code ?? "";

    expect(document?.fixtureIds).toContain("modal.composable-search");
    expect(code).toContain("trigger={<Button");
    expect(code).toContain("open={open}");
    expect(code).toContain("onOpenChange={setOpen}");
    expect(code).toContain("initialFocus={searchRef}");
    expect(code).toContain("footer={null}");
    expect(renderToString(<>{fixture?.render()}</>)).toContain("Search Paper");
  });
});
