import { DateTime } from "luxon";
import { useId } from "react";
import { tv } from "../../../foundations/variants";

const chart = tv({
  slots: {
    root: "paper-stacked-bar-chart grid min-w-0 gap-2",
    plot: "paper-stacked-bar-chart__plot block h-auto w-full overflow-visible",
    grid: "stroke-rule-subtle stroke-hairline",
    axis: "paper-stacked-bar-chart__axis fill-muted font-body text-2xs tabular-nums",
    day: "paper-stacked-bar-chart__day",
    total: "paper-stacked-bar-chart__total fill-none stroke-ink stroke-ring",
    hover: "fill-transparent",
    legend: "paper-stacked-bar-chart__legend m-0 flex list-none flex-wrap gap-x-4 gap-y-1 p-0 text-fine text-muted",
    legendItem: "inline-flex items-center gap-1.5",
    swatch: "inline-block size-3 flex-none",
    totalSwatch: "inline-block h-0.5 w-3 flex-none bg-ink",
  },
});

const slots = chart();
const number = new Intl.NumberFormat("en", { maximumFractionDigits: 1 });
const thousands = (value: number) => (value >= 1000 ? `${number.format(value / 1000)}k` : number.format(value));

export interface StackedBarChartSeries {
  readonly id: string;
  readonly label: string;
  /** A CSS colour, such as `var(--color-activity-reading)`. */
  readonly color: string;
}

export interface StackedBarChartDay {
  /** An ISO date. Days outside `start` to `end` are ignored. */
  readonly date: string;
  /** Points per series id. Missing series count as 0. */
  readonly values: Readonly<Record<string, number>>;
}

export interface StackedBarChartProps {
  /** First ISO date of the range. Every day from `start` to `end` gets a slot. */
  readonly start: string;
  /** Last ISO date of the range, inclusive. */
  readonly end: string;
  readonly days: readonly StackedBarChartDay[];
  /** Stacked from the bottom in this order. The legend lists those with points. */
  readonly series: readonly StackedBarChartSeries[];
  /** Draws the cumulative total as a line on its own right axis. */
  readonly runningTotal?: boolean;
  /** The chart's accessible name. */
  readonly label: string;
  /** A text summary for assistive technology, such as the total and the best day. */
  readonly description?: string;
}

interface Slot {
  readonly date: DateTime;
  readonly values: Readonly<Record<string, number>>;
  readonly total: number;
  readonly running: number;
}

interface Geometry {
  readonly name: "desktop" | "phone";
  readonly width: number;
  readonly height: number;
  readonly left: number;
  readonly right: number;
  readonly dayStep: number;
  readonly monthStep: number;
}

const TOP = 8;
const BOTTOM = 20;
const GEOMETRIES: readonly Geometry[] = [
  { name: "desktop", width: 960, height: 230, left: 40, right: 44, dayStep: 3, monthStep: 1 },
  { name: "phone", width: 340, height: 180, left: 32, right: 34, dayStep: 7, monthStep: 2 },
];

const STEPS = [1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10];

function roundMaximum(value: number) {
  if (value <= 0) return 10;
  const magnitude = 10 ** Math.floor(Math.log10(value));
  return STEPS.map((step) => step * magnitude).find((candidate) => candidate >= value) ?? 10 * magnitude;
}

function dayTitle(slot: Slot, series: readonly StackedBarChartSeries[], runningTotal: boolean) {
  const parts = series
    .filter((item) => (slot.values[item.id] ?? 0) > 0)
    .map((item) => `${item.label} ${number.format(slot.values[item.id])}`);
  return [
    `${slot.date.toFormat("d LLL yyyy", { locale: "en-GB" })}: ${number.format(slot.total)} points`,
    runningTotal ? `, ${number.format(slot.running)} in total` : "",
    parts.length ? ` (${parts.join(", ")})` : "",
  ].join("");
}

/**
 * Points per day across a date range, stacked by series, with an optional running-total line. It renders on the
 * server: a desktop and a phone drawing are both in the markup and the tablet breakpoint shows one.
 */
export function StackedBarChart({
  start,
  end,
  days,
  series,
  runningTotal = false,
  label,
  description,
}: StackedBarChartProps) {
  const descriptionId = useId();
  const first = DateTime.fromISO(start, { zone: "utc" });
  const count = Math.max(0, Math.round(DateTime.fromISO(end, { zone: "utc" }).diff(first, "days").days) + 1);
  const byDate = new Map(days.map((day) => [day.date, day.values]));

  let running = 0;
  const slotsByDay: Slot[] = Array.from({ length: count }, (_, index) => {
    const date = first.plus({ days: index });
    const values = byDate.get(date.toISODate() ?? "") ?? {};
    const total = series.reduce((sum, item) => sum + (values[item.id] ?? 0), 0);
    running += total;
    return { date, values, total, running };
  });
  const present = series.filter((item) => slotsByDay.some((slot) => (slot.values[item.id] ?? 0) > 0));
  const scale = {
    points: roundMaximum(Math.max(0, ...slotsByDay.map((slot) => slot.total))),
    total: roundMaximum(running),
  };

  return (
    <div className={slots.root()} role="img" aria-label={label} aria-describedby={description ? descriptionId : undefined}>
      {GEOMETRIES.map((geometry) => (
        <Drawing
          key={geometry.name}
          geometry={geometry}
          days={slotsByDay}
          series={series}
          scale={scale}
          runningTotal={runningTotal}
        />
      ))}
      <ul className={slots.legend()}>
        {present.map((item) => (
          <li key={item.id} className={slots.legendItem()}>
            <span className={slots.swatch()} style={{ background: item.color }} />
            {item.label}
          </li>
        ))}
        {runningTotal ? (
          <li className={slots.legendItem()}>
            <span className={slots.totalSwatch()} />
            Running total
          </li>
        ) : null}
      </ul>
      {description ? <p id={descriptionId} hidden>{description}</p> : null}
    </div>
  );
}

function Drawing({
  geometry,
  days,
  series,
  scale,
  runningTotal,
}: {
  readonly geometry: Geometry;
  readonly days: readonly Slot[];
  readonly series: readonly StackedBarChartSeries[];
  readonly scale: { readonly points: number; readonly total: number };
  readonly runningTotal: boolean;
}) {
  const { width, height, left, dayStep, monthStep } = geometry;
  const right = runningTotal ? geometry.right : 8;
  const plotWidth = width - left - right;
  const plotHeight = height - TOP - BOTTOM;
  const slotWidth = plotWidth / Math.max(days.length, 1);
  const bottom = TOP + plotHeight;
  const y = (value: number, max: number) => bottom - (plotHeight * value) / max;
  const short = days.length <= 45;
  const path = days
    .map((day, index) => `${index ? "L" : "M"}${(left + (index + 0.5) * slotWidth).toFixed(2)},${y(day.running, scale.total).toFixed(2)}`)
    .join("");

  return (
    <svg
      className={slots.plot({
        className: `paper-stacked-bar-chart__plot--${geometry.name} ${geometry.name === "desktop" ? "max-tablet:hidden" : "tablet:hidden"}`,
      })}
      viewBox={`0 0 ${width} ${height}`}
    >
      {[0, 0.5, 1].map((fraction) => {
        const lineY = bottom - plotHeight * fraction;
        return (
          <g key={fraction}>
            <line className={slots.grid()} x1={left} x2={width - right} y1={lineY} y2={lineY} />
            <text
              className={slots.axis({ className: "paper-stacked-bar-chart__axis--points" })}
              x={left - 4}
              y={lineY + 3}
              textAnchor="end"
            >
              {number.format(scale.points * fraction)}
            </text>
            {runningTotal ? (
              <text className={slots.axis({ className: "paper-stacked-bar-chart__axis--total" })} x={width - right + 4} y={lineY + 3}>
                {thousands(scale.total * fraction)}
              </text>
            ) : null}
          </g>
        );
      })}
      {days.map((day, index) => {
        const x = left + index * slotWidth;
        let stackTop = bottom;
        return (
          <g key={index} className={slots.day()}>
            <title>{dayTitle(day, series, runningTotal)}</title>
            <rect className={slots.hover()} x={x} y={TOP} width={slotWidth} height={plotHeight} />
            {series.map((item) => {
              const value = day.values[item.id] ?? 0;
              if (value <= 0) return null;
              const barHeight = (plotHeight * value) / scale.points;
              stackTop -= barHeight;
              return (
                <rect
                  key={item.id}
                  data-series={item.id}
                  x={(x + slotWidth * 0.12).toFixed(2)}
                  y={stackTop.toFixed(2)}
                  width={Math.max(slotWidth * 0.76, 0.8).toFixed(2)}
                  height={barHeight.toFixed(2)}
                  style={{ fill: item.color }}
                />
              );
            })}
          </g>
        );
      })}
      {runningTotal ? <path className={slots.total()} d={path} /> : null}
      {days.map((day, index) => {
        const tick = short
          ? index % dayStep === 0
          : day.date.day === 1 && (day.date.month - 1) % monthStep === 0;
        if (!tick) return null;
        return (
          <text
            key={index}
            className={slots.axis({ className: "paper-stacked-bar-chart__axis--date" })}
            x={left + index * slotWidth}
            y={height - 5}
          >
            {day.date.toFormat(short ? "d LLL" : "LLL", { locale: "en-GB" })}
          </text>
        );
      })}
    </svg>
  );
}
