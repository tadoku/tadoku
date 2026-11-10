import { DateTime, Interval } from "luxon";
import {
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import { tv } from "../../foundations/variants";

const heatmap = tv({
  slots: {
    frame: "paper-heatmap__frame flex w-full items-start",
    weekdays: "paper-heatmap__weekdays sticky start-0 z-lift h-auto shrink-0 grow-0",
    chart: "paper-heatmap block h-auto min-w-0 max-w-none flex-1 overflow-visible",
    axis: "paper-heatmap__axis-label fill-ink font-body text-heat-axis max-tablet:text-heat-axis-phone",
    layer: "paper-heatmap__tooltip-layer pointer-events-none",
    legend: [
      "paper-heatmap__legend sticky start-0 m-0 mt-2 flex w-max max-w-full flex-wrap items-center gap-x-3 gap-y-2",
      "text-metadata text-muted",
    ],
    legendTitle: "max-tablet:basis-full",
    legendItem: "paper-heatmap__legend-item inline-flex items-center gap-1",
    cell: [
      "paper-heatmap__cell focus:stroke-ink focus:stroke-ring focus:outline-none",
      "group-focus/day:stroke-ink group-focus/day:stroke-ring",
      "forced-colors:fill-forced-canvas forced-colors:stroke-forced-ink forced-colors:stroke-hairline",
    ],
    link: "paper-heatmap__link group/day focus:outline-none",
    tooltip: "paper-heatmap__tooltip pointer-events-none",
    tooltipBackground: "paper-heatmap__tooltip-background fill-overlay stroke-rule-strong stroke-hairline",
    tooltipText: "paper-heatmap__tooltip-text fill-ink font-body text-metadata",
  },
  variants: {
    surface: { paper: { weekdays: "bg-paper" }, canvas: { weekdays: "bg-canvas" } },
    scrolls: { true: { frame: "min-w-heatmap", legend: "max-w-heatmap-legend" } },
  },
  defaultVariants: { surface: "paper" },
});

const slots = heatmap();

const levelFills = {
  "0": "fill-heat-0",
  "1": "fill-heat-1",
  "2": "fill-heat-2",
  "3": "fill-heat-3",
  "4": "fill-heat-4",
} as const;

function cellClassName(level: string | undefined, future: boolean) {
  const fill = level && level in levelFills ? levelFills[level as keyof typeof levelFills] : "fill-transparent";
  return slots.cell({ className: future ? "fill-heat-future" : fill });
}

export interface HeatmapChartDatum {
  readonly date: string;
  readonly value: number;
  readonly tooltip?: string;
}

export interface HeatmapChartProps {
  readonly data: readonly HeatmapChartDatum[];
  readonly year: number;
  readonly id: string;
  /** Makes each day in `data` a link, including one whose value is 0, for example to that day's records. */
  readonly getHref?: (date: string) => string;
  /** The current ISO day; later days are drawn as not yet happened rather than as empty. */
  readonly today?: string;
  /** The surface behind the chart, which the pinned weekday labels repeat. */
  readonly surface?: "paper" | "canvas";
  /** Keeps the year at a readable width inside a sideways-scrolling region, with the legend in view. */
  readonly scrolls?: boolean;
}

interface CalendarCell {
  readonly date: DateTime;
  readonly value: number;
  readonly tooltip?: string;
  /** The day has a datum, so it links even when its value is 0. */
  readonly logged: boolean;
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
export const LEVELS = [
  { level: "0", lowerBound: 0, label: "0" },
  { level: "1", lowerBound: 0, label: "Under 30" },
  { level: "2", lowerBound: 30, label: "30–99" },
  { level: "3", lowerBound: 100, label: "100–299" },
  { level: "4", lowerBound: 300, label: "300+" },
] as const;

/**
 * Tadoku's compact annual activity calendar. The geometry and interaction
 * follow the original ui HeatmapChart so existing consumers can pass the
 * same dated activity records without reshaping them into a generic matrix.
 */
export function HeatmapChart({ id, data, year, getHref, today, surface = "paper", scrolls = false }: HeatmapChartProps) {
  const variant = heatmap({ surface, scrolls });
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
      logged: values.has(isoDate),
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
    return DateTime.fromObject({ month: lastMonth }, { locale: "en-US" }).toFormat("LLL");
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
      {/* The weekday gutter is its own SVG, sized to the same scale, so it can stay pinned while a narrow container
          scrolls the weeks sideways. */}
      <div className={variant.frame()}>
        <svg
          width={OFFSET.x}
          height={height}
          viewBox={`0 0 ${OFFSET.x} ${height}`}
          className={variant.weekdays()}
          style={{ flexBasis: `${(OFFSET.x / width) * 100}%` }}
          aria-hidden="true"
        >
          {WEEKDAYS.map((label, row) => label ? (
            <text
              key={label}
              className={slots.axis()}
              textAnchor="end"
              x={OFFSET.x - CELL_GAP * 2}
              y={OFFSET.y + CELL_HEIGHT * row + CELL_GAP * row}
              dominantBaseline="hanging"
            >
              {label}
            </text>
          ) : null)}
        </svg>
      <svg
        width={width - OFFSET.x}
        height={height}
        viewBox={`${OFFSET.x} 0 ${width - OFFSET.x} ${height}`}
        className={slots.chart()}
        // A group, not an image: active days are links that assistive technology must reach.
        role="group"
        aria-labelledby={titleId}
        data-year={year}
      >
        <title id={titleId}>Daily activity for {year}</title>
        {months.map((label, column) => label ? (
          <text
            key={label}
            className={slots.axis()}
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
        <g id={tooltipLayerId} className={slots.layer()} />
      </svg>
      </div>
      <p className={variant.legend()}>
        <span className={slots.legendTitle()}>Points per day</span>
        {LEVELS.map(({ level, label }) => (
          <span key={level} className={slots.legendItem()}>
            <svg width={CELL_WIDTH} height={CELL_HEIGHT} aria-hidden="true">
              <rect width={CELL_WIDTH} height={CELL_HEIGHT} className={cellClassName(level, false)} shapeRendering="crispEdges" data-level={level} />
            </svg>
            {label}
          </span>
        ))}
        {today && today < end.toISODate()! ? (
          <span className={slots.legendItem()}>
            <svg width={CELL_WIDTH} height={CELL_HEIGHT} aria-hidden="true">
              <rect width={CELL_WIDTH} height={CELL_HEIGHT} className={cellClassName(undefined, true)} shapeRendering="crispEdges" data-future="" />
            </svg>
            Not yet
          </span>
        ) : null}
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
  const tooltip = cell?.logged ? cell.tooltip : undefined;
  const tooltipId = `${tooltipLayerId}-${column}-${row}`;
  const level = getCellDepth(cell?.value);
  const href = cell?.logged && isoDate ? getHref?.(isoDate) : undefined;
  const future = Boolean(today && isoDate && isoDate > today);
  const label = cell ? (future ? `${isoDate}: not yet` : (cell.tooltip ?? `${isoDate}: ${cell.value}`)) : undefined;
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
      className={cellClassName(level, future)}
      shapeRendering="crispEdges"
      data-level={level}
      data-future={future ? "" : undefined}
      data-today={today && isoDate === today ? "" : undefined}
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
  const rect = href ? <a href={href} className={slots.link()} {...focusProps}>{shape}</a> : shape;

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

    // The weekday gutter is a separate, pinned SVG; tooltips stay right of it.
    if (x < OFFSET.x) x = OFFSET.x;
    if (x + width > parentWidth) x = parentWidth - width;
    x = Math.max(OFFSET.x, x);

    setRect({ x, y, width, height, arrowX });
  }, [children, column, parentWidth, row, visible]);

  return (
    <g
      id={id}
      role="tooltip"
      className={slots.tooltip({ className: visible ? undefined : "hidden" })}
      aria-hidden={!visible}
    >
      <path d={outlineForRect(rect)} className={slots.tooltipBackground()} />
      <text
        ref={textRef}
        x={rect.x + 6}
        y={rect.y + 8}
        dominantBaseline="hanging"
        className={slots.tooltipText()}
      >
        {children}
      </text>
    </g>
  );
}

/** The tooltip box and its pointer as one outline, so the border runs around the pointer instead of across it. */
function outlineForRect({ x, y, width, height, arrowX }: TooltipRect): string {
  const size = 8;
  const middle = arrowX + width / 2 + CELL_WIDTH / 2;
  const bottom = y + height;

  return [
    `M${x},${y}`,
    `H${x + width}`,
    `V${bottom}`,
    `H${middle + size / 2}`,
    `L${middle},${bottom + 5}`,
    `L${middle - size / 2},${bottom}`,
    `H${x}`,
    "Z",
  ].join(" ");
}

export function getCellDepth(value: number | undefined): "empty" | (typeof LEVELS)[number]["level"] {
  if (value === undefined) return "empty";

  // Bins follow the one-decimal precision scores are displayed with.
  const rounded = Math.round(value * 10) / 10;
  if (rounded === 0) return "0";
  return [...LEVELS].reverse().find((bin) => bin.level !== "0" && rounded >= bin.lowerBound)?.level ?? "0";
}
