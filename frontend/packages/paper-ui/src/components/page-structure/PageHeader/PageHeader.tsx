import type { ReactNode } from "react";
import { ChevronLeftIcon, iconClassName } from "../../../icons";
import type { NavigationLinkProps, NavigationLinkRenderer } from "../../navigation";
import { typography } from "../../../foundations/recipes";
import { tv } from "../../../foundations/variants";

const pageHeader = tv({
  slots: {
    root: "paper-page-header mb-8 mt-0 min-w-0",
    back: [
      "paper-page-header__back mb-4 inline-flex items-center gap-1 text-sm leading-150 text-link max-md:py-1.5",
      "underline underline-offset-link hover:text-link-hover hover:decoration-hover rtl:*:scale-x-mirror",
    ],
    main: "paper-page-header__main flex flex-wrap items-baseline justify-between gap-x-6 gap-y-4 lg:flex-nowrap",
    text: "paper-page-header__text flex min-w-0 flex-auto flex-col gap-2",
    heading: "paper-page-header__heading my-page-title-trim flex flex-wrap items-center gap-x-4 gap-y-3",
    title: "paper-page-header__title paper-type-page m-0 min-w-0 wrap-anywhere",
    status: "paper-page-header__status my-page-title-trim flex flex-wrap items-center gap-2",
    lead: "paper-page-header__lead m-0 max-w-measure text-base leading-165 text-muted",
    /** One row of controls centres on the title's first line and takes no more room than the heading row. */
    actions: "paper-page-header__actions my-page-actions-trim flex shrink-0 flex-wrap items-center gap-inline self-start whitespace-nowrap max-md:my-0",
  },
});

export type PageHeaderSlot = keyof typeof pageHeader.slots;

export interface PageHeaderProps {
  /** The page's only level-one heading. */
  readonly title: ReactNode;
  readonly lead?: ReactNode;
  /** Status tags shown beside the title, outside its accessible name. */
  readonly status?: ReactNode;
  readonly actions?: ReactNode;
  readonly backLink?: { readonly href: string; readonly label: string };
  /** Renders the back link with the application's router; defaults to a native link. */
  readonly renderLink?: NavigationLinkRenderer;
  /** Supporting content after the lead, such as a step indicator. */
  readonly children?: ReactNode;
  readonly className?: string;
  /** Classes merged over a part's own, for a screen's layout needs. */
  readonly classNames?: Partial<Record<PageHeaderSlot, string>>;
}

function nativeLink(props: NavigationLinkProps) {
  return <a {...props} />;
}

export function PageHeader({
  title,
  lead,
  status,
  actions,
  backLink,
  renderLink = nativeLink,
  children,
  className,
  classNames = {},
}: PageHeaderProps) {
  const slots = pageHeader();
  return (
    <header className={slots.root({ className: [classNames.root, className] })}>
      {backLink
        ? renderLink({
            href: backLink.href,
            className: slots.back({ className: classNames.back }),
            children: (
              <>
                <ChevronLeftIcon className={iconClassName("compact")} aria-hidden="true" />
                <span>{backLink.label}</span>
              </>
            ),
          })
        : null}
      <div className={slots.main({ className: classNames.main })}>
        <div className={slots.text({ className: classNames.text })}>
          <div className={slots.heading({ className: classNames.heading })}>
            <h1 className={slots.title({ className: [typography({ role: "page" }), classNames.title] })}>{title}</h1>
            {status ? <div className={slots.status({ className: classNames.status })}>{status}</div> : null}
          </div>
          {lead ? <p className={slots.lead({ className: classNames.lead })}>{lead}</p> : null}
          {children}
        </div>
        {actions ? <div className={slots.actions({ className: classNames.actions })}>{actions}</div> : null}
      </div>
    </header>
  );
}
