import {
  useId,
  type CSSProperties,
  type HTMLAttributes,
  type ReactNode,
} from "react";
import { chartCategoryStyles } from "../../foundations/chart-palette";

export interface HeatmapChartColumn {
  readonly id: string;
  readonly label: string;
  /** Optional compact visible label; the full label remains accessible. */
  readonly shortLabel?: string;
}

export interface HeatmapChartCell {
  readonly value: number | null;
  /** Optional semantic display label, such as `24 pages`. */
  readonly label?: string;
}

export interface HeatmapChartRow {
  readonly id: string;
  readonly label: string;
  readonly cells: readonly HeatmapChartCell[];
}

export interface HeatmapChartValueContext {
  readonly row: HeatmapChartRow;
  readonly column: HeatmapChartColumn;
  readonly rowIndex: number;
  readonly columnIndex: number;
}

export interface HeatmapChartProps
  extends Omit<HTMLAttributes<HTMLElement>, "children" | "title"> {
  readonly title: string;
  readonly description?: ReactNode;
  readonly columns: readonly HeatmapChartColumn[];
  readonly rows: readonly HeatmapChartRow[];
  readonly domain?: readonly [minimum: number, maximum: number];
  readonly formatValue?: (
    value: number | null,
    context: HeatmapChartValueContext,
  ) => ReactNode;
  readonly lowLabel?: string;
  readonly highLabel?: string;
  /** Selects one of Paper's eight theme-aware chart categories. */
  readonly colorIndex?: number;
  readonly minWidth?: string;
  readonly emptyMessage?: ReactNode;
}

type HeatmapStyle = CSSProperties & {
  readonly "--paper-heatmap-color": string;
  readonly "--paper-heatmap-min-inline-size": string;
};

function classNames(...values: readonly (string | undefined)[]): string {
  return values.filter(Boolean).join(" ");
}

function assertData(
  columns: readonly HeatmapChartColumn[],
  rows: readonly HeatmapChartRow[],
  domain: readonly [number, number] | undefined,
): void {
  if (columns.length === 0) {
    throw new Error("HeatmapChart requires at least one column.");
  }
  for (const row of rows) {
    if (row.cells.length !== columns.length) {
      throw new Error(
        `HeatmapChart row "${row.id}" has ${row.cells.length} cells; expected ${columns.length}.`,
      );
    }
    for (const cell of row.cells) {
      if (cell.value !== null && !Number.isFinite(cell.value)) {
        throw new Error("HeatmapChart cell values must be finite numbers or null.");
      }
    }
  }
  if (
    domain &&
    (!Number.isFinite(domain[0]) ||
      !Number.isFinite(domain[1]) ||
      domain[0] > domain[1])
  ) {
    throw new Error("HeatmapChart domain must be two finite numbers in ascending order.");
  }
}

function inferredDomain(
  rows: readonly HeatmapChartRow[],
): readonly [number, number] {
  const values = rows.flatMap((row) =>
    row.cells.flatMap((cell) => (cell.value === null ? [] : [cell.value])),
  );
  if (values.length === 0) return [0, 0];
  return [Math.min(...values), Math.max(...values)];
}

function intensityLevel(
  value: number | null,
  domain: readonly [number, number],
): 0 | 1 | 2 | 3 | 4 | "missing" {
  if (value === null) return "missing";
  const [minimum, maximum] = domain;
  if (minimum === maximum) return 2;
  const normalized = Math.min(1, Math.max(0, (value - minimum) / (maximum - minimum)));
  return Math.round(normalized * 4) as 0 | 1 | 2 | 3 | 4;
}

/**
 * A theme-aware heatmap rendered as a native table. Every cell keeps its value
 * visible, so intensity never relies on color alone.
 */
export function HeatmapChart({
  title,
  description,
  columns,
  rows,
  domain,
  formatValue = (value) => (value === null ? "No data" : String(value)),
  lowLabel = "Lower",
  highLabel = "Higher",
  colorIndex = 0,
  minWidth = `${Math.max(30, 9 + columns.length * 4)}rem`,
  emptyMessage = "No data to display.",
  className,
  style,
  ...figureProps
}: HeatmapChartProps) {
  assertData(columns, rows, domain);
  const reactId = useId();
  const titleId = `paper-heatmap-title-${reactId.replace(/:/gu, "")}`;
  const descriptionId = description
    ? `paper-heatmap-description-${reactId.replace(/:/gu, "")}`
    : undefined;
  const resolvedDomain = domain ?? inferredDomain(rows);
  const safeColorIndex = Number.isFinite(colorIndex) ? Math.trunc(colorIndex) : 0;
  const category = chartCategoryStyles[
    Math.min(chartCategoryStyles.length - 1, Math.max(0, safeColorIndex))
  ];
  const heatmapStyle: HeatmapStyle = {
    ...style,
    "--paper-heatmap-color": category.color,
    "--paper-heatmap-min-inline-size": minWidth,
  };

  return (
    <figure
      {...figureProps}
      className={classNames("paper-heatmap", className)}
      style={heatmapStyle}
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
    >
      <figcaption className="paper-heatmap__caption">
        <strong id={titleId} className="paper-heatmap__title">
          {title}
        </strong>
        {description ? (
          <span id={descriptionId} className="paper-heatmap__description">
            {description}
          </span>
        ) : null}
      </figcaption>
      <div className="paper-heatmap__legend" aria-hidden="true">
        <span>{lowLabel}</span>
        {[0, 1, 2, 3, 4].map((level) => (
          <span
            key={level}
            className={classNames(
              "paper-heatmap__legend-step",
              `paper-heatmap--pattern-${category.pattern}`,
            )}
            data-level={level}
          />
        ))}
        <span>{highLabel}</span>
      </div>
      <div
        className="paper-heatmap__region paper-focus-ring"
        role="region"
        aria-label={`${title} data table`}
        tabIndex={0}
      >
        <table className="paper-heatmap__table">
          <caption className="paper-heatmap__table-caption">{title} data</caption>
          <thead>
            <tr>
              <th scope="col" className="paper-heatmap__corner">
                <span className="paper-sr-only">Series</span>
              </th>
              {columns.map((column) => (
                <th key={column.id} scope="col" className="paper-heatmap__column-header">
                  {column.shortLabel ? (
                    <abbr title={column.label} aria-label={column.label}>
                      {column.shortLabel}
                    </abbr>
                  ) : (
                    column.label
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td className="paper-heatmap__empty" colSpan={columns.length + 1}>
                  {emptyMessage}
                </td>
              </tr>
            ) : rows.map((row, rowIndex) => (
              <tr key={row.id}>
                <th scope="row" className="paper-heatmap__row-header">
                  {row.label}
                </th>
                {row.cells.map((cell, columnIndex) => {
                  const column = columns[columnIndex];
                  const level = intensityLevel(cell.value, resolvedDomain);
                  const formatted =
                    cell.label ??
                    formatValue(cell.value, { row, column, rowIndex, columnIndex });
                  return (
                    <td
                      key={column.id}
                      className={classNames(
                        "paper-heatmap__cell",
                        `paper-heatmap--pattern-${category.pattern}`,
                      )}
                      data-level={level}
                    >
                      <span className="paper-heatmap__value">{formatted}</span>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </figure>
  );
}
