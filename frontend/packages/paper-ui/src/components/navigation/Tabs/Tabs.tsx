import { Tabs as BaseTabs } from "@base-ui/react/tabs";
import {
  forwardRef,
  type ButtonHTMLAttributes,
  type ForwardedRef,
  type HTMLAttributes,
  type KeyboardEvent,
} from "react";
import { tv } from "../../../foundations/variants";

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

const tabs = tv({
  slots: {
    root: [
      "paper-tabs group/tabs text-ink",
      "data-[orientation=vertical]:grid data-[orientation=vertical]:grid-cols-tabs-vertical",
    ],
    list: [
      "paper-tabs__list flex w-full max-w-full items-end overflow-x-auto overflow-y-hidden border-b border-rule",
      "group-data-[orientation=vertical]/tabs:w-auto group-data-[orientation=vertical]/tabs:max-w-none",
      "group-data-[orientation=vertical]/tabs:flex-col group-data-[orientation=vertical]/tabs:items-stretch",
      "group-data-[orientation=vertical]/tabs:overflow-visible group-data-[orientation=vertical]/tabs:border-b-0",
      "group-data-[orientation=vertical]/tabs:border-e forced-colors:border-forced-ink",
    ],
    tab: [
      "paper-tabs__tab relative inline-flex min-h-control flex-none cursor-pointer select-none items-center justify-center",
      "rounded-none border-0 border-b-edge border-b-transparent bg-transparent px-field-x py-field-y",
      "font-semibold leading-none text-muted transition-colors duration-quick motion-reduce:transition-none",
      "not-disabled:not-data-active:hover:bg-neutral-hover not-disabled:not-data-active:hover:text-ink",
      "data-active:border-b-link data-active:bg-transparent data-active:text-ink",
      "focus-visible:z-lift focus-visible:outline-ring focus-visible:outline-focus focus-visible:-outline-offset-ring",
      "disabled:cursor-default disabled:text-muted disabled:opacity-disabled",
      "data-disabled:cursor-default data-disabled:text-muted data-disabled:opacity-disabled",
      "group-data-[orientation=vertical]/tabs:justify-start group-data-[orientation=vertical]/tabs:text-start",
      "group-data-[orientation=vertical]/tabs:border-b-0 group-data-[orientation=vertical]/tabs:border-e-edge",
      "group-data-[orientation=vertical]/tabs:border-e-transparent",
      "group-data-[orientation=vertical]/tabs:data-active:border-e-link",
      "compact:min-h-9 compact:px-2.5 compact:py-1.5 compact:text-sm",
      "forced-colors:text-forced-ink forced-colors:data-active:border-b-forced-highlight",
      "forced-colors:focus-visible:outline-forced-highlight",
      "forced-colors:group-data-[orientation=vertical]/tabs:data-active:border-e-forced-highlight",
      "forced-colors:disabled:text-forced-muted forced-colors:disabled:opacity-full",
      "forced-colors:data-disabled:text-forced-muted forced-colors:data-disabled:opacity-full",
    ],
    panel: [
      "paper-tabs__panel bg-paper py-field-x text-ink",
      "group-data-[orientation=vertical]/tabs:px-field-x group-data-[orientation=vertical]/tabs:py-0",
    ],
  },
});

const slots = tabs();

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
      className={slots.root({ className })}
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
      className={slots.list({ className })}
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
      className={slots.tab({ className })}
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
      className={slots.panel({ className })}
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
