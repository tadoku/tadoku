import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { ChevronRightIcon } from "../../icons";
import { IconSlot } from "../../foundations/IconSlot";
import { a11y, icon, link } from "../../foundations/recipes";
import { breadcrumb, list } from "./navigation-styles";
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
    <nav aria-label={label}>
      <ol {...stylex.props(list.reset, breadcrumb.list)}>
        {items.map((item, index) => {
          const current = index === items.length - 1;
          return (
            <li {...stylex.props(breadcrumb.item, index >= items.length - 2 && breadcrumb.itemKept)} key={item.id}>
              {current || !item.href ? (
                <span {...stylex.props(breadcrumb.current)} aria-current={current ? "page" : undefined}>
                  {item.icon ? <IconSlot>{item.icon}</IconSlot> : null}
                  <span>{item.label}</span>
                </span>
              ) : renderNavigationLink(renderLink, {
                href: item.href,
                ...stylex.props(link.base, breadcrumb.link, a11y.focusRing),
                children: (
                  <>
                    {item.icon ? <IconSlot>{item.icon}</IconSlot> : null}
                    <span>{item.label}</span>
                  </>
                ),
              })}
              {!current ? <ChevronRightIcon {...stylex.props(icon.compact, breadcrumb.separator)} aria-hidden="true" /> : null}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
