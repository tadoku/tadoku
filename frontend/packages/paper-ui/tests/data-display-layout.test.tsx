import { render, screen, within } from "@testing-library/react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import {
  HeatmapChart,
  Table,
  type HeatmapChartProps,
  type TableColumn,
} from "../src/components/data-display";
import {
  phaseThreeDataLayoutDocuments,
  phaseThreeDataLayoutFixtures,
} from "../src/catalog/phase-three-data-layout";
import { REQUIRED_COMPONENT_SECTION_KEYS } from "../src/catalog/schema";
import { validateCatalogRegistry } from "../src/catalog/validation";

interface ReadingRow {
  readonly id: string;
  readonly title: string;
  readonly pages: number;
}

const columns: readonly TableColumn<ReadingRow>[] = [
  {
    id: "title",
    header: "Title",
    rowHeader: true,
    cell: (row) => row.title,
  },
  {
    id: "pages",
    header: "Pages",
    align: "end",
    cell: (row) => row.pages,
  },
];

const rows: readonly ReadingRow[] = [
  { id: "book-1", title: "Convenience Store Woman", pages: 73 },
  { id: "book-2", title: "The Three-Body Problem", pages: 42 },
];

const heatmapProps: HeatmapChartProps = {
  title: "Daily reading",
  description: "Pages read this week.",
  columns: [
    { id: "mon", label: "Monday", shortLabel: "Mon" },
    { id: "tue", label: "Tuesday", shortLabel: "Tue" },
    { id: "wed", label: "Wednesday", shortLabel: "Wed" },
    { id: "thu", label: "Thursday", shortLabel: "Thu" },
  ],
  rows: [
    {
      id: "week-1",
      label: "Aug 3",
      cells: [{ value: 0 }, { value: 50 }, { value: 100 }, { value: null }],
    },
  ],
  domain: [0, 100],
  formatValue: (value) => (value === null ? "No data" : `${value} pages`),
};

describe("Table", () => {
  it("renders native caption, column headers, and identifying row headers", () => {
    render(
      <Table
        caption="Recent reading"
        columns={columns}
        rows={rows}
        getRowKey={(row) => row.id}
      />,
    );

    const region = screen.getByRole("region", { name: "Recent reading" });
    const table = within(region).getByRole("table", { name: "Recent reading" });
    expect(region).toHaveAttribute("tabindex", "0");
    expect(region).toHaveStyle("--paper-table-min-inline-size: 30rem");
    expect(within(table).getAllByRole("columnheader")).toHaveLength(2);
    expect(within(table).getByRole("rowheader", { name: "Convenience Store Woman" })).toHaveAttribute(
      "scope",
      "row",
    );
    expect(within(table).getByRole("cell", { name: "73" })).toHaveClass(
      "paper-table__cell--end",
    );
  });

  it("keeps table structure and a specific message when rows are empty", () => {
    render(
      <Table
        caption="Recent reading"
        captionVisibility="screen-reader"
        columns={columns}
        rows={[]}
        emptyMessage="No reading logged yet."
      />,
    );

    expect(screen.getByRole("table", { name: "Recent reading" })).toBeInTheDocument();
    expect(screen.getByText("No reading logged yet.")).toHaveAttribute("colspan", "2");
    expect(screen.getByText("Recent reading")).toHaveClass(
      "paper-table__caption--screen-reader",
    );
  });

  it("rejects a table without columns instead of emitting invalid structure", () => {
    expect(() =>
      renderToStaticMarkup(<Table caption="Invalid" columns={[]} rows={rows} />),
    ).toThrow("Table requires at least one column.");
  });
});

describe("HeatmapChart", () => {
  it("uses native table semantics and full accessible labels for abbreviated columns", () => {
    render(<HeatmapChart {...heatmapProps} />);

    const figure = screen.getByRole("figure", { name: "Daily reading" });
    expect(figure).toHaveAccessibleDescription("Pages read this week.");
    const region = within(figure).getByRole("region", {
      name: "Daily reading data table",
    });
    expect(region).toHaveAttribute("tabindex", "0");
    expect(within(region).getByRole("table", { name: "Daily reading data" })).toBeInTheDocument();
    expect(within(region).getByRole("rowheader", { name: "Aug 3" })).toHaveAttribute(
      "scope",
      "row",
    );
    expect(within(region).getByRole("columnheader", { name: "Monday" })).toHaveTextContent(
      "Mon",
    );
  });

  it("prints every value and pairs chart-token intensity with a missing-data pattern", () => {
    render(<HeatmapChart {...heatmapProps} colorIndex={3} />);

    const figure = screen.getByRole("figure", { name: "Daily reading" });
    expect(figure.getAttribute("style")).toContain(
      "--paper-heatmap-color: var(--paper-color-chart-4)",
    );
    expect(screen.getByText("0 pages").closest("td")).toHaveAttribute("data-level", "0");
    expect(screen.getByText("50 pages").closest("td")).toHaveAttribute("data-level", "2");
    expect(screen.getByText("100 pages").closest("td")).toHaveAttribute("data-level", "4");
    expect(screen.getByText("No data").closest("td")).toHaveAttribute(
      "data-level",
      "missing",
    );
  });

  it("validates dimensions, finite values, and domain order", () => {
    expect(() =>
      renderToStaticMarkup(
        <HeatmapChart
          {...heatmapProps}
          rows={[{ id: "short", label: "Short", cells: [{ value: 1 }] }]}
        />,
      ),
    ).toThrow(/has 1 cells; expected 4/u);

    expect(() =>
      renderToStaticMarkup(
        <HeatmapChart
          {...heatmapProps}
          rows={[{
            id: "invalid",
            label: "Invalid",
            cells: [{ value: 1 }, { value: Number.NaN }, { value: 2 }, { value: 3 }],
          }]}
        />,
      ),
    ).toThrow(/finite numbers or null/u);

    expect(() =>
      renderToStaticMarkup(<HeatmapChart {...heatmapProps} domain={[10, 0]} />),
    ).toThrow(/ascending order/u);
  });

  it("keeps matrix headers and an explicit state when rows are empty", () => {
    render(
      <HeatmapChart
        {...heatmapProps}
        rows={[]}
        emptyMessage="No reading activity yet."
      />,
    );

    expect(screen.getByRole("table", { name: "Daily reading data" })).toBeInTheDocument();
    expect(screen.getByText("No reading activity yet.")).toHaveAttribute("colspan", "5");
  });
});

describe("Phase 3 data-display catalogue contracts", () => {
  it("keeps every copied data-display example self-identifying with imports", () => {
    for (const fixture of phaseThreeDataLayoutFixtures) {
      expect(fixture.code, fixture.id).toContain('from "paper-ui"');
    }
  });

  it("publishes deterministic fixtures and complete valid Stable documents", () => {
    expect(
      validateCatalogRegistry({
        documents: phaseThreeDataLayoutDocuments,
        fixtures: phaseThreeDataLayoutFixtures,
        redirects: [],
      }),
    ).toEqual({ valid: true, issues: [] });

    for (const document of phaseThreeDataLayoutDocuments) {
      expect(document.lifecycle).toBe("Stable");
      expect(Object.keys(document.sections?.required ?? {})).toEqual(
        expect.arrayContaining([...REQUIRED_COMPONENT_SECTION_KEYS]),
      );
      expect(Object.keys(document.sections?.required ?? {})).toHaveLength(16);
    }
    expect(phaseThreeDataLayoutFixtures.every((fixture) => fixture.deterministic)).toBe(
      true,
    );
  });
});
