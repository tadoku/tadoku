import Example44 from "./examples/heatmap-chart.empty";
import Example44Source from "./examples/heatmap-chart.empty.tsx?raw";
import Example43 from "./examples/heatmap-chart.reading-activity";
import Example43Source from "./examples/heatmap-chart.reading-activity.tsx?raw";
import Example42 from "./examples/table.empty";
import Example42Source from "./examples/table.empty.tsx?raw";
import Example41 from "./examples/table.reading-log";
import Example41Source from "./examples/table.reading-log.tsx?raw";
import {
COMPONENT_PAGE_SECTION_KEYS,
defineCatalogDocument,
defineCatalogFixture,
type CatalogDocument,
type ComponentDocumentationSections,
type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-09-05";
const PACKAGE_VERSION = "0.1.0";
export const phaseThreeDataLayoutFixtures = [
  defineCatalogFixture({ ...{
  "id": "table.reading-log",
  "name": "Reading log table",
  "description": "A responsive reading log with native column and row headers.",
  "tags": [
    "table",
    "responsive",
    "row headers",
    "reading log"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example41Source, render: () => <Example41 /> }),
  defineCatalogFixture({ ...{
  "id": "table.empty",
  "name": "Empty reading log",
  "description": "An empty table keeps its caption, headers, and explicit state message.",
  "tags": [
    "table",
    "empty",
    "status"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example42Source, render: () => <Example42 /> }),
  defineCatalogFixture({ ...{
  "id": "heatmap-chart.reading-activity",
  "name": "2023 annual reading activity",
  "description": "A full-year calendar of daily activity using the original compact week-by-week geometry.",
  "tags": [
    "heatmap",
    "chart",
    "calendar",
    "annual activity"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example43Source, render: () => <Example43 /> }),
  defineCatalogFixture({ ...{
  "id": "heatmap-chart.empty",
  "name": "Year without activity",
  "description": "Missing dates use the zero-activity baseline; provide an explicit explanation alongside the chart.",
  "tags": [
    "heatmap",
    "empty"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example44Source, render: () => <Example44 /> })
] as const;

function completeSections(values: RequiredComponentSections): ComponentDocumentationSections { return { required: values, pageSections: COMPONENT_PAGE_SECTION_KEYS }; }
const tableSections = completeSections({
  overview: { heading: "Overview", content: ["Table presents comparable records in native rows and columns while Paper supplies restrained rules, density, and responsive overflow."] },
  whenToUse: { heading: "When to use", content: ["Use Table when readers need to scan or compare several records across the same fields."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use Table for prose, a single record, or layouts whose cells do not have meaningful row and column relationships."] },
  choosingBetween: { heading: "Choosing between", content: ["Use a list for one-dimensional content, Surface for one grouped summary, and HeatmapChart when a compact annual calendar communicates daily activity."] },
  anatomy: { heading: "Anatomy", content: ["A labelled scroll region contains a native caption, column-header row, body rows, optional row headers, and an explicit empty state."] },
  recommendedExample: { heading: "Recommended example", content: ["The reading log identifies each row by title, aligns pages to the end, and uses stable IDs. Switch to Empty reading log to compare the preserved headers and empty message. On a phone, focus the labelled table region and scroll horizontally to inspect all columns."] },
  variants: { heading: "Variants", content: ["Captions may be visible or screen-reader-only; columns may align start, center, or end and one identifying column may become row headers."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Empty data retains the table structure. Narrow viewports preserve native semantics and expose horizontal keyboard scrolling instead of converting cells into ambiguous cards."] },
  behavior: { heading: "Behavior", content: ["The region receives focus for keyboard scrolling; Table does not own sorting, selection, pagination, loading, or data fetching."] },
  contentGuidance: { heading: "Content guidance", content: ["Write concise noun headers, use a caption that names the data set, format values consistently, and state an actionable empty message."] },
  accessibility: { heading: "Accessibility", content: ["Always provide a meaningful caption, mark the identifying column with rowHeader when rows need names, and never communicate state with alignment or color alone."] },
  implementation: { heading: "Implementation", content: ["Pass immutable rows and column definitions. Load paper-ui/styles.css once and keep interactive controls inside cells natively labelled."] },
  apiReference: { heading: "API reference", content: ["TableProps<Row> accepts caption, columns, rows, getRowKey, emptyMessage, captionVisibility, minWidth, and native region attributes; TableColumn<Row> defines header and cell rendering."] },
  relatedPatterns: { heading: "Related patterns", content: ["Related contracts include Surface for summaries, HeatmapChart for magnitude matrices, and ButtonGroup for a small visible set of row actions."] },
  migration: { heading: "Migration", content: ["Replace copied responsive-table wrappers, preserve native table markup, move data operations to the application, and explicitly identify row headers."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0; changes to native semantics, responsive overflow, or generic column typing require compatibility review."] }
});

const heatmapSections = completeSections({
  overview: { heading: "Overview", content: ["HeatmapChart is Tadoku's compact annual activity calendar, with weeks across the horizontal axis and weekdays down the vertical axis."] },
  whenToUse: { heading: "When to use", content: ["Use HeatmapChart to show the rhythm, gaps, and relative intensity of dated activity across one calendar year."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for exact record comparison, multiple unrelated series, negative-versus-positive direction, or trends that need a continuous axis."] },
  choosingBetween: { heading: "Choosing between", content: ["Use Table for exact multi-field records, a line chart for continuous trends, and HeatmapChart for a familiar contribution-calendar overview of daily activity."] },
  anatomy: { heading: "Anatomy", content: ["Sparse weekday labels and month labels frame a full year of 10 by 10 pixel cells separated by 3 pixel gaps. An SVG tooltip layer stays above every cell."] },
  recommendedExample: { heading: "Recommended example", content: ["Hover, touch, or Tab to a populated date to inspect its full date and points. At a narrow width, scroll the chart horizontally. Compare with Year without activity to see why empty data also needs explanatory text."] },
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
      props: [
  {
    "name": "caption",
    "type": "ReactNode",
    "description": "Names the data and its labelled scrolling region.",
    "required": true
  },
  {
    "name": "rows",
    "type": "readonly Row[]",
    "description": "Records to compare; supply an empty array for the empty state.",
    "required": true
  },
  {
    "name": "columns",
    "type": "readonly TableColumn<Row>[]",
    "description": "At least one column, each with id, header and cell(row, index).",
    "required": true
  },
  {
    "name": "columns[].rowHeader",
    "type": "boolean",
    "description": "Marks the cell that identifies the record as a row header.",
    "defaultValue": "false"
  },
  {
    "name": "columns[].align",
    "type": "\"start\" | \"center\" | \"end\"",
    "description": "Align numeric quantities to the end for comparison.",
    "defaultValue": "start"
  },
  {
    "name": "columns[].width",
    "type": "string",
    "description": "Optional CSS width hint; content can still influence native table sizing."
  },
  {
    "name": "getRowKey",
    "type": "(row, index) => Key",
    "description": "Use stable data IDs when records can reorder.",
    "defaultValue": "row index"
  },
  {
    "name": "captionVisibility",
    "type": "\"visible\" | \"screen-reader\"",
    "description": "Hide only if nearby visible context already identifies the data.",
    "defaultValue": "visible"
  },
  {
    "name": "emptyMessage",
    "type": "ReactNode",
    "description": "Specific empty state and useful next action.",
    "defaultValue": "No data to display."
  },
  {
    "name": "minWidth",
    "type": "string",
    "description": "Minimum table width before its own region scrolls; choose for the data.",
    "defaultValue": "max(30, columns.length × 9) rem"
  },
  {
    "name": "tableClassName",
    "type": "string",
    "description": "Optional class on the native table; className applies to the outer region."
  }
],
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
    fixtureIds: ["heatmap-chart.reading-activity", "heatmap-chart.empty"],
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
      props: [
  {
    "name": "id",
    "type": "string",
    "description": "Unique per mounted chart for its tooltip layer.",
    "required": true
  },
  {
    "name": "year",
    "type": "number",
    "description": "Calendar year to display.",
    "required": true
  },
  {
    "name": "data",
    "type": "readonly HeatmapChartDatum[]",
    "description": "ISO date, nonnegative value and optional tooltip for each activity record.",
    "required": true
  },
  {
    "name": "data[].date",
    "type": "string",
    "description": "YYYY-MM-DD calendar date within the year. Missing dates show zero.",
    "required": true
  },
  {
    "name": "data[].value",
    "type": "number",
    "description": "Activity magnitude in a consistent unit. Colors scale to the year’s maximum.",
    "required": true
  },
  {
    "name": "data[].tooltip",
    "type": "string",
    "description": "Full date, quantity and unit. Enables keyboard-focusable cell detail."
  }
],
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
