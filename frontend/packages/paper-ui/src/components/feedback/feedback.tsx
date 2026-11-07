import { Toast } from "@base-ui/react/toast";
import {
  type ButtonHTMLAttributes,
  type HTMLAttributes,
  type PropsWithChildren,
  type ReactNode,
} from "react";
import { XMarkIcon, iconClassName } from "../../icons";
import { buttonClassName } from "../actions/Button";
import { accentRail, surface } from "../../foundations/recipes";
import { cn, tv } from "../../foundations/variants";

const flash = tv({
  slots: {
    root: [
      "paper-flash relative grid items-start gap-3 border border-rule bg-paper py-3.6 pe-4 ps-flash-start",
      "before:pointer-events-none before:absolute before:-inset-y-px before:-start-px before:w-1",
      "forced-colors:before:bg-forced-ink forced-colors:before:forced-color-adjust-none",
    ],
    icon: "paper-flash__icon grid size-icon place-items-center",
    content: "paper-flash__content grid min-w-0 gap-0.8 wrap-anywhere",
    action: "paper-flash__action self-center",
  },
  variants: {
    variant: {
      information: { root: "paper-flash--information before:bg-information" },
      success: { root: "paper-flash--success before:bg-success" },
      warning: { root: "paper-flash--warning before:bg-warning" },
      danger: { root: "paper-flash--danger before:bg-danger" },
    },
    icon: { true: { root: "grid-cols-icon-content-end" }, false: { root: "grid-cols-content-end" } },
    alone: { true: { content: "col-span-full" } },
    action: { true: { content: "self-center" } },
  },
});

const loading = tv({
  slots: {
    root: "paper-loading inline-grid place-items-center",
    spinner: [
      "paper-loading__spinner size-icon rounded-full border-edge border-rule border-t-action",
      "animate-spin motion-reduce:animate-none",
    ],
  },
  variants: {
    size: {
      small: { root: "paper-loading--small", spinner: "size-icon-sm" },
      default: { root: "paper-loading--default" },
      large: { root: "paper-loading--large", spinner: "size-8 border-rail" },
    },
  },
});

const toast = tv({
  slots: {
    viewport: "paper-toast-viewport fixed bottom-4 end-4 z-toast flex w-toast flex-col gap-2 outline-none",
    root: [
      "paper-toast grid grid-cols-content-end gap-3 p-3.4",
      "transition-enter duration-standard motion-reduce:transition-none",
      "data-starting-style:translate-x-2 data-starting-style:opacity-0",
      "data-ending-style:translate-x-2 data-ending-style:opacity-0",
    ],
    content: "paper-toast__content grid gap-0.8",
    title: "paper-toast__title m-0 text-body leading-125 font-semibold",
    description: "m-0",
    close: "paper-toast__close size-8 cursor-pointer self-start border-0 bg-transparent text-muted",
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
}

export function Flash({
  variant = "information",
  title,
  icon,
  action,
  visible = true,
  children,
  className,
  ...props
}: FlashProps) {
  if (!visible) return null;
  const slots = flash({ variant, icon: Boolean(icon), alone: !icon && !action, action: Boolean(action) });
  return (
    <div
      {...props}
      className={slots.root({ className })}
      role={variant === "danger" ? "alert" : "status"}
    >
      {icon ? <span className={slots.icon()} aria-hidden="true">{icon}</span> : null}
      <div className={slots.content()}>
        {title ? <strong>{title}</strong> : null}
        <div>{children}</div>
      </div>
      {action ? <div className={slots.action()}>{action}</div> : null}
    </div>
  );
}

export interface LoadingProps extends HTMLAttributes<HTMLDivElement> {
  readonly label?: string;
  readonly size?: "small" | "default" | "large";
}

export function Loading({
  label = "Loading",
  size = "default",
  className,
  ...props
}: LoadingProps) {
  const slots = loading({ size });
  return (
    <div {...props} className={slots.root({ className })} role="status">
      <span className={slots.spinner()} aria-hidden="true" />
      <span className="paper-visually-hidden sr-only">{label}</span>
    </div>
  );
}

export interface SurfaceProps extends HTMLAttributes<HTMLElement> {
  readonly as?: "div" | "article" | "section";
  readonly elevation?: "flat" | "floating" | "showcase";
  readonly accent?: boolean;
}

export function surfaceClassName({
  elevation = "flat",
  accent = false,
  className,
}: Pick<SurfaceProps, "elevation" | "accent" | "className"> = {}): string {
  return cn(
    "paper-surface-card p-4",
    `paper-elevation-${elevation}`,
    surface({ elevation, className: "text-inherit" }),
    accent && ["paper-accent-rail", accentRail({ bordered: true })],
    className,
  );
}

export function Surface({
  as: Element = "div",
  elevation = "flat",
  accent = false,
  className,
  ...props
}: SurfaceProps) {
  return <Element {...props} className={surfaceClassName({ elevation, accent, className })} />;
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
    <div
      className={cn("paper-button-group flex flex-wrap gap-inline", align === "end" ? "paper-button-group--end justify-end" : "paper-button-group--start")}
      role="group"
      aria-label={label}
    >
      {actions.filter((action) => action.visible !== false).map((action) => {
        const className = buttonClassName({ variant: action.variant });
        return action.href ? (
          <a
            key={action.id}
            href={action.href}
            className={className}
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
            className={className}
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
  const slots = toast();
  return (
    <Toast.Viewport className={slots.viewport()}>
      {manager.toasts.map((toast) => (
        <Toast.Root key={toast.id} toast={toast} className={cn(surface({ tone: "overlay", elevation: "floating" }), slots.root())}>
          <Toast.Content className={slots.content()}>
            {toast.title ? <Toast.Title className={slots.title()}>{toast.title}</Toast.Title> : null}
            {toast.description ? <Toast.Description className={slots.description()}>{toast.description}</Toast.Description> : null}
          </Toast.Content>
          {toast.actionProps ? <Toast.Action className={buttonClassName({ variant: "outline" })} {...toast.actionProps} /> : null}
          <button
            type="button"
            className={slots.close()}
            aria-label="Dismiss notification"
            onClick={() => manager.close(toast.id)}
          >
            <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
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
