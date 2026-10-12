import type { HTMLAttributes, ReactNode, Ref } from "react";

export interface EmptyStateProps extends Omit<HTMLAttributes<HTMLDivElement>, "title"> {
  readonly title: ReactNode;
  /** Where the state sits in the page outline: 1 when it replaces the page. */
  readonly headingLevel: 1 | 2 | 3 | 4 | 5 | 6;
  readonly description?: ReactNode;
  readonly actions?: ReactNode;
  /** `surface` stands alone on paper; `plain` sits inside an existing container. */
  readonly variant?: "surface" | "plain";
  /** Receives the heading, which becomes programmatically focusable for focus moves. */
  readonly headingRef?: Ref<HTMLHeadingElement>;
}

export function EmptyState({
  title,
  headingLevel,
  description,
  actions,
  variant = "surface",
  headingRef,
  className,
  ...props
}: EmptyStateProps) {
  const Heading = `h${headingLevel}` as const;
  const typeRole =
    headingLevel === 1 ? "paper-type-page" : headingLevel === 2 ? "paper-type-section" : "paper-type-component";

  return (
    <div
      {...props}
      className={`paper-empty-state paper-empty-state--${variant}${className ? ` ${className}` : ""}`}
    >
      <Heading
        ref={headingRef}
        tabIndex={headingRef ? -1 : undefined}
        className={`paper-empty-state__title ${typeRole}`}
      >
        {title}
      </Heading>
      {description ? <p className="paper-empty-state__description">{description}</p> : null}
      {actions ? <div className="paper-empty-state__actions">{actions}</div> : null}
    </div>
  );
}
