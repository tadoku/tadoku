import { DateTime } from "luxon";
import { tv } from "../../foundations/variants";
import { getCellDepth, LEVELS, type HeatmapChartDatum } from "./HeatmapChart";

const recentHeatmap = tv({
  slots: {
    root: "paper-recent-heatmap grid min-w-0 gap-1",
    month: "paper-recent-heatmap__month min-w-0 overflow-visible whitespace-nowrap text-fine leading-150 text-muted",
    weekday: "paper-recent-heatmap__weekday pe-2 text-end text-fine leading-none text-muted self-center",
    cell: "paper-recent-heatmap__cell group/day relative block aspect-square min-w-0",
    tooltip: [
      "paper-recent-heatmap__tooltip pointer-events-none absolute bottom-full z-popover mb-1.5 hidden whitespace-nowrap",
      "border border-rule-strong bg-overlay px-2 py-1 text-metadata leading-150 text-ink",
      "group-hover/day:block group-focus-visible/day:block",
    ],
    legend: "paper-heatmap__legend m-0 mt-3 flex flex-wrap items-center justify-between gap-x-4 gap-y-1 text-metadata text-muted",
    legendItems: "flex flex-wrap items-center gap-x-3 gap-y-1",
    legendItem: "paper-heatmap__legend-item inline-flex items-center gap-1",
    swatch: "inline-block size-2.5",
  },
});

const slots = recentHeatmap();

const levelBackgrounds = {
  "0": "bg-heat-0",
  "1": "bg-heat-1",
  "2": "bg-heat-2",
  "3": "bg-heat-3",
  "4": "bg-heat-4",
} as const;

const WEEKDAYS = ["Mon", undefined, "Wed", undefined, "Fri", undefined, "Sun"] as const;

export interface RecentHeatmapProps {
  readonly data: readonly HeatmapChartDatum[];
  /** The current ISO day, the map's last. Later days in its week are left blank. */
  readonly today: string;
  /** How many weeks, Monday to Sunday, the map shows up to and including today's. */
  readonly weeks: number;
  /** Makes each day in `data` a link, for example to that day's records. */
  readonly getHref?: (date: string) => string;
  readonly className?: string;
}

/**
 * The last few weeks of daily activity in the annual calendar's colours and legend, with cells that grow to fill the
 * container's width, for places too narrow or too short-lived for a whole year.
 */
export function RecentHeatmap({ data, today, weeks, getHref, className }: RecentHeatmapProps) {
  const values = new Map(data.map((datum) => [datum.date, datum]));
  const start = DateTime.fromISO(today, { zone: "utc" }).startOf("week").minus({ weeks: weeks - 1 });
  const columns = Array.from({ length: weeks }, (_, week) =>
    Array.from({ length: 7 }, (_, day) => start.plus({ weeks: week, days: day }).toISODate()!),
  );
  // A month is named over the week holding its first day, and the first week names its month unless a new one
  // starts within the next two weeks, where the labels would collide.
  const months = columns.map((column, index) => {
    const first = column.find((date) => date.endsWith("-01"));
    if (first) return DateTime.fromISO(first).toFormat("LLL", { locale: "en-US" });
    if (index === 0 && !columns.slice(1, 3).flat().some((date) => date.endsWith("-01"))) {
      return DateTime.fromISO(column[0]).toFormat("LLL", { locale: "en-US" });
    }
    return undefined;
  });

  return (
    <div className={className}>
      <div
        className={slots.root()}
        role="group"
        aria-label={`Daily activity for the past ${weeks} weeks`}
        style={{ gridTemplateColumns: `auto repeat(${weeks}, minmax(0, 1fr))` }}
      >
        <span aria-hidden="true" />
        {months.map((label, index) => (
          <span key={index} className={slots.month()} aria-hidden="true">{label}</span>
        ))}
        {WEEKDAYS.map((weekday, row) => [
          <span key={`weekday-${row}`} className={slots.weekday()} aria-hidden="true">{weekday}</span>,
          ...columns.map((column, index) => {
            const date = column[row];
            // Days after today keep their place in the week but draw nothing, so the map ends on today.
            if (date > today) return <span key={date} aria-hidden="true" />;
            const datum = values.get(date);
            const level = getCellDepth(datum?.value ?? 0);
            const className = slots.cell({ className: levelBackgrounds[level as keyof typeof levelBackgrounds] });
            const label = datum?.tooltip ?? `No activity on ${DateTime.fromISO(date).toFormat("LLL d, yyyy", { locale: "en-US" })}`;
            const href = datum ? getHref?.(date) : undefined;
            // Tooltips near either edge open inward so they stay inside the map's column.
            const align = index < 3 ? "start-0" : index >= weeks - 3 ? "end-0" : "start-1/2 -translate-x-1/2";
            const tooltip = <span className={slots.tooltip({ className: align })} aria-hidden="true">{label}</span>;
            return href
              ? <a key={date} href={href} className={className} aria-label={label} data-date={date}>{tooltip}</a>
              : <span key={date} className={className} role="img" aria-label={label} data-date={date}>{tooltip}</span>;
          }),
        ])}
      </div>
      <p className={slots.legend()}>
        <span>Points per day</span>
        <span className={slots.legendItems()}>
          {LEVELS.map(({ level, label }) => (
            <span key={level} className={slots.legendItem()}>
              <span aria-hidden="true" className={slots.swatch({ className: levelBackgrounds[level] })} />
              {label}
            </span>
          ))}
        </span>
      </p>
    </div>
  );
}
