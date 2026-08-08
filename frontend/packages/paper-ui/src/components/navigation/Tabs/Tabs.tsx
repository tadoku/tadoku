import { Tabs as BaseTabs } from "@base-ui/react/tabs";
import {
  forwardRef,
  type ButtonHTMLAttributes,
  type ForwardedRef,
  type HTMLAttributes,
  type KeyboardEvent,
} from "react";

export type TabsValue = string;
export type TabsOrientation = "horizontal" | "vertical";

export interface TabsRootProps
  extends Omit<HTMLAttributes<HTMLDivElement>, "defaultValue"> {
  readonly value?: TabsValue;
  readonly defaultValue?: TabsValue;
  readonly orientation?: TabsOrientation;
  readonly onValueChange?: (value: TabsValue) => void;
}

export interface TabsListProps extends HTMLAttributes<HTMLDivElement> {
  /** Select the newly focused tab during keyboard navigation. Defaults to true. */
  readonly activateOnFocus?: boolean;
  /** Wrap focus from the final tab to the first and vice versa. Defaults to true. */
  readonly loopFocus?: boolean;
}

export interface TabsTabProps
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "value"> {
  readonly value: TabsValue;
}

export interface TabsPanelProps
  extends Omit<HTMLAttributes<HTMLDivElement>, "value"> {
  readonly value: TabsValue;
  /** Keep the panel in the DOM while inactive. Defaults to true. */
  readonly keepMounted?: boolean;
}

function classes(base: string, className?: string): string {
  return [base, className].filter(Boolean).join(" ");
}

function moveTabFocus(
  event: KeyboardEvent<HTMLDivElement>,
  loopFocus: boolean,
): void {
  const tabs = Array.from(
    event.currentTarget.querySelectorAll<HTMLElement>(
      '[role="tab"]:not([aria-disabled="true"])',
    ),
  );
  const currentIndex = tabs.indexOf(event.target as HTMLElement);
  if (currentIndex < 0 || tabs.length === 0) return;

  let nextIndex: number | undefined;
  const vertical = event.currentTarget.dataset.orientation === "vertical";
  if (event.key === "Home") nextIndex = 0;
  if (event.key === "End") nextIndex = tabs.length - 1;
  if (event.key === (vertical ? "ArrowDown" : "ArrowRight")) {
    nextIndex = currentIndex + 1;
  }
  if (event.key === (vertical ? "ArrowUp" : "ArrowLeft")) {
    nextIndex = currentIndex - 1;
  }
  if (nextIndex === undefined) return;

  if (loopFocus) nextIndex = (nextIndex + tabs.length) % tabs.length;
  else nextIndex = Math.max(0, Math.min(nextIndex, tabs.length - 1));

  event.preventDefault();
  event.stopPropagation();
  tabs[nextIndex]?.focus();
}

export const TabsRoot = forwardRef<HTMLDivElement, TabsRootProps>(function TabsRoot(
  { className, onValueChange, ...props },
  ref,
) {
  return (
    <BaseTabs.Root
      ref={ref}
      className={classes("paper-tabs", className)}
      onValueChange={(nextValue) => {
        if (typeof nextValue === "string") onValueChange?.(nextValue);
      }}
      {...props}
    />
  );
});

export const TabsList = forwardRef<HTMLDivElement, TabsListProps>(function TabsList(
  {
    activateOnFocus = true,
    className,
    loopFocus = true,
    onKeyDownCapture,
    ...props
  },
  ref,
) {
  return (
    <BaseTabs.List
      ref={ref}
      activateOnFocus={activateOnFocus}
      loopFocus={loopFocus}
      className={classes("paper-tabs__list", className)}
      onKeyDownCapture={(event) => {
        onKeyDownCapture?.(event);
        if (!event.defaultPrevented) moveTabFocus(event, loopFocus);
      }}
      {...props}
    />
  );
});

export const TabsTab = forwardRef<HTMLButtonElement, TabsTabProps>(function TabsTab(
  { className, type = "button", ...props },
  ref,
) {
  return (
    <BaseTabs.Tab
      ref={ref as ForwardedRef<HTMLElement>}
      type={type}
      className={classes("paper-tabs__tab", className)}
      {...props}
    />
  );
});

export const TabsPanel = forwardRef<HTMLDivElement, TabsPanelProps>(function TabsPanel(
  { className, keepMounted = true, value, ...props },
  ref,
) {
  return (
    <BaseTabs.Panel
      ref={ref}
      className={classes("paper-tabs__panel", className)}
      data-value={String(value)}
      keepMounted={keepMounted}
      value={value}
      {...props}
    />
  );
});

/** Accessible content tabs. Linked page navigation should use Tabbar instead. */
export const Tabs = {
  Root: TabsRoot,
  List: TabsList,
  Tab: TabsTab,
  Panel: TabsPanel,
} as const;
