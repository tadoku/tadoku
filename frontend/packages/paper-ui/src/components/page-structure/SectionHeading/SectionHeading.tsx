import type { ReactNode } from "react";
import { typography } from "../../../foundations/recipes";
import { tv } from "../../../foundations/variants";

const sectionHeading = tv({
  slots: {
    root: "paper-section-heading mb-4 mt-0 flex min-w-0 flex-wrap items-baseline justify-between gap-x-6 gap-y-3",
    text: "paper-section-heading__text flex min-w-0 flex-auto flex-col gap-2",
    heading: "paper-section-heading__heading flex flex-wrap items-center gap-x-3 gap-y-2",
    title: "paper-section-heading__title m-0 min-w-0 wrap-anywhere",
    status: "paper-section-heading__status flex flex-wrap items-center gap-inline",
    description: "paper-section-heading__description m-0 max-w-measure text-sm leading-160 text-muted",
    actions: "paper-section-heading__actions flex flex-wrap items-center gap-inline",
  },
});

export type SectionHeadingSlot = keyof typeof sectionHeading.slots;

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
  /** Classes merged over a part's own, for a screen's layout needs. */
  readonly classNames?: Partial<Record<SectionHeadingSlot, string>>;
}

export function SectionHeading({
  level,
  title,
  id,
  description,
  status,
  actions,
  className,
  classNames = {},
}: SectionHeadingProps) {
  const Heading = `h${level}` as const;
  const role = level === 2 ? "section" : "component";
  const slots = sectionHeading();

  return (
    <div className={slots.root({ className: [classNames.root, className] })}>
      <div className={slots.text({ className: classNames.text })}>
        <div className={slots.heading({ className: classNames.heading })}>
          <Heading
            id={id}
            className={slots.title({ className: [`paper-type-${role}`, typography({ role }), classNames.title] })}
          >
            {title}
          </Heading>
          {status ? <div className={slots.status({ className: classNames.status })}>{status}</div> : null}
        </div>
        {description ? <p className={slots.description({ className: classNames.description })}>{description}</p> : null}
      </div>
      {actions ? <div className={slots.actions({ className: classNames.actions })}>{actions}</div> : null}
    </div>
  );
}
