import { useEffect, useRef } from "react";
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
}

function TabbarList({
  links,
  currentPath,
  renderLink,
  orientation,
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
    <ul ref={listRef} className={`paper-tabbar__list paper-tabbar__list--${orientation}`}>
      {links.map((link) => {
        const current = linkIsCurrent(link, currentPath);
        return (
          <li key={link.id}>
            {renderNavigationLink(renderLink, {
              href: link.href,
              className: "paper-tabbar__link",
              "aria-current": current ? "page" : undefined,
              "aria-disabled": link.disabled || undefined,
              tabIndex: link.disabled ? -1 : undefined,
              onClick: link.disabled ? disabledLinkClick : () => link.onSelect?.(),
              children: (
                <>
                  {link.icon ? <span className="paper-navigation__icon" aria-hidden="true">{link.icon}</span> : null}
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

export function Tabbar({ label = "Page sections", ...props }: TabbarProps) {
  return <nav className="paper-tabbar" aria-label={label}><TabbarList {...props} label={label} orientation="horizontal" /></nav>;
}

export function VerticalTabbar({ label = "Page sections", ...props }: TabbarProps) {
  return <nav className="paper-tabbar paper-tabbar--vertical" aria-label={label}><TabbarList {...props} label={label} orientation="vertical" /></nav>;
}
