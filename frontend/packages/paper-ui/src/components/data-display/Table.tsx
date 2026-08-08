import {
  useId,
  type CSSProperties,
  type HTMLAttributes,
  type Key,
  type ReactNode,
} from "react";

export type TableColumnAlignment = "start" | "center" | "end";

export interface TableColumn<Row> {
  /** Stable identifier used for React keys and CSS hooks. */
  readonly id: string;
  readonly header: ReactNode;
  readonly cell: (row: Row, rowIndex: number) => ReactNode;
  readonly align?: TableColumnAlignment;
  /** Render this column's body cells as row headers. */
  readonly rowHeader?: boolean;
  /** Optional CSS width or min-width hint, for example `8rem` or `20%`. */
  readonly width?: string;
}

export interface TableProps<Row>
  extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
  readonly caption: ReactNode;
  readonly captionVisibility?: "visible" | "screen-reader";
  readonly columns: readonly TableColumn<Row>[];
  readonly rows: readonly Row[];
  readonly getRowKey?: (row: Row, rowIndex: number) => Key;
  readonly emptyMessage?: ReactNode;
  /** The table's minimum inline size before its keyboard-scrollable region overflows. */
  readonly minWidth?: string;
  readonly tableClassName?: string;
}

type TableStyle = CSSProperties & {
  readonly "--paper-table-min-inline-size": string;
};

function classNames(...values: readonly (string | false | undefined)[]): string {
  return values.filter(Boolean).join(" ");
}

/**
 * A native data table in a labelled, keyboard-scrollable responsive region.
 * Use `rowHeader` on the column that identifies each record.
 */
export function Table<Row>({
  caption,
  captionVisibility = "visible",
  columns,
  rows,
  getRowKey,
  emptyMessage = "No data to display.",
  minWidth = `${Math.max(30, columns.length * 9)}rem`,
  tableClassName,
  className,
  style,
  ...regionProps
}: TableProps<Row>) {
  if (columns.length === 0) {
    throw new Error("Table requires at least one column.");
  }
  const reactId = useId();
  const captionId = `paper-table-caption-${reactId.replace(/:/gu, "")}`;
  const regionStyle: TableStyle = {
    ...style,
    "--paper-table-min-inline-size": minWidth,
  };

  return (
    <div
      {...regionProps}
      className={classNames("paper-table-region", "paper-focus-ring", className)}
      style={regionStyle}
      role="region"
      aria-labelledby={captionId}
      tabIndex={0}
    >
      <table className={classNames("paper-table", tableClassName)}>
        <caption
          id={captionId}
          className={classNames(
            "paper-table__caption",
            captionVisibility === "screen-reader" && "paper-table__caption--screen-reader",
          )}
        >
          {caption}
        </caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                scope="col"
                className={classNames(
                  "paper-table__header",
                  `paper-table__cell--${column.align ?? "start"}`,
                )}
                style={column.width ? { width: column.width } : undefined}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td className="paper-table__empty" colSpan={columns.length}>
                {emptyMessage}
              </td>
            </tr>
          ) : (
            rows.map((row, rowIndex) => (
              <tr key={getRowKey?.(row, rowIndex) ?? rowIndex}>
                {columns.map((column) => {
                  const Cell = column.rowHeader ? "th" : "td";
                  return (
                    <Cell
                      key={column.id}
                      {...(column.rowHeader ? { scope: "row" as const } : {})}
                      className={classNames(
                        "paper-table__cell",
                        column.rowHeader && "paper-table__row-header",
                        `paper-table__cell--${column.align ?? "start"}`,
                      )}
                    >
                      {column.cell(row, rowIndex)}
                    </Cell>
                  );
                })}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
