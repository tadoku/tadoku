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

export function DescriptionList({ items, layout = "stacked", className }: DescriptionListProps) {
  return (
    <dl className={`paper-description-list paper-description-list--${layout}${className ? ` ${className}` : ""}`}>
      {items.map((item, index) => (
        <div key={index} className="paper-description-list__item">
          <dt className="paper-description-list__label">{item.label}</dt>
          <dd className="paper-description-list__value">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
