import { Toast } from "@base-ui/react/toast";
import {
  type ButtonHTMLAttributes,
  type HTMLAttributes,
  type PropsWithChildren,
  type ReactNode,
} from "react";
import { XMarkIcon, iconClassName } from "../../icons";
import { buttonClassName } from "../actions/Button";

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
  return (
    <div
      {...props}
      className={`paper-flash paper-flash--${variant}${className ? ` ${className}` : ""}`}
      role={variant === "danger" ? "alert" : "status"}
    >
      {icon ? <span className="paper-flash__icon" aria-hidden="true">{icon}</span> : null}
      <div className="paper-flash__content">
        {title ? <strong>{title}</strong> : null}
        <div>{children}</div>
      </div>
      {action ? <div className="paper-flash__action">{action}</div> : null}
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
  return (
    <div {...props} className={`paper-loading paper-loading--${size}${className ? ` ${className}` : ""}`} role="status">
      <span className="paper-loading__spinner" aria-hidden="true" />
      <span className="paper-visually-hidden">{label}</span>
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
  return [
    "paper-surface-card",
    `paper-elevation-${elevation}`,
    accent ? "paper-accent-rail" : "",
    className ?? "",
  ].filter(Boolean).join(" ");
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
    <div className={`paper-button-group paper-button-group--${align}`} role="group" aria-label={label}>
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
  return (
    <Toast.Viewport className="paper-toast-viewport">
      {manager.toasts.map((toast) => (
        <Toast.Root key={toast.id} toast={toast} className="paper-toast paper-elevation-floating">
          <Toast.Content className="paper-toast__content">
            {toast.title ? <Toast.Title className="paper-toast__title">{toast.title}</Toast.Title> : null}
            {toast.description ? <Toast.Description>{toast.description}</Toast.Description> : null}
          </Toast.Content>
          {toast.actionProps ? <Toast.Action className={buttonClassName({ variant: "outline" })} {...toast.actionProps} /> : null}
          <button
            type="button"
            className="paper-toast__close"
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
