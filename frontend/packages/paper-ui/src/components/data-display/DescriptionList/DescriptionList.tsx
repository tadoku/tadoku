import { tv } from "../../../foundations/variants";
import type { ReactNode } from "react";

export interface DescriptionListItem {
  readonly label: ReactNode;
  readonly value: ReactNode;
}

export interface DescriptionListProps {
  readonly items: readonly DescriptionListItem[];
  /** `stacked` puts each label above its value; `inline` puts them on one ruled row. */
  readonly layout?: "stacked" | "inline";
  readonly className?: string;
}

const descriptionList = tv({
  slots: {
    root: "paper-description-list m-0 min-w-0 text-sm leading-155",
    item: "paper-description-list__item min-w-0",
    label: "paper-description-list__label wrap-anywhere",
    value: "paper-description-list__value m-0 tabular-nums wrap-anywhere",
  },
  variants: {
    layout: {
      stacked: {
        root: "paper-description-list--stacked grid gap-4",
        label: "mb-1 text-metadata leading-140 text-muted",
      },
      inline: {
        root: "paper-description-list--inline",
        item: [
          "grid grid-cols-facts-inline items-baseline gap-x-4 gap-y-1 border-b border-rule py-3",
          "first:pt-0 last:border-b-0 last:pb-0",
        ],
        label: "text-muted",
        value: "text-start",
      },
    },
  },
});

export function DescriptionList({ items, layout = "stacked", className }: DescriptionListProps) {
  const slots = descriptionList({ layout });
  return (
    <dl className={slots.root({ className })}>
      {items.map((item, index) => (
        <div key={index} className={slots.item()}>
          <dt className={slots.label()}>{item.label}</dt>
          <dd className={slots.value()}>{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
