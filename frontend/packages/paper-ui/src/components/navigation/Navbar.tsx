import * as stylex from "@stylexjs/stylex";
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
import { Bars3Icon, ChevronDownIcon, XMarkIcon } from "../../icons";
import { IconSlot } from "../../foundations/IconSlot";
import type { StyleXProp } from "../../foundations/props";
import { a11y, control, elevation, icon } from "../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  fontWeight,
  motion,
  opacity,
  ratio,
  size,
  space,
  type,
  zIndex,
} from "../../tokens/tokens.stylex";
import {
  disabledLinkClick,
  linkIsCurrent,
  renderNavigationLink,
  type NavigationItem,
  type NavigationLinkRenderer,
} from "./navigation";

const loadingSlide = stylex.keyframes({
  from: { transform: "translateX(-100%)" },
  to: { transform: "translateX(350%)" },
});

const styles = stylex.create({
  nav: {
    backgroundColor: color.surfacePaper,
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    position: "relative",
    zIndex: zIndex.navbar,
  },
  inner: {
    gap: density.inlineGap,
    paddingInline: density.fieldPaddingInline,
    alignItems: "center",
    display: "flex",
    justifyContent: "space-between",
    minBlockSize: density.controlHeight,
  },
  brand: { alignItems: "center", display: "flex" },
  brandLink: {
    textDecoration: "none",
    alignItems: "center",
    color: color.textInk,
    display: "inline-flex",
    fontWeight: fontWeight.bold,
  },
  desktopLinks: {
    gap: space[4],
    alignItems: "stretch",
    display: { default: "flex", [breakpoints.md]: "none" },
    marginInlineStart: "auto",
  },
  actions: { gap: density.inlineGap, alignItems: "center", display: "flex", },
  actionsAfterBrand: { marginInlineStart: "auto" },
  link: {
    borderWidth: 0,
    gap: density.inlineGap,
    paddingInline: density.fieldPaddingInline,
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: { default: "transparent", ':hover:not([aria-disabled="true"])': color.actionNeutralHover },
    borderBlockEndWidth: 0,
    boxShadow: { '[aria-current="page"]': `inset 0 calc(-1 * ${border.indicator}) 0 ${color.textLink}`, default: null, },
    boxSizing: "border-box",
    color: color.textInk,
    cursor: { '[aria-disabled="true"]': "default", default: "pointer", },
    display: "inline-flex",
    fontWeight: fontWeight.semibold,
    minBlockSize: density.controlHeight,
    opacity: { '[aria-disabled="true"]': opacity.inactive, default: null, },
  },
  trigger: { backgroundColor: { default: "transparent", ":hover": color.actionNeutralHover } },
  mobileLink: { inlineSize: "100%" },
  mobileTrigger: {    padding: 0,
    borderWidth: 0,
    flex: "none",
    placeItems: "center",
    backgroundColor: "transparent",
    blockSize: size[44],
    color: color.textInk,
    cursor: { default: "pointer", ":disabled": "progress" },
    display: { default: "none", [breakpoints.md]: "grid" },
 fontFamily: "inherit",
 fontSize: "inherit",
    inlineSize: size[44],
    minBlockSize: size[44],
    minInlineSize: size[44],
  },
  positioner: { outlineStyle: "none", zIndex: zIndex.popover },
  dropdown: {
    borderColor: color.ruleStrong,
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: space[6],
    backgroundColor: color.surfaceOverlay,
    minInlineSize: size[208],
    opacity: { "[data-ending-style]": 0, "[data-starting-style]": 0, default: null, },
    transform: { "[data-ending-style]": "scale(0.98)", "[data-starting-style]": "scale(0.98)", default: null, },
    transformOrigin: "var(--transform-origin)",
    transitionDuration: motion.quick,
    transitionProperty: { default: "opacity, transform", [breakpoints.reducedMotion]: "none" },
  },
  dropdownItem: {
    gap: density.inlineGap,
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: { "[data-highlighted]": color.actionDefault, default: null, },
    boxSizing: "border-box",
    color: { "[data-highlighted]": color.actionText, default: color.textInk, },
    cursor: "default",
    display: "flex",
    minBlockSize: density.controlHeight,
    opacity: { "[data-disabled]": opacity.inactive, default: null, },
    outlineStyle: "none",
  },
  mobilePanel: {
    padding: space[8],
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "none",
  },
  mobilePanelOpen: { display: { default: "none", [breakpoints.md]: "block" } },
  mobileGroupTitle: {
    marginInline: density.fieldPaddingInline,
    color: color.textMuted,
    fontSize: type.metadataSize,
    fontWeight: fontWeight.bold,
    lineHeight: type.metadataLineHeight,
    marginBlockEnd: space[4],
    marginBlockStart: space[16],
    textTransform: "uppercase",
  },
  loading: {
    insetInline: 0,
    overflow: "hidden",
    backgroundColor: color.ruleSubtle,
    blockSize: size[2],
    insetBlockEnd: `calc(-1 * ${border.static})`,
    position: "absolute",
    "::after": {
      animationDuration: motion.progress,
      animationIterationCount: "infinite",
      animationName: { default: loadingSlide, [breakpoints.reducedMotion]: "none" },
      animationTimingFunction: motion.linear,
      backgroundColor: color.actionDefault,
      blockSize: "100%",
      content: '""',
      display: "block",
      inlineSize: { default: ratio.progress, [breakpoints.reducedMotion]: "100%" },
    },
  },
});

/** The look of a full-width row in the phone menu, for actions an application adds through `mobileFooter`. */
export const navbarMobileLinkStyles = [control.inherit, styles.link, styles.mobileLink, a11y.focusRing] as const;

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
  /** Layout for the bar, such as a maximum width and gutters. */
  readonly xstyle?: StyleXProp;
  readonly innerXstyle?: StyleXProp;
  readonly brandXstyle?: StyleXProp;
  readonly brandLinkXstyle?: StyleXProp;
  readonly actionsXstyle?: StyleXProp;
  readonly mobilePanelXstyle?: StyleXProp;
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
        {...stylex.props(control.inherit, styles.link, styles.trigger, a11y.focusRing)}
      >
        <span>{item.label}</span>
        <ChevronDownIcon {...stylex.props(icon.compact)} aria-hidden="true" />
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner {...stylex.props(styles.positioner)} sideOffset={6} align="end">
          <Menu.Popup {...stylex.props(styles.dropdown, elevation.floating)}>
            {item.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              if (link.disabled) {
                return (
                  <Menu.Item
                    key={link.id}
                    {...stylex.props(styles.dropdownItem)}
                    disabled
                    label={link.label}
                  >
                    {link.icon ? <IconSlot>{link.icon}</IconSlot> : null}
                    <span>{link.label}</span>
                  </Menu.Item>
                );
              }
              return (
                <Menu.LinkItem
                  key={link.id}
                  {...stylex.props(styles.dropdownItem)}
                  label={link.label}
                  closeOnClick
                  aria-current={current ? "page" : undefined}
                  onClick={link.onSelect}
                  render={renderNavigationLink(renderLink, {
                    href: link.href,
                    children: (
                      <>
                        {link.icon ? <IconSlot>{link.icon}</IconSlot> : null}
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
  xstyle,
  innerXstyle,
  brandXstyle,
  brandLinkXstyle,
  actionsXstyle,
  mobilePanelXstyle,
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
    <nav {...stylex.props(styles.nav, xstyle)} aria-label={label} onKeyDown={mobileNavigation && mobileOpen ? onMobileKeyDown : undefined} onClick={mobileNavigation && mobileOpen ? onMobileClick : undefined}>
      <div {...stylex.props(styles.inner, innerXstyle)}>
        <div {...stylex.props(styles.brand, brandXstyle)}>
          {renderNavigationLink(renderLink, {
            href: brandHref,
            ...stylex.props(styles.brandLink, a11y.focusRing, brandLinkXstyle),
            children: brand,
          })}
        </div>
        {navigation.length > 0 ? (
          <div {...stylex.props(styles.desktopLinks)}>
            {navigation.map((item) => {
              if (item.type === "dropdown") {
                return <NavbarDropdownMenu key={item.id} item={item} currentPath={currentPath} renderLink={renderLink} />;
              }
              const current = linkIsCurrent(item, currentPath);
              return renderNavigationLink(renderLink, {
                href: item.href,
                ...stylex.props(styles.link, a11y.focusRing),
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
          <div {...stylex.props(styles.actions, navigation.length === 0 && styles.actionsAfterBrand, actionsXstyle)}>
            {actions}
          </div>
        ) : null}
        {mobileNavigation ? (
          <button
            ref={mobileTriggerRef}
            type="button"
            {...stylex.props(styles.mobileTrigger, a11y.focusRing)}
            data-navbar-menu=""
            aria-controls={menuId}
            aria-expanded={mobileOpen}
            aria-label={mobileOpen ? `Close ${menuLabel.toLocaleLowerCase()}` : `Open ${menuLabel.toLocaleLowerCase()}`}
            disabled={!hydrated}
            onClick={() => setMobileOpen((open) => !open)}
          >
            {mobileOpen ? (
              <XMarkIcon {...stylex.props(icon.prominent)} aria-hidden="true" />
            ) : (
              <Bars3Icon {...stylex.props(icon.prominent)} aria-hidden="true" />
            )}
          </button>
        ) : null}
      </div>
      {mobileNavigation ? (
        <div
          id={menuId}
          {...stylex.props(styles.mobilePanel, mobileOpen && styles.mobilePanelOpen, mobilePanelXstyle)}
          data-open={mobileOpen ? "true" : "false"}
          hidden={!mobileOpen}
        >
          {navigation.map((item) =>
          item.type === "link" ? (
            <div key={item.id}>
              {renderNavigationLink(renderLink, {
                href: item.href,
                ...stylex.props(...navbarMobileLinkStyles),
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
            <section key={item.id} aria-labelledby={`${menuId}-${item.id}`}>
              <h2 {...stylex.props(styles.mobileGroupTitle)} id={`${menuId}-${item.id}`}>{item.label}</h2>
              {item.links.map((link) =>
                renderNavigationLink(renderLink, {
                  href: link.href,
                  ...stylex.props(...navbarMobileLinkStyles),
                  "aria-current": linkIsCurrent(link, currentPath) ? "page" : undefined,
                  "aria-disabled": link.disabled || undefined,
                  tabIndex: link.disabled ? -1 : undefined,
                  onClick: link.disabled ? disabledLinkClick : () => {
                    link.onSelect?.();
                    closeMobile();
                  },
                  children: (
                    <>
                      {link.icon ? <IconSlot>{link.icon}</IconSlot> : null}
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
      {isLoading ? (
        <div {...stylex.props(styles.loading)} role="status">
          <span {...stylex.props(a11y.visuallyHidden)}>Loading navigation</span>
        </div>
      ) : null}
    </nav>
  );
}
