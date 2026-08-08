import { Menu } from "@base-ui/react/menu";
import {
  useId,
  cloneElement,
  useRef,
  useState,
  type AnchorHTMLAttributes,
  type KeyboardEvent,
  type Key,
  type MouseEvent,
  type ReactElement,
  type ReactNode,
} from "react";
import {
  Bars3Icon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  EllipsisHorizontalIcon,
  XMarkIcon,
  iconClassName,
} from "../../icons";

export interface NavigationLinkProps
  extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, "href"> {
  readonly href: string;
}

export type NavigationLinkRenderer = (
  props: NavigationLinkProps,
) => ReactElement;

function defaultRenderLink(props: NavigationLinkProps) {
  return <a {...props} />;
}

function renderNavigationLink(
  renderLink: NavigationLinkRenderer | undefined,
  props: NavigationLinkProps,
  key?: Key,
) {
  const element = (renderLink ?? defaultRenderLink)(props);
  return key === undefined ? element : cloneElement(element, { key });
}

function disabledLinkClick(event: MouseEvent<HTMLAnchorElement>) {
  event.preventDefault();
}

export interface NavigationItem {
  readonly id: string;
  readonly label: string;
  readonly href: string;
  readonly current?: boolean;
  readonly disabled?: boolean;
  readonly icon?: ReactNode;
  readonly onSelect?: () => void;
}

export interface NavigationDropdown {
  readonly type: "dropdown";
  readonly id: string;
  readonly label: string;
  readonly links: readonly NavigationItem[];
}

export interface NavigationDirectLink extends NavigationItem {
  readonly type: "link";
}

export type NavbarItem = NavigationDirectLink | NavigationDropdown;

export interface NavbarProps {
  readonly navigation: readonly NavbarItem[];
  readonly brand: ReactNode;
  readonly brandHref: string;
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
  readonly menuLabel?: string;
  readonly isLoading?: boolean;
}

function linkIsCurrent(item: NavigationItem, currentPath?: string) {
  return item.current ?? (currentPath !== undefined && item.href === currentPath);
}

function NavbarDropdownMenu({
  item,
  currentPath,
  renderLink,
}: {
  readonly item: NavigationDropdown;
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
}) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);

  return (
    <Menu.Root>
      <Menu.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
        className="paper-navbar__dropdown-trigger"
      >
        <span>{item.label}</span>
        <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner className="paper-navbar__positioner" sideOffset={6} align="end">
          <Menu.Popup className="paper-navbar__dropdown paper-elevation-floating">
            {item.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              if (link.disabled) {
                return (
                  <Menu.Item
                    key={link.id}
                    className="paper-navbar__dropdown-item"
                    disabled
                    label={link.label}
                  >
                    {link.icon ? <span className="paper-navigation__icon" aria-hidden="true">{link.icon}</span> : null}
                    <span>{link.label}</span>
                  </Menu.Item>
                );
              }
              return (
                <Menu.LinkItem
                  key={link.id}
                  className="paper-navbar__dropdown-item"
                  label={link.label}
                  closeOnClick
                  aria-current={current ? "page" : undefined}
                  onClick={link.onSelect}
                  render={renderNavigationLink(renderLink, {
                    href: link.href,
                    children: (
                      <>
                        {link.icon ? <span className="paper-navigation__icon" aria-hidden="true">{link.icon}</span> : null}
                        <span>{link.label}</span>
                      </>
                    ),
                  })}
                />
              );
            })}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}

export function Navbar({
  navigation,
  brand,
  brandHref,
  currentPath,
  renderLink,
  label = "Main navigation",
  menuLabel = "Menu",
  isLoading = false,
}: NavbarProps) {
  const menuId = `paper-navbar-${useId().replace(/:/gu, "")}`;
  const [mobileOpen, setMobileOpen] = useState(false);
  const mobileTriggerRef = useRef<HTMLButtonElement>(null);

  const closeMobile = () => setMobileOpen(false);
  const onMobileKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Escape") return;
    setMobileOpen(false);
    mobileTriggerRef.current?.focus();
  };

  return (
    <nav className="paper-navbar" aria-label={label} onKeyDown={mobileOpen ? onMobileKeyDown : undefined}>
      <div className="paper-navbar__inner">
        <div className="paper-navbar__brand">
          {renderNavigationLink(renderLink, { href: brandHref, children: brand })}
        </div>
        <button
          ref={mobileTriggerRef}
          type="button"
          className="paper-navbar__mobile-trigger"
          aria-controls={menuId}
          aria-expanded={mobileOpen}
          aria-label={mobileOpen ? `Close ${menuLabel.toLocaleLowerCase()}` : `Open ${menuLabel.toLocaleLowerCase()}`}
          onClick={() => setMobileOpen((open) => !open)}
        >
          {mobileOpen ? (
            <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
          ) : (
            <Bars3Icon className={iconClassName("prominent")} aria-hidden="true" />
          )}
        </button>
        <div className="paper-navbar__desktop-links">
          {navigation.map((item) => {
            if (item.type === "dropdown") {
              return <NavbarDropdownMenu key={item.id} item={item} currentPath={currentPath} renderLink={renderLink} />;
            }
            const current = linkIsCurrent(item, currentPath);
            return renderNavigationLink(renderLink, {
              href: item.href,
              className: "paper-navbar__link",
              "aria-current": current ? "page" : undefined,
              "aria-disabled": item.disabled || undefined,
              tabIndex: item.disabled ? -1 : undefined,
              onClick: item.disabled ? disabledLinkClick : () => item.onSelect?.(),
              children: item.label,
            }, item.id);
          })}
        </div>
      </div>
      <div
        id={menuId}
        className="paper-navbar__mobile-panel"
        data-open={mobileOpen ? "true" : "false"}
        hidden={!mobileOpen}
      >
        {navigation.map((item) =>
          item.type === "link" ? (
            <div key={item.id}>
              {renderNavigationLink(renderLink, {
                href: item.href,
                className: "paper-navbar__mobile-link",
                "aria-current": linkIsCurrent(item, currentPath) ? "page" : undefined,
                "aria-disabled": item.disabled || undefined,
                tabIndex: item.disabled ? -1 : undefined,
                onClick: item.disabled ? disabledLinkClick : () => {
                  item.onSelect?.();
                  closeMobile();
                },
                children: item.label,
              })}
            </div>
          ) : (
            <section className="paper-navbar__mobile-group" key={item.id} aria-labelledby={`${menuId}-${item.id}`}>
              <h2 id={`${menuId}-${item.id}`}>{item.label}</h2>
              {item.links.map((link) =>
                renderNavigationLink(renderLink, {
                  href: link.href,
                  className: "paper-navbar__mobile-link",
                  "aria-current": linkIsCurrent(link, currentPath) ? "page" : undefined,
                  "aria-disabled": link.disabled || undefined,
                  tabIndex: link.disabled ? -1 : undefined,
                  onClick: link.disabled ? disabledLinkClick : () => {
                    link.onSelect?.();
                    closeMobile();
                  },
                  children: (
                    <>
                      {link.icon ? <span className="paper-navigation__icon" aria-hidden="true">{link.icon}</span> : null}
                      <span>{link.label}</span>
                    </>
                  ),
                }, link.id),
              )}
            </section>
          ),
        )}
      </div>
      {isLoading ? <div className="paper-navbar__loading" role="status"><span className="paper-sr-only">Loading navigation</span></div> : null}
    </nav>
  );
}

export interface SidebarSection {
  readonly id: string;
  readonly title: string;
  readonly links: readonly NavigationItem[];
}

export interface SidebarProps {
  readonly sections: readonly SidebarSection[];
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
}

export function Sidebar({
  sections,
  currentPath,
  renderLink,
  label = "Section navigation",
}: SidebarProps) {
  return (
    <nav className="paper-sidebar" aria-label={label}>
      {sections.map((section) => (
        <section key={section.id} className="paper-sidebar__section" aria-labelledby={`paper-sidebar-${section.id}`}>
          <h2 id={`paper-sidebar-${section.id}`} className="paper-sidebar__title">{section.title}</h2>
          <ul className="paper-sidebar__list">
            {section.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              return (
                <li key={link.id}>
                  {renderNavigationLink(renderLink, {
                    href: link.href,
                    className: "paper-sidebar__link",
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
        </section>
      ))}
    </nav>
  );
}

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
  return (
    <ul className={`paper-tabbar__list paper-tabbar__list--${orientation}`}>
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

type PaginationToken = number | "start-gap" | "end-gap";

function paginationTokens(totalPages: number, currentPage: number, siblingCount: number): PaginationToken[] {
  const visible = new Set([1, totalPages]);
  for (let page = currentPage - siblingCount; page <= currentPage + siblingCount; page += 1) {
    if (page >= 1 && page <= totalPages) visible.add(page);
  }
  const pages = [...visible].sort((left, right) => left - right);
  const tokens: PaginationToken[] = [];
  pages.forEach((page, index) => {
    const previous = pages[index - 1];
    if (previous !== undefined && page - previous > 1) {
      tokens.push(previous === 1 ? "start-gap" : "end-gap");
    }
    tokens.push(page);
  });
  return tokens;
}

export interface PaginationProps {
  readonly totalPages: number;
  readonly currentPage: number;
  readonly getHref?: (page: number) => string;
  readonly onPageChange?: (page: number) => void;
  readonly siblingCount?: number;
  readonly label?: string;
  readonly renderLink?: NavigationLinkRenderer;
}

export function Pagination({
  totalPages,
  currentPage,
  getHref,
  onPageChange,
  siblingCount = 2,
  label = "Pagination",
  renderLink,
}: PaginationProps) {
  if (!Number.isInteger(totalPages) || totalPages < 1) {
    throw new RangeError("Pagination totalPages must be a positive integer.");
  }
  if (!Number.isInteger(currentPage) || currentPage < 1 || currentPage > totalPages) {
    throw new RangeError("Pagination currentPage must be within totalPages.");
  }
  if (!Number.isInteger(siblingCount) || siblingCount < 0) {
    throw new RangeError("Pagination siblingCount must be a non-negative integer.");
  }
  if (!getHref && !onPageChange) {
    throw new Error("Pagination requires getHref or onPageChange.");
  }

  const activate = (page: number) => (event: MouseEvent<HTMLAnchorElement>) => {
    if (!onPageChange) return;
    event.preventDefault();
    onPageChange(page);
  };
  const tokens = paginationTokens(totalPages, currentPage, siblingCount);

  const control = (page: number, content: ReactNode, controlLabel: string, disabled: boolean) => {
    if (disabled) {
      return <span className="paper-pagination__control" aria-disabled="true">{content}</span>;
    }
    if (getHref) {
      return renderNavigationLink(renderLink, {
        href: getHref(page),
        className: "paper-pagination__control",
        "aria-label": controlLabel,
        onClick: activate(page),
        children: content,
      });
    }
    return <button type="button" className="paper-pagination__control" aria-label={controlLabel} onClick={() => onPageChange?.(page)}>{content}</button>;
  };

  return (
    <nav className="paper-pagination" aria-label={label}>
      <ul className="paper-pagination__list">
        <li>{control(currentPage - 1, <><ChevronLeftIcon className={iconClassName("compact")} aria-hidden="true" /><span>Previous</span></>, "Previous page", currentPage === 1)}</li>
        {tokens.map((token) =>
          typeof token === "string" ? (
            <li key={token} className="paper-pagination__gap" aria-hidden="true"><EllipsisHorizontalIcon className={iconClassName("default")} /></li>
          ) : (
            <li key={token}>
              {token === currentPage ? (
                <span className="paper-pagination__page" aria-current="page" aria-label={`Page ${token}`}>{token}</span>
              ) : getHref ? renderNavigationLink(renderLink, {
                href: getHref(token),
                className: "paper-pagination__page",
                "aria-label": `Page ${token}`,
                onClick: activate(token),
                children: token,
              }) : (
                <button type="button" className="paper-pagination__page" aria-label={`Page ${token}`} onClick={() => onPageChange?.(token)}>{token}</button>
              )}
            </li>
          ),
        )}
        <li>{control(currentPage + 1, <><span>Next</span><ChevronRightIcon className={iconClassName("compact")} aria-hidden="true" /></>, "Next page", currentPage === totalPages)}</li>
      </ul>
    </nav>
  );
}
