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
import { surface } from "../../foundations/recipes";
import { cn } from "../../foundations/variants";
import { navbar, navigationIcon, type NavbarSlot } from "./styles";


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
  readonly className?: string;
  /** Sits inside an application frame's header band, within the frame width and gutter. */
  readonly contained?: boolean;
  /** Classes merged over a part's own, for the application shell's layout. */
  readonly classNames?: Partial<Record<NavbarSlot, string>>;
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
  const slots = navbar();

  return (
    <Menu.Root>
      <Menu.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
        className={slots.trigger()}
      >
        <span>{item.label}</span>
        <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner className={slots.positioner()} sideOffset={6} align="end">
          <Menu.Popup className={cn(surface({ tone: "overlay", elevation: "floating" }), slots.dropdown())}>
            {item.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              if (link.disabled) {
                return (
                  <Menu.Item
                    key={link.id}
                    className={slots.item()}
                    disabled
                    label={link.label}
                  >
                    {link.icon ? <span className={navigationIcon} aria-hidden="true">{link.icon}</span> : null}
                    <span>{link.label}</span>
                  </Menu.Item>
                );
              }
              return (
                <Menu.LinkItem
                  key={link.id}
                  className={slots.item()}
                  label={link.label}
                  closeOnClick
                  aria-current={current ? "page" : undefined}
                  onClick={link.onSelect}
                  render={renderNavigationLink(renderLink, {
                    href: link.href,
                    children: (
                      <>
                        {link.icon ? <span className={navigationIcon} aria-hidden="true">{link.icon}</span> : null}
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
  className,
  contained = false,
  classNames = {},
}: NavbarProps) {
  const slots = navbar({ contained });
  const menuId = `paper-navbar-${useId().replace(/:/gu, "")}`;
  const [mobileOpen, setMobileOpen] = useState(false);
  // False in server HTML and while hydrating it, where React would drop the trigger's clicks and a link opens the panel
  // instead; true for client renders.
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
    <nav className={slots.root({ className: [classNames.root, className] })} aria-label={label} onKeyDown={mobileNavigation && mobileOpen ? onMobileKeyDown : undefined} onClick={mobileNavigation && mobileOpen ? onMobileClick : undefined}>
      <div className={slots.inner({ className: classNames.inner })}>
        <div className={slots.brand({ className: classNames.brand })}>
          {renderNavigationLink(renderLink, {
            href: brandHref,
            className: slots.brandLink({ className: classNames.brandLink }),
            children: brand,
          })}
        </div>
        {navigation.length > 0 ? (
          <div className={slots.links({ className: classNames.links })}>
            {navigation.map((item) => {
              if (item.type === "dropdown") {
                return <NavbarDropdownMenu key={item.id} item={item} currentPath={currentPath} renderLink={renderLink} />;
              }
              const current = linkIsCurrent(item, currentPath);
              return renderNavigationLink(renderLink, {
                href: item.href,
                className: slots.link({ className: classNames.link }),
                "aria-current": current ? "page" : undefined,
                "aria-disabled": item.disabled || undefined,
                tabIndex: item.disabled ? -1 : undefined,
                onClick: item.disabled ? disabledLinkClick : () => item.onSelect?.(),
                children: item.label,
              }, item.id);
            })}
          </div>
        ) : null}
        {actions ? (
          <div className={navbar({ contained, alone: navigation.length === 0 }).actions({ className: classNames.actions })}>{actions}</div>
        ) : null}
        {mobileNavigation && !hydrated ? (
          <a
            href={`#${menuId}`}
            className={slots.mobileFallback({ className: classNames.mobileTrigger })}
            aria-label={`Open ${menuLabel.toLocaleLowerCase()}`}
          >
            <Bars3Icon className={iconClassName("prominent")} aria-hidden="true" />
          </a>
        ) : null}
        {mobileNavigation && hydrated ? (
          <button
            ref={mobileTriggerRef}
            type="button"
            className={slots.mobileTrigger({ className: classNames.mobileTrigger })}
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
        ) : null}
      </div>
      {mobileNavigation ? (
        <div
          id={menuId}
          className={navbar({ contained, open: mobileOpen, fallback: !hydrated }).panel({ className: classNames.panel })}
          data-open={mobileOpen ? "true" : "false"}
          // Paper hides [hidden] with !important, so the server panel relies on its classes, which :target overrides.
          hidden={hydrated ? !mobileOpen : undefined}
        >
          {navigation.map((item) =>
          item.type === "link" ? (
            <div key={item.id}>
              {renderNavigationLink(renderLink, {
                href: item.href,
                className: slots.mobileLink({ className: classNames.mobileLink }),
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
              <h2 id={`${menuId}-${item.id}`} className={slots.groupTitle()}>{item.label}</h2>
              {item.links.map((link) =>
                renderNavigationLink(renderLink, {
                  href: link.href,
                  className: slots.mobileLink({ className: classNames.mobileLink }),
                  "aria-current": linkIsCurrent(link, currentPath) ? "page" : undefined,
                  "aria-disabled": link.disabled || undefined,
                  tabIndex: link.disabled ? -1 : undefined,
                  onClick: link.disabled ? disabledLinkClick : () => {
                    link.onSelect?.();
                    closeMobile();
                  },
                  children: (
                    <>
                      {link.icon ? <span className={navigationIcon} aria-hidden="true">{link.icon}</span> : null}
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
      {isLoading ? <div className={slots.loading()} role="status"><span className="sr-only">Loading navigation</span></div> : null}
    </nav>
  );
}
