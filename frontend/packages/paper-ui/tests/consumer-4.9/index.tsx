import { createElement } from "react";
import { renderToString } from "react-dom/server";
import {
  Button,
  buttonClassName,
  chartPalette,
  chartSeries,
  type AutocompleteInputProps,
  type ButtonProps,
  type ChartSeries,
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
void buttonClassName({ variant: "outline" });
void renderToString(createElement(Button, buttonProps));
void autocompleteProps;
void toast;
void document;
void series;
void validateCatalogRegistry(registry);
void renderToString(createElement("div", null, fixture.render()));
