import type { ReactNode } from "react";
import { ChevronRightIcon, iconClassName } from "../../icons";
import { renderNavigationLink, type NavigationLinkRenderer } from "./navigation";
import { breadcrumb, navigationIcon } from "./styles";

const slots = breadcrumb();

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
      <ol className={slots.list()}>
        {items.map((item, index) => {
          const current = index === items.length - 1;
          return (
            <li className={slots.item({ className: index < items.length - 2 ? slots.distant() : undefined })} key={item.id}>
              {current || !item.href ? (
                <span className={slots.current()} aria-current={current ? "page" : undefined}>
                  {item.icon ? <span className={navigationIcon} aria-hidden="true">{item.icon}</span> : null}
                  <span>{item.label}</span>
                </span>
              ) : renderNavigationLink(renderLink, {
                href: item.href,
                className: slots.link(),
                children: (
                  <>
                    {item.icon ? <span className={navigationIcon} aria-hidden="true">{item.icon}</span> : null}
                    <span>{item.label}</span>
                  </>
                ),
              })}
              {!current ? <ChevronRightIcon className={iconClassName("compact", slots.separator())} aria-hidden="true" /> : null}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
