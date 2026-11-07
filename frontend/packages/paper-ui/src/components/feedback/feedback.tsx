import { Toast } from "@base-ui/react/toast";
import {
  type ButtonHTMLAttributes,
  type HTMLAttributes,
  type PropsWithChildren,
  type ReactNode,
} from "react";
import * as stylex from "@stylexjs/stylex";
import { XMarkIcon } from "../../icons";
import { buttonStyles } from "../actions/Button";
import { sx, type StyleXProp } from "../../foundations/props";
import { IconSlot } from "../../foundations/IconSlot";
import { a11y, accentRail, elevation as elevations, icon } from "../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  fluid,
  fontWeight,
  lineHeight,
  motion,
  radius,
  size,
  space,
  zIndex,
} from "../../tokens/tokens.stylex";

const spin = stylex.keyframes({ to: { transform: "rotate(1turn)" } });

const flash = stylex.create({
  root: {
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    gap: space[12],
    paddingBlock: space[14],
    alignItems: "start",
    backgroundColor: color.surfacePaper,
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) auto",
    paddingInlineEnd: space[16],
    paddingInlineStart: `calc(${space[16]} + ${border.accentRail})`,
    position: "relative",
    "::before": {
      insetBlock: `calc(-1 * ${border.static})`,
      content: '""',
      inlineSize: size[4],
      insetInlineStart: `calc(-1 * ${border.static})`,
      pointerEvents: "none",
      position: "absolute",
    },
  },
  withIcon: { gridTemplateColumns: "auto minmax(0, 1fr) auto" },
  information: { "::before": { backgroundColor: { default: color.statusInformation, [breakpoints.forcedColors]: "CanvasText" } } },
  success: { "::before": { backgroundColor: { default: color.statusSuccess, [breakpoints.forcedColors]: "CanvasText" } } },
  warning: { "::before": { backgroundColor: { default: color.statusWarning, [breakpoints.forcedColors]: "CanvasText" } } },
  danger: { "::before": { backgroundColor: { default: color.statusDanger, [breakpoints.forcedColors]: "CanvasText" } } },
  forced: { "::before": { forcedColorAdjust: { default: null, [breakpoints.forcedColors]: "none" } } },
  content: { gap: space[3], display: "grid", minInlineSize: 0, overflowWrap: "anywhere", },
  contentAlone: { gridColumnEnd: '-1',
 gridColumnStart: '1' },
  centered: { alignSelf: "center" },
});

const loading = stylex.create({
  root: { placeItems: "center", display: "inline-grid", },
  spinner: {
    borderColor: color.ruleDefault,
    borderRadius: radius.round,
    borderStyle: "solid",
    borderWidth: border.spinner,
    animationDuration: motion.spin,
    animationIterationCount: "infinite",
    animationName: { default: spin, [breakpoints.reducedMotion]: "none" },
    animationTimingFunction: motion.linear,
    blockSize: size[20],
    borderBlockStartColor: color.actionDefault,
    boxSizing: "border-box",
    inlineSize: size[20],
  },
  small: { blockSize: size[16], inlineSize: size[16] },
  default: {},
  large: { borderWidth: border.spinnerLarge, blockSize: size[32], inlineSize: size[32], },
});

const surfaceStyle = stylex.create({
  card: {    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
 boxSizing: "border-box",
    color: color.textInk,
  },
  default: { padding: space[16] },
  none: { padding: 0 },
  fluid: { padding: `clamp(${space[16]}, ${fluid[4]}, ${space[32]})`, overflowWrap: "anywhere", },
});

const group = stylex.create({
  root: { gap: density.inlineGap, display: "flex", flexWrap: "wrap", },
  start: {},
  end: { justifyContent: "end" },
});

const toast = stylex.create({
  viewport: {
    gap: space[8],
    display: "flex",
    flexDirection: "column",
    inlineSize: `min(${size[384]}, calc(100vw - ${space[32]}))`,
    insetBlockEnd: space[16],
    insetInlineEnd: space[16],
    outlineStyle: "none",
    position: "fixed",
    zIndex: zIndex.toast,
  },
  root: {
    padding: space[14],
    borderColor: color.ruleStrong,
    borderStyle: "solid",
    borderWidth: border.static,
    gap: space[12],
    backgroundColor: color.surfaceOverlay,
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) auto",
    opacity: { "[data-ending-style]": 0, "[data-starting-style]": 0, default: null, },
    transform: {
      "[data-ending-style]": `translateX(${space[8]})`,
      "[data-starting-style]": `translateX(${space[8]})`,
      default: null,
    },
    transitionDuration: motion.standard,
    transitionProperty: { default: "opacity, transform", [breakpoints.reducedMotion]: "none" },
  },
  content: { gap: space[3], display: "grid", },
  title: { margin: 0, fontSize: density.bodySize, fontWeight: fontWeight.semibold, lineHeight: lineHeight.snug, },
  description: { margin: 0 },
  close: {
    padding: 0,
    borderWidth: 0,
    alignSelf: "start",
    backgroundColor: "transparent",
    blockSize: size[32],
    color: color.textMuted,
    cursor: "pointer",
    inlineSize: size[32],
  },
});

export const FLASH_VARIANTS = ["information", "success", "warning", "danger"] as const;
export type FlashVariant = (typeof FLASH_VARIANTS)[number];

export interface FlashProps extends HTMLAttributes<HTMLDivElement> {
  readonly variant?: FlashVariant;
  readonly title?: string;
  readonly icon?: ReactNode;
  readonly action?: ReactNode;
  readonly visible?: boolean;
  readonly xstyle?: StyleXProp;
}

export function Flash({
  variant = "information",
  title,
  icon,
  action,
  visible = true,
  children,
  className,
  xstyle,
  ...props
}: FlashProps) {
  if (!visible) return null;
  return (
    <div
      {...props}
      data-variant={variant}
      {...sx(className, flash.root, Boolean(icon) && flash.withIcon, flash[variant], flash.forced, xstyle)}
      role={variant === "danger" ? "alert" : "status"}
    >
      {icon ? <IconSlot>{icon}</IconSlot> : null}
      <div {...stylex.props(flash.content, !icon && !action && flash.contentAlone, Boolean(action) && flash.centered)}>
        {title ? <strong>{title}</strong> : null}
        <div>{children}</div>
      </div>
      {action ? <div {...stylex.props(flash.centered)}>{action}</div> : null}
    </div>
  );
}

export interface LoadingProps extends HTMLAttributes<HTMLDivElement> {
  readonly label?: string;
  readonly size?: "small" | "default" | "large";
  readonly xstyle?: StyleXProp;
}

export function Loading({
  label = "Loading",
  size = "default",
  className,
  xstyle,
  ...props
}: LoadingProps) {
  return (
    <div {...props} {...sx(className, loading.root, xstyle)} role="status">
      <span {...stylex.props(loading.spinner, loading[size])} aria-hidden="true" />
      <span {...stylex.props(a11y.visuallyHidden)}>{label}</span>
    </div>
  );
}

export interface SurfaceProps extends HTMLAttributes<HTMLElement> {
  readonly as?: "div" | "article" | "section";
  readonly elevation?: "flat" | "floating" | "showcase";
  readonly accent?: boolean;
  /** `none` leaves padding to the content, such as a table or tab bar that runs edge to edge; `fluid` grows with the viewport. */
  readonly inset?: "default" | "none" | "fluid";
  readonly xstyle?: StyleXProp;
}

/** The surface card's appearance for other elements: `stylex.props(...surfaceStyles())`. */
export function surfaceStyles({
  elevation = "flat",
  accent = false,
  inset = "default",
}: Pick<SurfaceProps, "elevation" | "accent" | "inset"> = {}) {
  return [
    surfaceStyle.card,
    surfaceStyle[inset],
    elevations[elevation],
    accent && accentRail.onBorder,
    accent && accentRail.forcedColors,
  ] as const;
}

export function Surface({
  as: Element = "div",
  elevation = "flat",
  accent = false,
  inset = "default",
  className,
  xstyle,
  ...props
}: SurfaceProps) {
  return <Element {...props} {...sx(className, ...surfaceStyles({ elevation, accent, inset }), xstyle)} />;
}

export interface ButtonGroupAction {
  readonly id: string;
  readonly label: ReactNode;
  readonly href?: string;
  readonly onSelect?: () => void;
  readonly variant?: "default" | "outline" | "ghost" | "link" | "destructive";
  readonly disabled?: boolean;
  readonly visible?: boolean;
}

export interface ButtonGroupProps {
  readonly actions: readonly ButtonGroupAction[];
  readonly label?: string;
  readonly align?: "start" | "end";
}

export function ButtonGroup({ actions, label = "Actions", align = "start" }: ButtonGroupProps) {
  return (
    <div {...stylex.props(group.root, group[align])} role="group" aria-label={label}>
      {actions.filter((action) => action.visible !== false).map((action) => {
        const appearance = stylex.props(...buttonStyles({ variant: action.variant }));
        return action.href ? (
          <a
            key={action.id}
            href={action.href}
            {...appearance}
            aria-disabled={action.disabled || undefined}
            onClick={(event) => {
              if (action.disabled) event.preventDefault();
              else action.onSelect?.();
            }}
          >
            {action.label}
          </a>
        ) : (
          <button
            key={action.id}
            type="button"
            {...appearance}
            disabled={action.disabled}
            onClick={action.onSelect}
          >
            {action.label}
          </button>
        );
      })}
    </div>
  );
}

export interface ToastProviderProps extends PropsWithChildren {
  readonly timeout?: number;
  readonly limit?: number;
}

export interface ToastOptions {
  readonly id?: string;
  readonly title?: ReactNode;
  readonly description?: ReactNode;
  readonly type?: string;
  readonly timeout?: number;
  readonly priority?: "low" | "high";
  readonly onClose?: () => void;
  readonly onRemove?: () => void;
  readonly actionProps?: ButtonHTMLAttributes<HTMLButtonElement>;
}

export type ToastUpdateOptions = Omit<ToastOptions, "id">;

export interface ToastPromiseOptions<Value> {
  readonly loading: string | ToastUpdateOptions;
  readonly success:
    | string
    | ToastUpdateOptions
    | ((result: Value) => string | ToastUpdateOptions);
  readonly error:
    | string
    | ToastUpdateOptions
    | ((error: unknown) => string | ToastUpdateOptions);
}

export interface ToastManager {
  readonly add: (options: ToastOptions) => string;
  readonly close: (toastId?: string) => void;
  readonly update: (toastId: string, options: ToastUpdateOptions) => void;
  readonly promise: <Value>(
    promise: Promise<Value>,
    options: ToastPromiseOptions<Value>,
  ) => Promise<Value>;
}

function PaperToastViewport() {
  const manager = Toast.useToastManager();
  return (
    <Toast.Viewport {...stylex.props(toast.viewport)}>
      {manager.toasts.map((item) => (
        <Toast.Root key={item.id} toast={item} {...stylex.props(toast.root, elevations.floating)}>
          <Toast.Content {...stylex.props(toast.content)}>
            {item.title ? <Toast.Title {...stylex.props(toast.title)}>{item.title}</Toast.Title> : null}
            {item.description ? <Toast.Description {...stylex.props(toast.description)}>{item.description}</Toast.Description> : null}
          </Toast.Content>
          {item.actionProps ? <Toast.Action {...stylex.props(...buttonStyles({ variant: "outline" }))} {...item.actionProps} /> : null}
          <button
            type="button"
            {...stylex.props(toast.close)}
            aria-label="Dismiss notification"
            onClick={() => manager.close(item.id)}
          >
            <XMarkIcon {...stylex.props(icon.prominent)} aria-hidden="true" />
          </button>
        </Toast.Root>
      ))}
    </Toast.Viewport>
  );
}

export function ToastProvider({ children, timeout, limit }: ToastProviderProps) {
  return (
    <Toast.Provider timeout={timeout} limit={limit}>
      {children}
      <PaperToastViewport />
    </Toast.Provider>
  );
}

/** Migration alias for the legacy root export. Prefer ToastProvider. */
export const ToastContainer = ToastProvider;

export function useToast(): ToastManager {
  return Toast.useToastManager();
}
