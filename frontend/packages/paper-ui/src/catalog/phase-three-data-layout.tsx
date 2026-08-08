import { HeatmapChart } from "../components/data-display/HeatmapChart";
import { Table, type TableColumn } from "../components/data-display/Table";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type ComponentDocumentationSections,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const PACKAGE_VERSION = "0.1.0";
const OWNER = "Tadoku design systems";
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

const heatmapColumns = [
  { id: "mon", label: "Monday", shortLabel: "Mon" },
  { id: "tue", label: "Tuesday", shortLabel: "Tue" },
  { id: "wed", label: "Wednesday", shortLabel: "Wed" },
  { id: "thu", label: "Thursday", shortLabel: "Thu" },
  { id: "fri", label: "Friday", shortLabel: "Fri" },
  { id: "sat", label: "Saturday", shortLabel: "Sat" },
  { id: "sun", label: "Sunday", shortLabel: "Sun" },
] as const;

const heatmapRows = [
  { id: "week-1", label: "Jul 27", cells: [{ value: 18 }, { value: 24 }, { value: 0 }, { value: 31 }, { value: 12 }, { value: 46 }, { value: 38 }] },
  { id: "week-2", label: "Aug 3", cells: [{ value: 22 }, { value: null }, { value: 17 }, { value: 52 }, { value: 33 }, { value: 64 }, { value: 41 }] },
] as const;

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
    name: "Two weeks of reading activity",
    description: "Daily pages remain printed inside theme-aware intensity cells.",
    tags: ["heatmap", "chart", "non-color", "responsive"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { HeatmapChart } from "paper-ui";

<HeatmapChart
  title="Daily reading activity"
  description="Pages read during the last two weeks."
  columns={days}
  rows={weeks}
  domain={[0, 70]}
  formatValue={(value) => value === null ? "No data" : String(value) + " pages"}
/>`,
    render: () => (
      <HeatmapChart
        title="Daily reading activity"
        description="Pages read during the last two weeks."
        columns={heatmapColumns}
        rows={heatmapRows}
        domain={[0, 70]}
        formatValue={(value) => (value === null ? "No data" : `${value} pages`)}
      />
    ),
  }),
] as const;

function completeSections(
  values: RequiredComponentSections,
): ComponentDocumentationSections {
  return { required: values };
}

const tableSections = completeSections({
  overview: { heading: "Overview", content: ["Table presents comparable records in native rows and columns while Paper supplies restrained rules, density, and responsive overflow."] },
  whenToUse: { heading: "When to use", content: ["Use Table when readers need to scan or compare several records across the same fields."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use Table for prose, a single record, or layouts whose cells do not have meaningful row and column relationships."] },
  choosingBetween: { heading: "Choosing between", content: ["Use a list for one-dimensional content, Surface for one grouped summary, and HeatmapChart when a dense matrix communicates magnitude."] },
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
  overview: { heading: "Overview", content: ["HeatmapChart compares magnitude across a row-and-column matrix using Paper chart tokens, printed values, and native table semantics."] },
  whenToUse: { heading: "When to use", content: ["Use HeatmapChart to reveal clusters, gaps, and relative intensity across two categorical dimensions such as week and weekday."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for precise trend lines, unrelated categories, negative-versus-positive direction, or data that cannot fit a coherent matrix."] },
  choosingBetween: { heading: "Choosing between", content: ["Use Table for exact multi-field records, a line chart for continuous trends, and HeatmapChart for dense magnitude comparison across two dimensions."] },
  anatomy: { heading: "Anatomy", content: ["A title and description precede a low-to-high legend and a scrollable native table with column headers, row headers, intensity cells, and printed values."] },
  recommendedExample: { heading: "Recommended example", content: ["Show daily reading totals by week, use a fixed domain for comparable previews, abbreviate visible weekday labels without shortening their accessible names, and print units in cells."] },
  variants: { heading: "Variants", content: ["Choose one of eight theme-aware chart categories with colorIndex and customize legend endpoints, formatting, domain, and responsive minimum width."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["An empty matrix keeps its headers and state message. Null values render as No data with a distinct pattern, values clamp to the supplied scale, constant domains use the middle intensity, and narrow matrices scroll with sticky row headers."] },
  behavior: { heading: "Behavior", content: ["HeatmapChart validates matrix dimensions and finite values, derives a domain when none is supplied, and leaves filtering, selection, animation, and fetching to the application."] },
  contentGuidance: { heading: "Content guidance", content: ["Name both dimensions, include units in formatted values, distinguish zero from missing data, and use legend labels that describe magnitude rather than judgment."] },
  accessibility: { heading: "Accessibility", content: ["Every value remains visible in a native data cell associated with row and column headers; the scroll region is keyboard focusable and forced-colors mode removes decorative fills."] },
  implementation: { heading: "Implementation", content: ["Supply one cell per column in every row, use null only for missing observations, load Paper styles once, and hold domains stable when users compare multiple heatmaps."] },
  apiReference: { heading: "API reference", content: ["HeatmapChartProps accepts title, description, columns, rows, domain, formatValue, lowLabel, highLabel, colorIndex, minWidth, emptyMessage, and native figure attributes."] },
  relatedPatterns: { heading: "Related patterns", content: ["Related contracts include chartSeries for category cues, Table for record comparison, and the Color foundation for theme-aware visualization tokens."] },
  migration: { heading: "Migration", content: ["Replace canvas-only heatmaps with explicit row, column, and cell data; preserve exact values, convert missing sentinels to null, and remove hard-coded visualization colors."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0; changes to matrix validation, intensity mapping, native table semantics, or chart-token selection require compatibility review."] },
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
    owner: OWNER,
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
    summary: "Shows matrix magnitude with chart-token intensity, printed values, and native table semantics.",
    keywords: ["heatmap", "chart", "matrix", "non-color"],
    sourcePath: "src/components/data-display/HeatmapChart.tsx",
    fixtureIds: ["heatmap-chart.reading-activity"],
    behaviorTestIds: ["heatmap.native-semantics", "heatmap.non-color-values", "heatmap.matrix-validation"],
    guidance: {
      whenToUse: ["Compare magnitude across two categorical dimensions."],
      whenNotToUse: ["Do not use it when direction or exact trends matter more than clusters."],
      content: ["Name dimensions and include units in the printed cell values."],
      commonMistakes: ["Do not rely on hue or intensity as the only way to read a cell."],
    },
    accessibility: {
      requirements: ["Keep values visible and associated with native row and column headers."],
      keyboard: ["Tab focuses the overflow region; arrow keys scroll matrices wider than the viewport."],
      knownConstraints: ["Very large matrices should be summarized or divided before rendering."],
    },
    api: {
      react: ["HeatmapChart"],
      cssClasses: ["paper-heatmap", "paper-heatmap__*"],
      publicTypes: ["HeatmapChartProps", "HeatmapChartColumn", "HeatmapChartRow", "HeatmapChartCell", "HeatmapChartValueContext"],
      defaults: ["Domain is inferred, colorIndex is 0, null reads No data, an empty chart has an explicit message, and numeric values use deterministic string formatting."],
      invalidCombinations: ["Columns cannot be empty; every row needs one cell per column; values and domain endpoints must be finite."],
    },
    migration: {
      legacy: ["canvas-only heatmaps", "hard-coded chart palette classes"],
      notes: ["Convert missing values to null and retain row, column, value, and unit text."],
    },
    sections: heatmapSections,
  }),
] as const;
