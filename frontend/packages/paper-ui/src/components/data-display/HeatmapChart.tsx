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
  /** Makes each day with activity a link, for example to that day's records. */
  readonly getHref?: (date: string) => string;
  /** The current ISO day; later days are drawn as not yet happened rather than as empty. */
  readonly today?: string;
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

// Fixed points-per-day bins keep a record day from recolouring earlier activity.
const LEVELS = [
  { level: "0", lowerBound: 0, label: "0" },
  { level: "1", lowerBound: 0, label: "< 30" },
  { level: "2", lowerBound: 30, label: "30+" },
  { level: "3", lowerBound: 100, label: "100+" },
  { level: "4", lowerBound: 300, label: "300+" },
] as const;

/**
 * Tadoku's compact annual activity calendar. The geometry and interaction
 * follow the original ui HeatmapChart so existing consumers can pass the
 * same dated activity records without reshaping them into a generic matrix.
 */
export function HeatmapChart({ id, data, year, getHref, today }: HeatmapChartProps) {
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
    <>
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
              getHref={getHref}
              today={today}
              column={columnPosition}
              row={rowPosition}
              parentWidth={width}
            />
          )),
        )}
        <g id={tooltipLayerId} className="paper-heatmap__tooltip-layer" />
      </svg>
      <p className="paper-heatmap__legend">
        <span>Points per day</span>
        {LEVELS.map(({ level, label }) => (
          <span key={level} className="paper-heatmap__legend-item">
            <svg width={CELL_WIDTH} height={CELL_HEIGHT} aria-hidden="true">
              <rect width={CELL_WIDTH} height={CELL_HEIGHT} className="paper-heatmap__cell" data-level={level} />
            </svg>
            {label}
          </span>
        ))}
      </p>
    </>
  );
}

function Cell({
  cell,
  tooltipLayerId,
  getHref,
  today,
  row,
  column,
  parentWidth,
}: {
  readonly cell: CalendarCell | undefined;
  readonly tooltipLayerId: string;
  readonly getHref: ((date: string) => string) | undefined;
  readonly today: string | undefined;
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
  const level = getCellDepth(cell?.value);
  const href = cell && isoDate && cell.value !== 0 ? getHref?.(isoDate) : undefined;
  const label = cell ? (cell.tooltip ?? `${isoDate}: ${cell.value}`) : undefined;
  const showTooltip = tooltip ? () => setTooltipVisible(true) : undefined;
  const hideTooltip = tooltip ? () => setTooltipVisible(false) : undefined;
  const focusProps = {
    "aria-label": label,
    "aria-describedby": tooltip && tooltipVisible ? tooltipId : undefined,
    onFocus: showTooltip,
    onBlur: hideTooltip,
  };

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  const shape = (
    <rect
      width={CELL_WIDTH}
      height={CELL_HEIGHT}
      x={x}
      y={y}
      className="paper-heatmap__cell"
      data-level={level}
      data-future={today && isoDate && isoDate > today ? "" : undefined}
      data-date={isoDate}
      data-value={cell?.value}
      aria-hidden={cell === undefined || href ? true : undefined}
      role={cell && !href ? "graphics-symbol" : undefined}
      tabIndex={tooltip && !href ? 0 : undefined}
      {...(href ? {} : focusProps)}
      onMouseOver={showTooltip}
      onMouseOut={hideTooltip}
      onTouchStart={showTooltip}
      onTouchMove={showTooltip}
      onTouchEnd={hideTooltip}
      onTouchCancel={hideTooltip}
    />
  );
  const rect = href ? <a href={href} className="paper-heatmap__link" {...focusProps}>{shape}</a> : shape;

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
    if (!visible || textRef.current === null) return;
    const textRect = textRef.current.getBBox();
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

function getCellDepth(value: number | undefined): "empty" | (typeof LEVELS)[number]["level"] {
  if (value === undefined) return "empty";

  // Bins follow the one-decimal precision scores are displayed with.
  const rounded = Math.round(value * 10) / 10;
  if (rounded === 0) return "0";
  return [...LEVELS].reverse().find((bin) => bin.level !== "0" && rounded >= bin.lowerBound)?.level ?? "0";
}
