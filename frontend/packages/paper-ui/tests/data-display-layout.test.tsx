import { fireEvent, render, screen, within } from "@testing-library/react";
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
  id: "reading-activity-2023",
  year: 2023,
  data: [
    { date: "2023-01-01", value: 0, tooltip: "0 points on January 1, 2023" },
    { date: "2023-01-02", value: 10, tooltip: "10 points on January 2, 2023" },
    { date: "2023-01-03", value: 30, tooltip: "30 points on January 3, 2023" },
    { date: "2023-01-04", value: 60, tooltip: "60 points on January 4, 2023" },
    { date: "2023-01-05", value: 100, tooltip: "100 points on January 5, 2023" },
  ],
};

function calendarCells(svg: SVGSVGElement): readonly SVGRectElement[] {
  return Array.from(
    svg.querySelectorAll<SVGRectElement>('rect[width="10"][height="10"]'),
  );
}

function intensityLevel(cell: SVGRectElement): string | undefined {
  const explicitLevel = cell.getAttribute("data-level");
  if (explicitLevel) return explicitLevel;

  const legacyClasses = [
    "fill-stone-200",
    "fill-teal-300",
    "fill-teal-500",
    "fill-teal-700",
    "fill-teal-900",
  ];
  const legacyLevel = legacyClasses.findIndex((className) =>
    cell.classList.contains(className),
  );
  return legacyLevel === -1 ? undefined : String(legacyLevel);
}

function tooltipIsHidden(tooltip: SVGGElement): boolean {
  return tooltip.classList.contains("hidden") ||
    tooltip.hasAttribute("hidden") ||
    tooltip.getAttribute("aria-hidden") === "true";
}

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
  it("renders the requested year as the original month-by-weekday SVG calendar", () => {
    const { container } = render(<HeatmapChart {...heatmapProps} />);

    const svg = container.querySelector("svg");
    expect(svg).toBeInstanceOf(SVGSVGElement);
    const labels = Array.from(svg!.querySelectorAll("text"), (label) => label.textContent);
    expect(labels).toEqual(expect.arrayContaining([
      "Jan", "Feb", "Mar", "Apr", "May", "Jun",
      "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
    ]));
    expect(labels).toEqual(expect.arrayContaining(["Mon", "Wed", "Fri", "Sun"]));

    const cells = calendarCells(svg!);
    expect(cells).toHaveLength(53 * 7);
    expect(cells[0]).toHaveAttribute("x", "30");
    expect(cells[0]).toHaveAttribute("y", "15");
    expect(cells[1]).toHaveAttribute("y", "28");
    expect(cells[7]).toHaveAttribute("x", "43");
    expect(cells.every((cell) =>
      cell.getAttribute("width") === "10" && cell.getAttribute("height") === "10"
    )).toBe(true);
  });

  it("maps dated activity onto five intensity levels without changing cell geometry", () => {
    const { container } = render(<HeatmapChart {...heatmapProps} />);
    const svg = container.querySelector("svg")!;
    const cells = calendarCells(svg);

    // 2023 starts on Sunday. January 1 occupies the final row of the first
    // week, followed by January 2-5 in the first four rows of the next week.
    expect(intensityLevel(cells[6])).toBe("0");
    expect(intensityLevel(cells[7])).toBe("1");
    expect(intensityLevel(cells[8])).toBe("2");
    expect(intensityLevel(cells[9])).toBe("3");
    expect(intensityLevel(cells[10])).toBe("4");
  });

  it("reveals the supplied activity copy while its dated cell is hovered", () => {
    const { container } = render(<HeatmapChart {...heatmapProps} />);
    const svg = container.querySelector("svg")!;
    const januaryFifth = calendarCells(svg)[10];
    const tooltipText = screen.getByText("100 points on January 5, 2023");
    const tooltip = tooltipText.closest("g");

    expect(tooltip).toBeInstanceOf(SVGGElement);
    expect(tooltipIsHidden(tooltip!)).toBe(true);
    fireEvent.mouseOver(januaryFifth);
    expect(tooltipIsHidden(tooltip!)).toBe(false);
    fireEvent.mouseOut(januaryFifth);
    expect(tooltipIsHidden(tooltip!)).toBe(true);
  });
});

describe("Phase 3 data-display catalogue contracts", () => {
  it("keeps every copied data-display example self-identifying with imports", () => {
    for (const fixture of phaseThreeDataLayoutFixtures) {
      expect(fixture.code, fixture.id).toContain('from "paper-ui"');
    }
  });

  it("documents HeatmapChart with an original-style full-year activity fixture", () => {
    const fixture = phaseThreeDataLayoutFixtures.find(
      (candidate) => candidate.id === "heatmap-chart.reading-activity",
    );
    expect(fixture).toBeDefined();
    expect(fixture!.name).toMatch(/year|annual|2023/iu);
    expect(fixture!.description).toMatch(/year|calendar|daily activity/iu);
    expect(fixture!.code).toContain("year={year}");
    expect(fixture!.code).toContain("data={data}");
    expect(fixture!.code).not.toContain("columns=");
    expect(fixture!.code).not.toContain("rows=");

    const { container } = render(fixture!.render());
    const svg = container.querySelector("svg");
    expect(svg).toBeInstanceOf(SVGSVGElement);
    expect(calendarCells(svg!)).toHaveLength(53 * 7);
    const labels = Array.from(svg!.querySelectorAll("text"), (label) => label.textContent);
    expect(labels).toEqual(expect.arrayContaining(["Jan", "Dec"]));
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
