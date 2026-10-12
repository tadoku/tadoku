import type { ReactNode } from "react";

export interface SectionHeadingProps {
  /** h2 and h3+ use the section and component type roles. PageHeader owns h1. */
  readonly level: 2 | 3 | 4 | 5 | 6;
  readonly title: ReactNode;
  /** Applied to the heading so the enclosing section can use aria-labelledby. */
  readonly id?: string;
  readonly description?: ReactNode;
  /** Status tags shown beside the title, outside its accessible name. */
  readonly status?: ReactNode;
  readonly actions?: ReactNode;
  readonly className?: string;
}

export function SectionHeading({
  level,
  title,
  id,
  description,
  status,
  actions,
  className,
}: SectionHeadingProps) {
  const Heading = `h${level}` as const;
  const typeRole = level === 2 ? "paper-type-section" : "paper-type-component";

  return (
    <div className={className ? `paper-section-heading ${className}` : "paper-section-heading"}>
      <div className="paper-section-heading__text">
        <div className="paper-section-heading__heading">
          <Heading id={id} className={`paper-section-heading__title ${typeRole}`}>{title}</Heading>
          {status ? <div className="paper-section-heading__status">{status}</div> : null}
        </div>
        {description ? <p className="paper-section-heading__description">{description}</p> : null}
      </div>
      {actions ? <div className="paper-section-heading__actions">{actions}</div> : null}
    </div>
  );
}
