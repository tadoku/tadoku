import * as stylex from "@stylexjs/stylex";
import {
  useId,
  type HTMLAttributes,
  type Key,
  type ReactNode,
} from "react";
import { sx, type StyleXProp } from "../../foundations/props";
import { a11y } from "../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  font,
  fontWeight,
  ratio,
  space,
  type,
} from "../../tokens/tokens.stylex";

const styles = stylex.create({
  /*
   * Edge shades appear only while columns sit off-screen: covers in the surface colour scroll with the table and hide
   * the fixed shades at whichever end is fully visible.
   */
  region: {
    background: `linear-gradient(to right, ${color.surfacePaper} ${ratio.scrollCover}, transparent) left / ${space[32]} 100% no-repeat local, linear-gradient(to left, ${color.surfacePaper} ${ratio.scrollCover}, transparent) right / ${space[32]} 100% no-repeat local, radial-gradient(farthest-side at 0 50%, ${color.scrollShade}, transparent) left / ${space[12]} 100% no-repeat scroll, radial-gradient(farthest-side at 100% 50%, ${color.scrollShade}, transparent) right / ${space[12]} 100% no-repeat scroll`,
    maxInlineSize: "100%",
    overscrollBehaviorInline: "contain",
    scrollBehavior: { default: null, ":focus": { default: "smooth", [breakpoints.reducedMotion]: "auto" } },
    overflowX: "auto",
  },
  canvas: { background: `linear-gradient(to right, ${color.surfaceCanvas} ${ratio.scrollCover}, transparent) left / ${space[32]} 100% no-repeat local, linear-gradient(to left, ${color.surfaceCanvas} ${ratio.scrollCover}, transparent) right / ${space[32]} 100% no-repeat local, radial-gradient(farthest-side at 0 50%, ${color.scrollShade}, transparent) left / ${space[12]} 100% no-repeat scroll, radial-gradient(farthest-side at 100% 50%, ${color.scrollShade}, transparent) right / ${space[12]} 100% no-repeat scroll` },
  raised: { background: `linear-gradient(to right, ${color.surfaceRaised} ${ratio.scrollCover}, transparent) left / ${space[32]} 100% no-repeat local, linear-gradient(to left, ${color.surfaceRaised} ${ratio.scrollCover}, transparent) right / ${space[32]} 100% no-repeat local, radial-gradient(farthest-side at 0 50%, ${color.scrollShade}, transparent) left / ${space[12]} 100% no-repeat scroll, radial-gradient(farthest-side at 100% 50%, ${color.scrollShade}, transparent) right / ${space[12]} 100% no-repeat scroll` },
  paper: {},
  table: {
    borderBlockStyle: "solid",
    borderBlockColor: color.ruleDefault,
    borderBlockWidth: border.static,
    borderCollapse: "collapse",
    color: color.textInk,
    fontVariantNumeric: "tabular-nums",
    inlineSize: "100%",
  },
  minWidth: (value: string) => ({ minInlineSize: value }),
  sheet: { borderBlockWidth: 0 },
  caption: {
    color: color.textInk,
    fontFamily: font.display,
    fontSize: type.componentSize,
    fontWeight: fontWeight.bold,
    lineHeight: type.componentLineHeight,
    paddingBlockEnd: space[10],
    textAlign: "start",
  },
  cell: {    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
 boxSizing: "border-box",
    textAlign: "start",
    verticalAlign: "top",
  },
  /** In a sheet the last row meets the sheet's own edge. */
  lastRowCell: { borderBlockEndWidth: 0 },
  /** Headers sit on the row's baseline edge, so a wrapped header doesn't leave its neighbours floating. */
  header: {
    color: color.textMuted,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.semibold,
    lineHeight: type.labelLineHeight,
    verticalAlign: "bottom",
  },
  rowHeader: { fontWeight: fontWeight.semibold },
  start: {},
  center: { textAlign: "center" },
  end: { textAlign: "end" },
  width: (value: string) => ({ width: value }),
  empty: {
    paddingBlock: space[24],
    paddingInline: density.fieldPaddingInline,
    color: color.textMuted,
    textAlign: "center",
  },
  row: {
    backgroundColor: {
      '[aria-current="true"]': { default: color.actionSoft, ":hover": color.currentRowHover },
      default: null,
    },
  },
  linkedRow: {
    backgroundColor: {
      '[aria-current="true"]': { default: color.actionSoft, ":hover": color.currentRowHover },
      default: null,
      ":hover": color.actionNeutralHover,
    },
    position: "relative",
  },
  /** The row header's link covers its row; it stays the row's only tab stop and accessible name. */
  rowLink: { "::after": { inset: 0, content: '""', position: "absolute", } },
});

function TableRow({ linked, className, ...props }: HTMLAttributes<HTMLTableRowElement> & { readonly linked: boolean }) {
  return <tr {...props} {...sx(className, linked ? styles.linkedRow : styles.row)} />;
}

/** Apply to the link in a `linkedRows` table's row header so it covers the whole row. */
export const tableRowLink = styles.rowLink;

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
  /** Styles for this column's header and body cells. */
  readonly xstyle?: StyleXProp;
  /** Styles for this column's body cells only. */
  readonly cellXstyle?: StyleXProp;
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
  /** Let the link in each row's header cover its whole row; style that link with `tableRowLink`. */
  readonly linkedRows?: boolean;
  /** `sheet` drops the outer rules for a table that fills a bordered sheet or card. */
  readonly variant?: "ruled" | "sheet";
  /** The surface behind the table, which its scroll covers match. */
  readonly surface?: "paper" | "canvas" | "raised";
  readonly xstyle?: StyleXProp;
  readonly tableXstyle?: StyleXProp;
  /** Styles for every header and body cell, such as a wider inline inset. */
  readonly cellXstyle?: StyleXProp;
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
  getRowProps,
  emptyMessage = "No data to display.",
  minWidth = `${Math.max(30, columns.length * 9)}rem`,
  tableClassName,
  linkedRows = false,
  variant = "ruled",
  surface = "paper",
  className,
  xstyle,
  tableXstyle,
  cellXstyle,
  ...regionProps
}: TableProps<Row>) {
  if (columns.length === 0) {
    throw new Error("Table requires at least one column.");
  }
  const reactId = useId();
  const captionId = `paper-table-caption-${reactId.replace(/:/gu, "")}`;
  const sheet = variant === "sheet";

  return (
    <div
      {...regionProps}
      {...sx(className, styles.region, styles[surface], a11y.focusRing, xstyle)}
      role="region"
      aria-labelledby={captionId}
      tabIndex={0}
    >
      <table {...sx(tableClassName, styles.table, styles.minWidth(minWidth), sheet && styles.sheet, tableXstyle)}>
        <caption
          id={captionId}
          {...stylex.props(captionVisibility === "screen-reader" ? a11y.visuallyHidden : styles.caption)}
        >
          {caption}
        </caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                scope="col"
                {...stylex.props(
                  styles.cell,
                  styles.header,
                  styles[column.align ?? "start"],
                  column.width !== undefined && styles.width(column.width),
                  cellXstyle,
                  column.xstyle,
                )}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td {...stylex.props(styles.empty)} colSpan={columns.length}>
                {emptyMessage}
              </td>
            </tr>
          ) : (
            rows.map((row, rowIndex) => (
              <TableRow
                {...getRowProps?.(row, rowIndex)}
                linked={linkedRows}
                key={getRowKey?.(row, rowIndex) ?? rowIndex}
              >
                {columns.map((column) => {
                  const Cell = column.rowHeader ? "th" : "td";
                  return (
                    <Cell
                      key={column.id}
                      {...(column.rowHeader ? { scope: "row" as const } : {})}
                      {...stylex.props(
                        styles.cell,
                        column.rowHeader && styles.rowHeader,
                        styles[column.align ?? "start"],
                        sheet && rowIndex === rows.length - 1 && styles.lastRowCell,
                        cellXstyle,
                        column.xstyle,
                        column.cellXstyle,
                      )}
                    >
                      {column.cell(row, rowIndex)}
                    </Cell>
                  );
                })}
              </TableRow>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
