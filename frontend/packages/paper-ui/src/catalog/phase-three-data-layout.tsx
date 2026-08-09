import { DateTime, Interval } from "luxon";
import { HeatmapChart } from "../components/data-display/HeatmapChart";
import { Table, type TableColumn } from "../components/data-display/Table";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  COMPONENT_PAGE_SECTION_KEYS,
  type CatalogDocument,
  type ComponentDocumentationSections,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const PACKAGE_VERSION = "0.1.0";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

interface ReadingRow {
  readonly id: string;
  readonly title: string;
  readonly language: string;
  readonly progress: number;
  readonly status: string;
}

const readingRows: readonly ReadingRow[] = [
  { id: "1", title: "The Housekeeper and the Professor", language: "Japanese", progress: 184, status: "Finished" },
  { id: "2", title: "Convenience Store Woman", language: "Japanese", progress: 73, status: "Reading" },
  { id: "3", title: "The Three-Body Problem", language: "Chinese", progress: 42, status: "Reading" },
];

const readingColumns: readonly TableColumn<ReadingRow>[] = [
  { id: "title", header: "Title", rowHeader: true, width: "18rem", cell: (row) => row.title },
  { id: "language", header: "Language", cell: (row) => row.language },
  { id: "progress", header: "Pages", align: "end", cell: (row) => row.progress },
  { id: "status", header: "Status", cell: (row) => row.status },
];

const heatmapYear = 2023;
const heatmapData = Interval.fromDateTimes(
  DateTime.fromObject({ year: heatmapYear, month: 1, day: 1 }),
  DateTime.fromObject({ year: heatmapYear, month: 12, day: 31 }).endOf("day"),
).splitBy({ day: 1 }).flatMap((interval, index) => {
  const date = interval.start;
  if (date === null) return [];
  const value = index % 5 === 0 ? 0 : ((index * 37) % 100) + 1;
  return [{
    date: date.toISODate() ?? "",
    value,
    tooltip: `${value} points on ${date.toLocaleString(DateTime.DATE_FULL)}`,
  }];
});

export const phaseThreeDataLayoutFixtures = [
  defineCatalogFixture({
    id: "table.reading-log",
    name: "Reading log table",
    description: "A responsive reading log with native column and row headers.",
    tags: ["table", "responsive", "row headers", "reading log"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Table } from "paper-ui";

<Table
  caption="Recent reading"
  rows={entries}
  getRowKey={(entry) => entry.id}
  columns={[
    { id: "title", header: "Title", rowHeader: true, cell: (entry) => entry.title },
    { id: "language", header: "Language", cell: (entry) => entry.language },
    { id: "pages", header: "Pages", align: "end", cell: (entry) => entry.pages },
  ]}
/>`,
    render: () => (
      <Table
        caption="Recent reading"
        rows={readingRows}
        columns={readingColumns}
        getRowKey={(row) => row.id}
      />
    ),
  }),
  defineCatalogFixture({
    id: "table.empty",
    name: "Empty reading log",
    description: "An empty table keeps its caption, headers, and explicit state message.",
    tags: ["table", "empty", "status"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Table, type TableColumn } from "paper-ui";

const columns = [
  { id: "title", header: "Title", rowHeader: true, cell: (entry) => entry.title },
] satisfies readonly TableColumn<{ title: string }>[];

<Table caption="Recent reading" columns={columns} rows={[]} emptyMessage="No reading logged yet." />`,
    render: () => (
      <Table
        caption="Recent reading"
        rows={[]}
        columns={readingColumns}
        emptyMessage="No reading logged yet."
      />
    ),
  }),
  defineCatalogFixture({
    id: "heatmap-chart.reading-activity",
    name: "2023 annual reading activity",
    description: "A full-year calendar of daily activity using the original compact week-by-week geometry.",
    tags: ["heatmap", "chart", "calendar", "annual activity"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { DateTime, Interval } from "luxon";
import { HeatmapChart } from "paper-ui";

const year = 2023;
const data = Interval.fromDateTimes(
  DateTime.fromObject({ year, month: 1, day: 1 }),
  DateTime.fromObject({ year, month: 12, day: 31 }).endOf("day"),
).splitBy({ day: 1 }).flatMap((interval, index) => {
  const date = interval.start;
  if (date === null) return [];
  const value = index % 5 === 0 ? 0 : ((index * 37) % 100) + 1;
  return [{
    date: date.toISODate(),
    value,
    tooltip: value + " points on " + date.toLocaleString(DateTime.DATE_FULL),
  }];
});

<HeatmapChart id="reading-activity" year={year} data={data} />`,
    render: () => (
      <HeatmapChart
        id="reading-activity"
        year={heatmapYear}
        data={heatmapData}
      />
    ),
  }),
] as const;

function completeSections(
  values: RequiredComponentSections,
): ComponentDocumentationSections {
  return { required: values, pageSections: COMPONENT_PAGE_SECTION_KEYS };
}

const tableSections = completeSections({
  overview: { heading: "Overview", content: ["Table presents comparable records in native rows and columns while Paper supplies restrained rules, density, and responsive overflow."] },
  whenToUse: { heading: "When to use", content: ["Use Table when readers need to scan or compare several records across the same fields."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use Table for prose, a single record, or layouts whose cells do not have meaningful row and column relationships."] },
  choosingBetween: { heading: "Choosing between", content: ["Use a list for one-dimensional content, Surface for one grouped summary, and HeatmapChart when a compact annual calendar communicates daily activity."] },
  anatomy: { heading: "Anatomy", content: ["A labelled scroll region contains a native caption, column-header row, body rows, optional row headers, and an explicit empty state."] },
  recommendedExample: { heading: "Recommended example", content: ["Identify reading-log rows by title, align numeric progress to the end, and provide stable keys from application data."] },
  variants: { heading: "Variants", content: ["Captions may be visible or screen-reader-only; columns may align start, center, or end and one identifying column may become row headers."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Empty data retains the table structure. Narrow viewports preserve native semantics and expose horizontal keyboard scrolling instead of converting cells into ambiguous cards."] },
  behavior: { heading: "Behavior", content: ["The region receives focus for keyboard scrolling; Table does not own sorting, selection, pagination, loading, or data fetching."] },
  contentGuidance: { heading: "Content guidance", content: ["Write concise noun headers, use a caption that names the data set, format values consistently, and state an actionable empty message."] },
  accessibility: { heading: "Accessibility", content: ["Always provide a meaningful caption, mark the identifying column with rowHeader when rows need names, and never communicate state with alignment or color alone."] },
  implementation: { heading: "Implementation", content: ["Pass immutable rows and column definitions. Load paper-ui/styles.css once and keep interactive controls inside cells natively labelled."] },
  apiReference: { heading: "API reference", content: ["TableProps<Row> accepts caption, columns, rows, getRowKey, emptyMessage, captionVisibility, minWidth, and native region attributes; TableColumn<Row> defines header and cell rendering."] },
  relatedPatterns: { heading: "Related patterns", content: ["Related contracts include Surface for summaries, HeatmapChart for magnitude matrices, and ButtonGroup for a small visible set of row actions."] },
  migration: { heading: "Migration", content: ["Replace copied responsive-table wrappers, preserve native table markup, move data operations to the application, and explicitly identify row headers."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0; changes to native semantics, responsive overflow, or generic column typing require compatibility review."] },
});

const heatmapSections = completeSections({
  overview: { heading: "Overview", content: ["HeatmapChart is Tadoku's compact annual activity calendar, with weeks across the horizontal axis and weekdays down the vertical axis."] },
  whenToUse: { heading: "When to use", content: ["Use HeatmapChart to show the rhythm, gaps, and relative intensity of dated activity across one calendar year."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for exact record comparison, multiple unrelated series, negative-versus-positive direction, or trends that need a continuous axis."] },
  choosingBetween: { heading: "Choosing between", content: ["Use Table for exact multi-field records, a line chart for continuous trends, and HeatmapChart for a familiar contribution-calendar overview of daily activity."] },
  anatomy: { heading: "Anatomy", content: ["Sparse weekday labels and month labels frame a full year of 10 by 10 pixel cells separated by 3 pixel gaps. An SVG tooltip layer stays above every cell."] },
  recommendedExample: { heading: "Recommended example", content: ["Pass the reporting year and one dated record per day. Include concise tooltip copy with the value, unit, and full date."] },
  variants: { heading: "Variants", content: ["The calendar geometry is intentionally fixed. Zero activity uses the neutral cell color, while positive values occupy four theme-aware intensity bands relative to the year's maximum."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Days without a record count as zero. Cells outside the requested year remain transparent, and an all-zero year stays on the neutral baseline."] },
  behavior: { heading: "Behavior", content: ["HeatmapChart maps ISO dates into week and weekday positions, uses the last repeated date value, scales positive activity against the annual maximum, and reveals supplied copy on hover, touch, or keyboard focus."] },
  contentGuidance: { heading: "Content guidance", content: ["Use ISO calendar dates and consistent numeric units. Write tooltips as short standalone descriptions, such as 42 points on January 5, 2023."] },
  accessibility: { heading: "Accessibility", content: ["The SVG has a year-specific accessible name, dated cells expose their date and value, and cells with tooltip copy can receive keyboard focus. Forced-colors mode outlines the cells."] },
  implementation: { heading: "Implementation", content: ["Pass the original id, year, and data contract and load paper-ui/styles.css once. Keep ids unique when more than one heatmap appears on a page."] },
  apiReference: { heading: "API reference", content: ["HeatmapChartProps accepts id, year, and data. Each HeatmapChartDatum contains an ISO date, numeric value, and optional tooltip string."] },
  relatedPatterns: { heading: "Related patterns", content: ["Related contracts include Table for exact record comparison and the Color foundation for the semantic chart and forced-color tokens used by the calendar."] },
  migration: { heading: "Migration", content: ["Replace the incompatible Paper matrix props with the original ui HeatmapChart contract unchanged: id, year, and dated data records."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0; calendar geometry, five-level intensity mapping, and tooltip placement preserve the original ui component contract."] },
});

function componentDocument(
  options: Pick<
    CatalogDocument,
    | "id"
    | "route"
    | "name"
    | "summary"
    | "keywords"
    | "sourcePath"
    | "guidance"
    | "accessibility"
    | "api"
    | "fixtureIds"
    | "migration"
    | "behaviorTestIds"
    | "sections"
  >,
): CatalogDocument {
  return defineCatalogDocument({
    ...options,
    kind: "component",
    category: "data-display",
    aliases: [],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    packageVersion: PACKAGE_VERSION,
    dependencies: { documents: [], packages: ["paper-ui", "react"] },
    changelog: [{ date: REVIEW_DATE, note: "Published the Stable Phase 3 contract." }],
  });
}

export const phaseThreeDataLayoutDocuments = [
  componentDocument({
    id: "component.table",
    route: "/components/data-display/table",
    name: "Table",
    summary: "Presents comparable records with native headers and keyboard-scrollable responsive behavior.",
    keywords: ["table", "data", "responsive", "row header"],
    sourcePath: "src/components/data-display/Table.tsx",
    fixtureIds: ["table.reading-log", "table.empty"],
    behaviorTestIds: ["table.native-semantics", "table.responsive-region", "table.empty-state"],
    guidance: {
      whenToUse: ["Compare records that share a stable set of fields."],
      whenNotToUse: ["Do not use a table only to position unrelated content."],
      content: ["Use concise headers, consistent value formats, and a specific caption."],
      commonMistakes: ["Do not replace native table semantics with a grid of generic elements."],
    },
    accessibility: {
      requirements: ["Provide a caption and mark identifying cells as row headers."],
      keyboard: ["Tab focuses the overflow region; arrow keys scroll it when content is wider than the viewport."],
      knownConstraints: ["Application-owned interactive cell controls need their own accessible names."],
    },
    api: {
      react: ["Table"],
      cssClasses: ["paper-table-region", "paper-table", "paper-table__*"],
      publicTypes: ["TableProps<Row>", "TableColumn<Row>", "TableColumnAlignment"],
      defaults: ["Caption is visible, cells align to start, and minWidth derives from the column count."],
      invalidCombinations: ["At least one column is required; a rowHeader column should identify rather than merely describe a row."],
    },
    migration: {
      legacy: ["ui/components/Table", "application-owned responsive table wrappers"],
      notes: ["Preserve native headers and move sorting, selection, and pagination to application composition."],
    },
    sections: tableSections,
  }),
  componentDocument({
    id: "component.heatmap-chart",
    route: "/components/data-display/heatmap-chart",
    name: "HeatmapChart",
    summary: "Shows a full year of dated activity in the original compact week-by-week calendar.",
    keywords: ["heatmap", "chart", "calendar", "activity", "year"],
    sourcePath: "src/components/data-display/HeatmapChart.tsx",
    fixtureIds: ["heatmap-chart.reading-activity"],
    behaviorTestIds: ["heatmap.calendar-geometry", "heatmap.intensity-bands", "heatmap.tooltip-interaction"],
    guidance: {
      whenToUse: ["Show the rhythm and intensity of daily activity across one calendar year."],
      whenNotToUse: ["Do not use it when exact values or continuous trends matter more than the annual pattern."],
      content: ["Include the value, unit, and full date in each supplied tooltip."],
      commonMistakes: ["Do not reshape the dated records into generic rows and columns."],
    },
    accessibility: {
      requirements: ["Provide tooltip text for meaningful activity so its exact date and value are exposed alongside the visual intensity."],
      keyboard: ["Tab reaches cells with tooltip content; focus reveals the same tooltip shown by hover and touch."],
      knownConstraints: ["The compact annual overview supplements rather than replaces an exact activity log."],
    },
    api: {
      react: ["HeatmapChart"],
      cssClasses: ["paper-heatmap", "paper-heatmap__*"],
      publicTypes: ["HeatmapChartProps", "HeatmapChartDatum"],
      defaults: ["Missing dates have value zero; positive values use four bands relative to the maximum value in the requested year."],
      invalidCombinations: ["Ids must be unique per page, dates use ISO calendar strings, and consumers should keep records within the requested year."],
    },
    migration: {
      legacy: ["ui/components/HeatmapChart", "the temporary Paper generic matrix HeatmapChart"],
      notes: ["Keep the original id, year, and data props; remove temporary columns, rows, domain, and formatting props."],
    },
    sections: heatmapSections,
  }),
] as const;
