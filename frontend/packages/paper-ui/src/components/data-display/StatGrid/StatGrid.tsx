import type { ReactNode } from "react";

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
  return (
    <dl className={className ? `paper-stat-grid ${className}` : "paper-stat-grid"}>
      {stats.map((stat, index) => (
        <div key={index} className="paper-stat-grid__item">
          <dt className="paper-stat-grid__label">{stat.label}</dt>
          <dd className="paper-stat-grid__value">
            {stat.value}
            {stat.detail ? <span className="paper-stat-grid__detail">{stat.detail}</span> : null}
          </dd>
        </div>
      ))}
    </dl>
  );
}
