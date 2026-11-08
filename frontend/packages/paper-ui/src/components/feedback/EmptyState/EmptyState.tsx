import type { HTMLAttributes, ReactNode, Ref } from "react";
import { typography } from "../../../foundations/recipes";
import { tv } from "../../../foundations/variants";

const emptyState = tv({
  slots: {
    root: "paper-empty-state flex min-w-0 flex-col items-start gap-3",
    title: "paper-empty-state__title m-0 max-w-full wrap-anywhere",
    description: "paper-empty-state__description m-0 max-w-measure-description",
    actions: "paper-empty-state__actions mt-2 flex flex-wrap items-center gap-inline",
  },
  variants: {
    variant: {
      surface: { root: "paper-empty-state--surface border border-rule bg-paper px-6 py-8" },
      plain: { root: "paper-empty-state--plain py-6" },
    },
    focusable: { true: { title: "focus:outline-none" } },
  },
});

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
  const role = headingLevel === 1 ? "page" : headingLevel === 2 ? "section" : "component";
  const slots = emptyState({ variant, focusable: Boolean(headingRef) });

  return (
    <div
      {...props}
      className={slots.root({ className })}
    >
      <Heading
        ref={headingRef}
        tabIndex={headingRef ? -1 : undefined}
        className={slots.title({ className: [`paper-type-${role}`, typography({ role }), "font-display"] })}
      >
        {title}
      </Heading>
      {description ? <p className={slots.description()}>{description}</p> : null}
      {actions ? <div className={slots.actions()}>{actions}</div> : null}
    </div>
  );
}
