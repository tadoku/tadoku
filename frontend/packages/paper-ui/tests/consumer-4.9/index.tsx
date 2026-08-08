import { createElement } from "react";
import { renderToString } from "react-dom/server";
import {
  Button,
  DRAWER_PLACEMENTS,
  Drawer,
  Tabs,
  buttonClassName,
  chartPalette,
  chartSeries,
  type AutocompleteInputProps,
  type ButtonProps,
  type ChartSeries,
  type DrawerProps,
  type RadioSelectProps,
  type TabsRootProps,
  type ToastOptions,
} from "paper-ui";
import {
  CATALOG_LIFECYCLES,
  catalogRegistry,
  defineCatalogFixture,
  validateCatalogRegistry,
  type CatalogDocument,
  type CatalogFixture,
  type CatalogRegistry,
} from "paper-ui/catalog";

const registry: CatalogRegistry = catalogRegistry;
const document: CatalogDocument | undefined = registry.documents[0];
const series: ChartSeries | undefined = chartSeries[0];
const fixture: CatalogFixture = defineCatalogFixture({
  id: "consumer.ts49",
  name: "TypeScript 4.9 consumer",
  description: "Proves public declarations stay compiler-compatible.",
  tags: ["consumer"],
  themes: ["light"],
  densities: ["comfortable"],
  viewports: [{ id: "desktop", label: "Desktop", width: 1280, height: 800 }],
  deterministic: true,
  render: () => createElement("p", null, "TypeScript 4.9"),
});

void CATALOG_LIFECYCLES;
void chartPalette;
const buttonProps: ButtonProps = { variant: "outline", children: "Review log" };
const autocompleteProps: AutocompleteInputProps<{ id: string; label: string }> = {
  name: "language",
  label: "Language",
  options: [{ id: "ja", label: "Japanese" }],
  format: (option) => option.label,
  getId: (option) => option.id,
};
const toast: ToastOptions = {
  title: "Entry saved",
  description: "12 pages added.",
  priority: "low",
};
const tabsProps: TabsRootProps = { defaultValue: "summary" };
const drawerProps: DrawerProps = {
  trigger: createElement(Button, { variant: "outline" }, "Review filters"),
  title: "Filters",
  children: createElement("p", null, "Entry filters"),
};
const radioSelectProps: RadioSelectProps = {
  name: "viewport",
  label: "Preview size",
  variant: "segmented",
  required: true,
  options: [
    { value: "phone", label: "Phone" },
    { value: "tablet", label: "Tablet" },
    { value: "desktop", label: "Desktop" },
  ],
};
void buttonClassName({ variant: "outline" });
void renderToString(createElement(Button, buttonProps));
void autocompleteProps;
void toast;
void DRAWER_PLACEMENTS;
void renderToString(
  createElement(
    Tabs.Root,
    tabsProps,
    createElement(
      Tabs.List,
      { "aria-label": "Views" },
      createElement(Tabs.Tab, { value: "summary" }, "Summary"),
    ),
    createElement(Tabs.Panel, { value: "summary" }, "Reading summary"),
  ),
);
void renderToString(createElement(Drawer, drawerProps));
void radioSelectProps;
void document;
void series;
void validateCatalogRegistry(registry);
void renderToString(createElement("div", null, fixture.render()));
