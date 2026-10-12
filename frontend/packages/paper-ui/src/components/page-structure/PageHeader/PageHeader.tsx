import type { ReactNode } from "react";
import { ChevronLeftIcon, iconClassName } from "../../../icons";
import type { NavigationLinkProps, NavigationLinkRenderer } from "../../navigation";

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
}: PageHeaderProps) {
  return (
    <header className={className ? `paper-page-header ${className}` : "paper-page-header"}>
      {backLink
        ? renderLink({
            href: backLink.href,
            className: "paper-page-header__back",
            children: (
              <>
                <ChevronLeftIcon className={iconClassName("compact")} aria-hidden="true" />
                <span>{backLink.label}</span>
              </>
            ),
          })
        : null}
      <div className="paper-page-header__main">
        <div className="paper-page-header__text">
          <div className="paper-page-header__heading">
            <h1 className="paper-page-header__title paper-type-page">{title}</h1>
            {status ? <div className="paper-page-header__status">{status}</div> : null}
          </div>
          {lead ? <p className="paper-page-header__lead">{lead}</p> : null}
          {children}
        </div>
        {actions ? <div className="paper-page-header__actions">{actions}</div> : null}
      </div>
    </header>
  );
}
