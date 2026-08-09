import { DateTime, Interval } from "luxon";
import {
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";

export interface HeatmapChartDatum {
  readonly date: string;
  readonly value: number;
  readonly tooltip?: string;
}

export interface HeatmapChartProps {
  readonly data: readonly HeatmapChartDatum[];
  readonly year: number;
  readonly id: string;
}

interface CalendarCell {
  readonly date: DateTime;
  readonly value: number;
  readonly tooltip?: string;
}

interface TooltipRect {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
  readonly arrowX: number;
}

const CELL_WIDTH = 10;
const CELL_HEIGHT = 10;
const CELL_GAP = 3;
const OFFSET = { x: 30, y: 15 } as const;

const WEEKDAYS = ["Mon", undefined, "Wed", undefined, "Fri", undefined, "Sun"] as const;

/**
 * Tadoku's compact annual activity calendar. The geometry and interaction
 * follow the original ui HeatmapChart so existing consumers can pass the
 * same dated activity records without reshaping them into a generic matrix.
 */
export function HeatmapChart({ id, data, year }: HeatmapChartProps) {
  const start = DateTime.fromObject({ year, month: 1, day: 1 });
  const end = DateTime.fromObject({ year, month: 12, day: 31 });
  const values = new Map<string, number>();
  const tooltips = new Map<string, string>();

  for (const datum of data) {
    values.set(datum.date, datum.value);
    if (datum.tooltip) tooltips.set(datum.date, datum.tooltip);
  }

  const dates = Interval.fromDateTimes(start, end.endOf("day"))
    .splitBy({ day: 1 })
    .map((interval) => interval.start);
  const weekCount = start.weeksInWeekYear + 1;
  const columns: (CalendarCell | undefined)[][] = Array.from(
    { length: weekCount },
    () => Array.from({ length: 7 }),
  );

  let columnIndex = 0;
  for (const date of dates) {
    if (date === null) continue;
    const isoDate = date.toISODate();
    if (isoDate === null) continue;

    columns[columnIndex][date.weekday - 1] = {
      date,
      value: values.get(isoDate) ?? 0,
      tooltip: tooltips.get(isoDate),
    };
    if (date.weekday === 7) columnIndex += 1;
  }

  const maxValue = Math.max(
    ...columns.flatMap((column) => column.map((cell) => cell?.value ?? 0)),
  );
  let lastMonth = 0;
  const months = columns.map((column) => {
    const shouldRender = column
      .filter((cell): cell is CalendarCell => cell !== undefined)
      .some((cell) => cell.date.month > lastMonth);
    if (!shouldRender) return undefined;

    lastMonth += 1;
    return DateTime.fromObject({ month: lastMonth }).toFormat("LLL");
  });

  const tooltipLayerId = `paper-heatmap-tooltip-${id}`;
  const titleId = `paper-heatmap-title-${id}`;
  const width =
    weekCount * CELL_WIDTH +
    (weekCount - 1) * CELL_GAP +
    OFFSET.x +
    10;
  const height = 7 * CELL_HEIGHT + 6 * CELL_GAP + OFFSET.y;

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className="paper-heatmap"
      role="img"
      aria-labelledby={titleId}
      data-year={year}
    >
      <title id={titleId}>Daily activity for {year}</title>
      {WEEKDAYS.map((label, row) => label ? (
        <text
          key={label}
          className="paper-heatmap__axis-label"
          textAnchor="end"
          x={OFFSET.x - CELL_GAP * 2}
          y={OFFSET.y + CELL_HEIGHT * row + CELL_GAP * row}
          dominantBaseline="hanging"
        >
          {label}
        </text>
      ) : null)}
      {months.map((label, column) => label ? (
        <text
          key={label}
          className="paper-heatmap__axis-label"
          textAnchor="start"
          x={OFFSET.x + CELL_WIDTH * column + CELL_GAP * column}
          y={0}
          dominantBaseline="hanging"
        >
          {label}
        </text>
      ) : null)}
      {columns.map((column, columnPosition) =>
        column.map((cell, rowPosition) => (
          <Cell
            key={`${columnPosition}-${rowPosition}`}
            cell={cell}
            tooltipLayerId={tooltipLayerId}
            maxValue={maxValue}
            column={columnPosition}
            row={rowPosition}
            parentWidth={width}
          />
        )),
      )}
      <g id={tooltipLayerId} className="paper-heatmap__tooltip-layer" />
    </svg>
  );
}

function Cell({
  cell,
  tooltipLayerId,
  maxValue,
  row,
  column,
  parentWidth,
}: {
  readonly cell: CalendarCell | undefined;
  readonly tooltipLayerId: string;
  readonly maxValue: number;
  readonly row: number;
  readonly column: number;
  readonly parentWidth: number;
}) {
  const [mounted, setMounted] = useState(false);
  const [tooltipVisible, setTooltipVisible] = useState(false);
  const x = OFFSET.x + CELL_WIDTH * column + CELL_GAP * column;
  const y = OFFSET.y + CELL_HEIGHT * row + CELL_GAP * row;
  const isoDate = cell?.date.toISODate() ?? undefined;
  const tooltip = cell && cell.value !== 0 ? cell.tooltip : undefined;
  const tooltipId = `${tooltipLayerId}-${column}-${row}`;
  const level = getCellDepth(maxValue, cell?.value);

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  const rect = (
    <rect
      width={CELL_WIDTH}
      height={CELL_HEIGHT}
      x={x}
      y={y}
      className="paper-heatmap__cell"
      data-level={level}
      data-date={isoDate}
      data-value={cell?.value}
      aria-hidden={cell === undefined ? true : undefined}
      aria-label={cell ? (cell.tooltip ?? `${isoDate}: ${cell.value}`) : undefined}
      aria-describedby={tooltip && tooltipVisible ? tooltipId : undefined}
      role={cell ? "graphics-symbol" : undefined}
      tabIndex={tooltip ? 0 : undefined}
      onMouseOver={tooltip ? () => setTooltipVisible(true) : undefined}
      onMouseOut={tooltip ? () => setTooltipVisible(false) : undefined}
      onFocus={tooltip ? () => setTooltipVisible(true) : undefined}
      onBlur={tooltip ? () => setTooltipVisible(false) : undefined}
      onTouchStart={tooltip ? () => setTooltipVisible(true) : undefined}
      onTouchMove={tooltip ? () => setTooltipVisible(true) : undefined}
      onTouchEnd={tooltip ? () => setTooltipVisible(false) : undefined}
      onTouchCancel={tooltip ? () => setTooltipVisible(false) : undefined}
    />
  );

  if (!mounted || !tooltip) return rect;
  const tooltipLayer = document.getElementById(tooltipLayerId);

  return (
    <>
      {rect}
      {tooltipLayer ? createPortal(
        <Tooltip
          id={tooltipId}
          row={row}
          column={column}
          visible={tooltipVisible}
          parentWidth={parentWidth}
        >
          {tooltip}
        </Tooltip>,
        tooltipLayer,
      ) : null}
    </>
  );
}

function Tooltip({
  id,
  row,
  column,
  children,
  visible,
  parentWidth,
}: {
  readonly id: string;
  readonly row: number;
  readonly column: number;
  readonly children: ReactNode;
  readonly visible: boolean;
  readonly parentWidth: number;
}) {
  const textRef = useRef<SVGTextElement>(null);
  const [rect, setRect] = useState<TooltipRect>({
    x: 0,
    y: 0,
    width: 0,
    height: 0,
    arrowX: 0,
  });

  useEffect(() => {
    if (textRef.current === null) return;
    const textRect = textRef.current.getBoundingClientRect();
    const width = textRect.width + 12;
    const height = textRect.height + 12;
    const arrowX =
      OFFSET.x + CELL_WIDTH * column + CELL_GAP * column - width / 2;
    const y = OFFSET.y + CELL_HEIGHT * row + CELL_GAP * row - height - 2;
    let x = arrowX;

    if (x < 0) x = 20;
    if (x + width > parentWidth) x = parentWidth - width;
    x = Math.max(0, x);

    setRect({ x, y, width, height, arrowX });
  }, [children, column, parentWidth, row, visible]);

  return (
    <g
      id={id}
      role="tooltip"
      className={`paper-heatmap__tooltip${visible ? "" : " hidden"}`}
      aria-hidden={!visible}
    >
      <rect
        width={rect.width}
        height={rect.height}
        x={rect.x}
        y={rect.y}
        className="paper-heatmap__tooltip-background"
      />
      <polygon
        points={pointsForRect(rect)}
        className="paper-heatmap__tooltip-background"
      />
      <text
        ref={textRef}
        x={rect.x + 6}
        y={rect.y + 8}
        dominantBaseline="hanging"
        className="paper-heatmap__tooltip-text"
      >
        {children}
      </text>
    </g>
  );
}

function pointsForRect({
  y,
  width,
  height,
  arrowX,
}: TooltipRect): string {
  const size = 8;
  const middle = arrowX + width / 2 + CELL_WIDTH / 2;
  const left = middle - size / 2;
  const right = middle + size / 2;
  const top = y + height - 1;
  const bottom = top + 5;

  return `${left},${top} ${right},${top} ${middle},${bottom}`;
}

function getCellDepth(
  maxValue: number,
  value: number | undefined,
): "empty" | "0" | "1" | "2" | "3" | "4" {
  if (value === undefined) return "empty";
  if (value === 0) return "0";

  const ratio = maxValue === 0 ? 0 : value / maxValue;
  if (ratio < 0.25) return "1";
  if (ratio < 0.5) return "2";
  if (ratio < 0.75) return "3";
  return "4";
}
