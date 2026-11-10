import type { ReactNode } from "react";
import { tv } from "../../../foundations/variants";

const statGrid = tv({
  slots: {
    root: "paper-stat-grid m-0 grid min-w-0 grid-cols-stats gap-x-6 gap-y-5 max-sm:gap-x-4",
    item: "paper-stat-grid__item row-span-2 grid min-w-0 grid-rows-subgrid gap-y-1",
    label: "paper-stat-grid__label self-start text-sm leading-150 text-muted wrap-anywhere",
    value: "paper-stat-grid__value m-0 text-section leading-130 font-bold text-ink tabular-nums wrap-anywhere",
    unit: "paper-stat-grid__unit ms-1 text-metadata font-regular text-muted",
    detail: "paper-stat-grid__detail mt-1 block text-fine leading-150 font-regular text-muted",
  },
  variants: {
    variant: {
      plain: {},
      sheet: {
        root: [
          "paper-stat-grid--sheet grid-flow-col auto-cols-fr gap-px border border-rule-subtle bg-rule-subtle",
          "max-tablet:grid-flow-row max-tablet:grid-cols-halves max-sm:gap-x-px",
        ],
        item: "gap-y-0.5 bg-paper px-3.5 py-3",
        label: "font-semibold",
        value: "text-xl leading-130",
        unit: "text-fine",
        detail: "mt-0.5",
      },
    },
  },
  defaultVariants: { variant: "plain" },
});

export interface StatGridItem {
  readonly label: ReactNode;
  readonly value: ReactNode;
  /** A unit read beside the value, such as "points". */
  readonly unit?: ReactNode;
  /** Supporting context read with the value, such as "12 active days". */
  readonly detail?: ReactNode;
}

export interface StatGridProps {
  readonly stats: readonly StatGridItem[];
  /** `plain` sits in the page; `sheet` boxes the figures in one ruled row, two columns on phones. */
  readonly variant?: "plain" | "sheet";
  readonly className?: string;
}

export function StatGrid({ stats, variant = "plain", className }: StatGridProps) {
  const slots = statGrid({ variant });
  return (
    <dl className={slots.root({ className })}>
      {stats.map((stat, index) => (
        <div key={index} className={slots.item()}>
          <dt className={slots.label()}>{stat.label}</dt>
          <dd className={slots.value()}>
            {stat.value}
            {stat.unit ? <span className={slots.unit()}>{stat.unit}</span> : null}
            {stat.detail ? <span className={slots.detail()}>{stat.detail}</span> : null}
          </dd>
        </div>
      ))}
    </dl>
  );
}
