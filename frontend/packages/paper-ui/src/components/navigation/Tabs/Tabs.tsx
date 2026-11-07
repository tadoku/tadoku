import * as stylex from "@stylexjs/stylex";
import { Tabs as BaseTabs } from "@base-ui/react/tabs";
import { sx } from "../../../foundations/props";
import { control } from "../../../foundations/recipes";
import { border, breakpoints, color, density, fontWeight, lineHeight, motion, opacity, zIndex } from "../../../tokens/tokens.stylex";
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


const styles = stylex.create({
  root: {
    color: color.textInk,
    display: { ['[data-orientation="vertical"]']: "grid", default: null, },
    gridTemplateColumns: { ['[data-orientation="vertical"]']: "max-content minmax(0, 1fr)", default: null, },
  },
  list: {
    alignItems: { ['[data-orientation="vertical"]']: "stretch", default: "end", },
    borderBlockEndColor: { default: color.ruleDefault, [breakpoints.forcedColors]: "CanvasText" },
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: { ['[data-orientation="vertical"]']: 0, default: border.static, },
    borderInlineEndColor: { default: color.ruleDefault, [breakpoints.forcedColors]: "CanvasText" },
    borderInlineEndStyle: "solid",
    borderInlineEndWidth: { ['[data-orientation="vertical"]']: border.static, default: 0, },
    display: "flex",
    flexDirection: { ['[data-orientation="vertical"]']: "column", default: null, },
    inlineSize: { ['[data-orientation="vertical"]']: "auto", default: "100%", },
    maxInlineSize: { ['[data-orientation="vertical"]']: "none", default: "100%", },
    overflowX: { ['[data-orientation="vertical"]']: "visible", default: "auto", },
    overflowY: { ['[data-orientation="vertical"]']: "visible", default: "hidden", },
  },
  tab: {
    borderRadius: 0,
    borderWidth: 0,
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    alignItems: "center",
    backgroundColor: { default: "transparent", [":not(:disabled):not([data-active]):hover"]: color.actionNeutralHover },
    borderBlockEndColor: {
      "[data-active]": { default: color.textLink, [breakpoints.forcedColors]: "Highlight" },
      default: "transparent",
    },
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: { ['[data-orientation="vertical"]']: 0, default: border.actionEdge, },
    borderInlineEndColor: {
      "[data-active]": { default: color.textLink, [breakpoints.forcedColors]: "Highlight" },
      default: "transparent",
    },
    borderInlineEndStyle: "solid",
    borderInlineEndWidth: { ['[data-orientation="vertical"]']: border.actionEdge, default: 0, },
    boxSizing: "border-box",
    color: {
      "[data-active]": color.textInk,
      default: color.textMuted,
      [breakpoints.forcedColors]: { "[data-disabled]": "GrayText", default: "CanvasText", ":disabled": "GrayText", },
      [":not(:disabled):not([data-active]):hover"]: color.textInk,
    },
    cursor: { "[data-disabled]": "default", default: "pointer", ":disabled": "default", },
    display: "inline-flex",
    flexBasis: 'auto',
    flexGrow: '0',
    flexShrink: '0',
    fontSize: density.buttonFontSize,
    fontWeight: fontWeight.semibold,
    justifyContent: { ['[data-orientation="vertical"]']: "start", default: "center", },
    lineHeight: lineHeight.none,
    minBlockSize: density.controlHeight,
    opacity: {
      "[data-disabled]": { default: opacity.disabled, [breakpoints.forcedColors]: opacity.visible },
      default: null,
      ":disabled": { default: opacity.disabled, [breakpoints.forcedColors]: opacity.visible },
    },
    outlineColor: {
      default: null,
      ":focus-visible": {
        default: color.focusRing,
        [breakpoints.forcedColors]: "Highlight",
      },
    },
    outlineOffset: { default: null, ":focus-visible": `calc(-1 * ${border.focusRing})` },
    outlineStyle: {
      default: null,
      ":focus-visible": {
        default: "solid",
        [breakpoints.forcedColors]: "solid",
      },
    },
    outlineWidth: {
      default: null,
      ":focus-visible": {
        default: border.focusRing,
        [breakpoints.forcedColors]: border.focusRing,
      },
    },
    position: "relative",
    textAlign: { ['[data-orientation="vertical"]']: "start", default: null, },
    transitionDuration: motion.quick,
    transitionProperty: { default: "background-color, border-color, color", [breakpoints.reducedMotion]: "none" },
    userSelect: "none",
    zIndex: { default: null, ":focus-visible": zIndex.raised },
  },
  panel: {
    paddingBlock: { ['[data-orientation="vertical"]']: 0, default: density.fieldPaddingInline, },
    paddingInline: { ['[data-orientation="vertical"]']: density.fieldPaddingInline, default: null, },
    backgroundColor: color.surfacePaper,
    color: color.textInk,
    display: { "[hidden]": "none", default: null, },
  },
});

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
      {...sx(className, styles.root)}
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
      {...sx(className, styles.list)}
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
      {...sx(className, control.inherit, styles.tab)}
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
      {...sx(className, styles.panel)}
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
