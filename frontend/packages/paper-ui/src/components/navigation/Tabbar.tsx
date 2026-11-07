import { useEffect, useRef } from "react";
import {
  disabledLinkClick,
  linkIsCurrent,
  renderNavigationLink,
  type NavigationItem,
  type NavigationLinkRenderer,
} from "./navigation";
import { navigationIcon, tabbar, type TabbarSlot } from "./styles";

export interface TabbarItem extends NavigationItem {
  readonly href: string;
}

export interface TabbarProps {
  readonly links: readonly TabbarItem[];
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
  readonly className?: string;
  /** Classes merged over a part's own, for a screen's layout needs. */
  readonly classNames?: Partial<Record<TabbarSlot, string>>;
}

function TabbarList({
  links,
  currentPath,
  renderLink,
  orientation,
  classNames,
}: TabbarProps & { readonly orientation: "horizontal" | "vertical" }) {
  const slots = tabbar({ orientation });
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
    <ul ref={listRef} className={slots.list({ className: classNames?.list })}>
      {links.map((link) => {
        const current = linkIsCurrent(link, currentPath);
        return (
          <li key={link.id}>
            {renderNavigationLink(renderLink, {
              href: link.href,
              className: slots.link({ className: classNames?.link }),
              "aria-current": current ? "page" : undefined,
              "aria-disabled": link.disabled || undefined,
              tabIndex: link.disabled ? -1 : undefined,
              onClick: link.disabled ? disabledLinkClick : () => link.onSelect?.(),
              children: (
                <>
                  {link.icon ? <span className={navigationIcon} aria-hidden="true">{link.icon}</span> : null}
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
  return (
    <nav className={tabbar({ orientation: "horizontal" }).root({ className: [props.classNames?.root, props.className] })} aria-label={label}>
      <TabbarList {...props} label={label} orientation="horizontal" />
    </nav>
  );
}

export function VerticalTabbar({ label = "Page sections", ...props }: TabbarProps) {
  return (
    <nav className={tabbar({ orientation: "vertical" }).root({ className: [props.classNames?.root, props.className] })} aria-label={label}>
      <TabbarList {...props} label={label} orientation="vertical" />
    </nav>
  );
}
