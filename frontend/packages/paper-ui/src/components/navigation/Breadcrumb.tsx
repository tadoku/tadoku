import type { ReactNode } from "react";
import { ChevronRightIcon } from "../../icons";
import { renderNavigationLink, type NavigationLinkRenderer } from "./navigation";

export interface BreadcrumbItem {
  readonly id: string;
  readonly label: string;
  readonly href?: string;
  readonly icon?: ReactNode;
}

export interface BreadcrumbProps {
  readonly items: readonly BreadcrumbItem[];
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
}

export function Breadcrumb({ items, renderLink, label = "Breadcrumb" }: BreadcrumbProps) {
  return (
    <nav className="paper-breadcrumb" aria-label={label}>
      <ol className="paper-breadcrumb__list">
        {items.map((item, index) => {
          const current = index === items.length - 1;
          return (
            <li className="paper-breadcrumb__item" key={item.id}>
              {current || !item.href ? (
                <span className="paper-breadcrumb__current" aria-current={current ? "page" : undefined}>
                  {item.icon ? <span className="paper-navigation__icon" aria-hidden="true">{item.icon}</span> : null}
                  <span>{item.label}</span>
                </span>
              ) : renderNavigationLink(renderLink, {
                href: item.href,
                className: "paper-breadcrumb__link",
                children: (
                  <>
                    {item.icon ? <span className="paper-navigation__icon" aria-hidden="true">{item.icon}</span> : null}
                    <span>{item.label}</span>
                  </>
                ),
              })}
              {!current ? <ChevronRightIcon className="paper-breadcrumb__separator paper-icon-compact" aria-hidden="true" /> : null}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
