import * as stylex from "@stylexjs/stylex";
import { useEffect, useRef } from "react";
import { IconSlot } from "../../foundations/IconSlot";
import type { StyleXProp } from "../../foundations/props";
import { a11y, control } from "../../foundations/recipes";
import { list, tabbar } from "./navigation-styles";
import {
  disabledLinkClick,
  linkIsCurrent,
  renderNavigationLink,
  type NavigationItem,
  type NavigationLinkRenderer,
} from "./navigation";

export interface TabbarItem extends NavigationItem {
  readonly href: string;
}

export interface TabbarProps {
  readonly links: readonly TabbarItem[];
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
  readonly xstyle?: StyleXProp;
  /** Layout for the list of tabs, such as an inline inset matching the content below. */
  readonly listXstyle?: StyleXProp;
  readonly linkXstyle?: StyleXProp;
}

function TabbarList({
  links,
  currentPath,
  renderLink,
  orientation,
  listXstyle,
  linkXstyle,
}: TabbarProps & { readonly orientation: "horizontal" | "vertical" }) {
  const listRef = useRef<HTMLUListElement>(null);
  const currentId = links.find((link) => linkIsCurrent(link, currentPath))?.id;
  useEffect(() => {
    const list = listRef.current;
    if (!list || orientation !== "horizontal" || !currentId) return;
    const revealCurrent = () => {
      const current = list.querySelector('[aria-current="page"]');
      if (!current) return;
      const viewport = list.getBoundingClientRect();
      const destination = current.getBoundingClientRect();
      if (destination.left < viewport.left) list.scrollLeft += destination.left - viewport.left;
      else if (destination.right > viewport.right) list.scrollLeft += destination.right - viewport.right;
    };
    revealCurrent();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(revealCurrent);
    observer.observe(list);
    return () => observer.disconnect();
  }, [currentId, orientation]);

  return (
    <ul ref={listRef} {...stylex.props(list.reset, tabbar[orientation], listXstyle)}>
      {links.map((link) => {
        const current = linkIsCurrent(link, currentPath);
        return (
          <li key={link.id}>
            {renderNavigationLink(renderLink, {
              href: link.href,
              ...stylex.props(
                control.inherit,
                tabbar.link,
                orientation === "horizontal" ? tabbar.horizontalLink : tabbar.verticalLink,
                a11y.focusRing,
                linkXstyle,
              ),
              "aria-current": current ? "page" : undefined,
              "aria-disabled": link.disabled || undefined,
              tabIndex: link.disabled ? -1 : undefined,
              onClick: link.disabled ? disabledLinkClick : () => link.onSelect?.(),
              children: (
                <>
                  {link.icon ? <IconSlot>{link.icon}</IconSlot> : null}
                  <span>{link.label}</span>
                </>
              ),
            })}
          </li>
        );
      })}
    </ul>
  );
}

export function Tabbar({ label = "Page sections", xstyle, ...props }: TabbarProps) {
  return <nav {...stylex.props(xstyle)} aria-label={label}><TabbarList {...props} label={label} orientation="horizontal" /></nav>;
}

export function VerticalTabbar({ label = "Page sections", xstyle, ...props }: TabbarProps) {
  return <nav {...stylex.props(xstyle)} aria-label={label}><TabbarList {...props} label={label} orientation="vertical" /></nav>;
}
