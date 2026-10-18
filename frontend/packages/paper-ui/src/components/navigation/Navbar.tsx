import { Menu } from "@base-ui/react/menu";
import {
  useId,
  useRef,
  useState,
  useSyncExternalStore,
  type KeyboardEvent,
  type MouseEvent,
  type ReactNode,
} from "react";
import { Bars3Icon, ChevronDownIcon, XMarkIcon, iconClassName } from "../../icons";
import {
  disabledLinkClick,
  linkIsCurrent,
  renderNavigationLink,
  type NavigationItem,
  type NavigationLinkRenderer,
} from "./navigation";

function noSubscription() {
  return () => {};
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
  /** Optional controls or account utilities rendered at the end of the bar. */
  readonly actions?: ReactNode;
  /** Narrow-only account or utility content; call closeMenu after activating an action. */
  readonly mobileFooter?: (closeMenu: () => void) => ReactNode;
  /** Set to false when the application provides its own narrow navigation pattern. */
  readonly mobileNavigation?: boolean;
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
  readonly menuLabel?: string;
  readonly isLoading?: boolean;
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
  actions,
  mobileFooter,
  mobileNavigation = true,
  currentPath,
  renderLink,
  label = "Main navigation",
  menuLabel = "Menu",
  isLoading = false,
}: NavbarProps) {
  const menuId = `paper-navbar-${useId().replace(/:/gu, "")}`;
  const [mobileOpen, setMobileOpen] = useState(false);
  // False in server HTML and while hydrating it, where React would drop the trigger's clicks; true for client renders.
  const hydrated = useSyncExternalStore(noSubscription, () => true, () => false);
  const mobileTriggerRef = useRef<HTMLButtonElement>(null);

  const closeMobile = () => setMobileOpen(false);
  const onMobileClick = (event: MouseEvent<HTMLElement>) => {
    const link = (event.target as Element).closest("a[href]");
    if (link && event.currentTarget.contains(link) && link.getAttribute("aria-disabled") !== "true") closeMobile();
  };
  const onMobileKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Escape") return;
    setMobileOpen(false);
    mobileTriggerRef.current?.focus();
  };

  return (
    <nav className="paper-navbar" aria-label={label} onKeyDown={mobileNavigation && mobileOpen ? onMobileKeyDown : undefined} onClick={mobileNavigation && mobileOpen ? onMobileClick : undefined}>
      <div className="paper-navbar__inner">
        <div className="paper-navbar__brand">
          {renderNavigationLink(renderLink, { href: brandHref, children: brand })}
        </div>
        {navigation.length > 0 ? (
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
        ) : null}
        {actions ? <div className="paper-navbar__actions">{actions}</div> : null}
        {mobileNavigation ? (
          <button
            ref={mobileTriggerRef}
            type="button"
            className="paper-navbar__mobile-trigger"
            aria-controls={menuId}
            aria-expanded={mobileOpen}
            aria-label={mobileOpen ? `Close ${menuLabel.toLocaleLowerCase()}` : `Open ${menuLabel.toLocaleLowerCase()}`}
            disabled={!hydrated}
            onClick={() => setMobileOpen((open) => !open)}
          >
            {mobileOpen ? (
              <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
            ) : (
              <Bars3Icon className={iconClassName("prominent")} aria-hidden="true" />
            )}
          </button>
        ) : null}
      </div>
      {mobileNavigation ? (
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
          {mobileFooter?.(closeMobile)}
        </div>
      ) : null}
      {isLoading ? <div className="paper-navbar__loading" role="status"><span className="paper-sr-only">Loading navigation</span></div> : null}
    </nav>
  );
}
