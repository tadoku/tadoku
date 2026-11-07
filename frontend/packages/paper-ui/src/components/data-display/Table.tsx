import { focusRing } from "../../foundations/recipes";
import { cn, tv } from "../../foundations/variants";
import {
  useId,
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
  /** Classes for this column's header cell, merged over Paper's. */
  readonly headerClassName?: string;
  /** Classes for this column's body cells, merged over Paper's. */
  readonly cellClassName?: string;
}

export interface TableProps<Row>
  extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
  readonly caption: ReactNode;
  readonly captionVisibility?: "visible" | "screen-reader";
  readonly columns: readonly TableColumn<Row>[];
  readonly rows: readonly Row[];
  readonly getRowKey?: (row: Row, rowIndex: number) => Key;
  /** Native row attributes for focus targets, current records and application styling. */
  readonly getRowProps?: (row: Row, rowIndex: number) => HTMLAttributes<HTMLTableRowElement>;
  readonly emptyMessage?: ReactNode;
  /** The table's minimum inline size before its keyboard-scrollable region overflows. */
  readonly minWidth?: string;
  readonly tableClassName?: string;
  /** Let the link in each row's header cover its whole row, so a click anywhere on the row opens the record. */
  readonly linkedRows?: boolean;
}

const table = tv({
  slots: {
    region: [
      "paper-table-region max-w-full overflow-x-auto overscroll-x-contain [background:var(--effect-table-scroll)]",
      "focus:scroll-smooth motion-reduce:focus:scroll-auto",
    ],
    table: "paper-table w-full border-collapse border-y border-rule text-ink tabular-nums",
    caption: [
      "paper-table__caption pb-2.5 text-start font-display text-component leading-component font-bold text-ink",
    ],
    header: [
      "paper-table__header border-b border-rule px-field-x py-field-y text-start align-bottom",
      "text-metadata leading-label font-semibold text-muted",
    ],
    cell: "paper-table__cell border-b border-rule px-field-x py-field-y text-start align-top",
    row: "",
    empty: "paper-table__empty px-field-x py-6 text-center text-muted",
  },
  variants: {
    captionVisibility: { visible: {}, "screen-reader": { caption: "paper-table__caption--screen-reader sr-only" } },
    linkedRows: {
      true: { table: "paper-table--linked-rows", row: "relative hover:bg-neutral-hover" },
    },
  },
});

const alignments = {
  start: "paper-table__cell--start text-start",
  center: "paper-table__cell--center text-center",
  end: "paper-table__cell--end text-end",
} as const;

const rowHeaderClassName = "paper-table__row-header font-semibold";
const linkedRowHeaderClassName = "paper-table__row-header font-semibold [&_a]:after:absolute [&_a]:after:inset-0";

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
  getRowProps,
  emptyMessage = "No data to display.",
  minWidth = `${Math.max(30, columns.length * 9)}rem`,
  tableClassName,
  linkedRows = false,
  className,
  style,
  ...regionProps
}: TableProps<Row>) {
  if (columns.length === 0) {
    throw new Error("Table requires at least one column.");
  }
  const reactId = useId();
  const captionId = `paper-table-caption-${reactId.replace(/:/gu, "")}`;
  const slots = table({ captionVisibility, linkedRows });

  return (
    <div
      {...regionProps}
      className={cn(slots.region(), "paper-focus-ring", focusRing(), className)}
      style={style}
      role="region"
      aria-labelledby={captionId}
      tabIndex={0}
    >
      <table className={slots.table({ className: tableClassName })} style={{ minInlineSize: minWidth }}>
        <caption id={captionId} className={slots.caption()}>
          {caption}
        </caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                scope="col"
                className={slots.header({ className: [alignments[column.align ?? "start"], column.headerClassName] })}
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
              <td className={slots.empty()} colSpan={columns.length}>
                {emptyMessage}
              </td>
            </tr>
          ) : (
            rows.map((row, rowIndex) => {
              const rowProps = getRowProps?.(row, rowIndex);
              return (
              <tr
                {...rowProps}
                className={slots.row({ className: rowProps?.className })}
                key={getRowKey?.(row, rowIndex) ?? rowIndex}
              >
                {columns.map((column) => {
                  const Cell = column.rowHeader ? "th" : "td";
                  return (
                    <Cell
                      key={column.id}
                      {...(column.rowHeader ? { scope: "row" as const } : {})}
                      className={slots.cell({
                        className: [
                          column.rowHeader && (linkedRows ? linkedRowHeaderClassName : rowHeaderClassName),
                          alignments[column.align ?? "start"],
                          column.cellClassName,
                        ],
                      })}
                    >
                      {column.cell(row, rowIndex)}
                    </Cell>
                  );
                })}
              </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
