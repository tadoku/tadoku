import type { ReactNode } from "react";
import { tv } from "../../../foundations/variants";

const statGrid = tv({
  slots: {
    root: "paper-stat-grid m-0 grid min-w-0 grid-cols-stats gap-x-8 gap-y-5",
    item: "paper-stat-grid__item row-span-2 grid min-w-0 grid-rows-subgrid gap-y-1",
    label: "paper-stat-grid__label self-start text-sm leading-150 text-muted wrap-anywhere",
    value: "paper-stat-grid__value m-0 text-section leading-130 font-bold text-ink tabular-nums wrap-anywhere",
    detail: "paper-stat-grid__detail mt-1 block text-fine leading-150 font-regular text-muted",
  },
});

export interface StatGridItem {
  readonly label: ReactNode;
  readonly value: ReactNode;
  /** Supporting context read with the value, such as "12 active days". */
  readonly detail?: ReactNode;
}

export interface StatGridProps {
  readonly stats: readonly StatGridItem[];
  readonly className?: string;
}

export function StatGrid({ stats, className }: StatGridProps) {
  const slots = statGrid();
  return (
    <dl className={slots.root({ className })}>
      {stats.map((stat, index) => (
        <div key={index} className={slots.item()}>
          <dt className={slots.label()}>{stat.label}</dt>
          <dd className={slots.value()}>
            {stat.value}
            {stat.detail ? <span className={slots.detail()}>{stat.detail}</span> : null}
          </dd>
        </div>
      ))}
    </dl>
  );
}
