import {
  cloneElement,
  type AnchorHTMLAttributes,
  type Key,
  type MouseEvent,
  type ReactElement,
  type ReactNode,
} from "react";

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

export function renderNavigationLink(
  renderLink: NavigationLinkRenderer | undefined,
  props: NavigationLinkProps,
  key?: Key,
) {
  const element = (renderLink ?? defaultRenderLink)(props);
  return key === undefined ? element : cloneElement(element, { key });
}

export function disabledLinkClick(event: MouseEvent<HTMLAnchorElement>) {
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

export function linkIsCurrent(item: NavigationItem, currentPath?: string) {
  return item.current ?? (currentPath !== undefined && item.href === currentPath);
}
