import {
  forwardRef,
  useId,
  type ButtonHTMLAttributes,
  type ReactNode,
} from "react";

export const BUTTON_VARIANTS = [
  "default",
  "outline",
  "ghost",
  "link",
  "destructive",
] as const;

export type ButtonVariant = (typeof BUTTON_VARIANTS)[number];

export interface ButtonRecipeOptions {
  readonly variant?: ButtonVariant;
  readonly fullWidth?: boolean;
  readonly loading?: boolean;
  readonly className?: string;
}

export function buttonClassName({
  variant = "default",
  fullWidth = false,
  loading = false,
  className,
}: ButtonRecipeOptions = {}): string {
  return [
    "paper-button",
    `paper-button--${variant}`,
    fullWidth ? "paper-button--full-width" : "",
    loading ? "paper-button--loading" : "",
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");
}

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  readonly variant?: ButtonVariant;
  readonly fullWidth?: boolean;
  readonly loading?: boolean;
  /** Visually hidden status announced without replacing the accessible name. */
  readonly loadingLabel?: string;
  readonly leadingIcon?: ReactNode;
  readonly trailingIcon?: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  {
    variant = "default",
    fullWidth = false,
    loading = false,
    loadingLabel = "Working",
    leadingIcon,
    trailingIcon,
    className,
    children,
    disabled,
    type = "button",
    "aria-describedby": describedBy,
    "aria-label": ariaLabel,
    ...props
  },
  ref,
) {
  const generatedId = useId();
  const loadingStatusId = `paper-button-status-${generatedId.replace(/:/gu, "")}`;
  return (
    <button
      {...props}
      ref={ref}
      type={type}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      aria-label={
        ariaLabel ?? (loading && typeof children === "string" ? children : undefined)
      }
      aria-describedby={
        [describedBy, loading ? loadingStatusId : undefined].filter(Boolean).join(" ") ||
        undefined
      }
      className={buttonClassName({ variant, fullWidth, loading, className })}
    >
      {loading ? <span className="paper-button__spinner" aria-hidden="true" /> : null}
      {!loading && leadingIcon ? (
        <span className="paper-button__icon" aria-hidden="true">
          {leadingIcon}
        </span>
      ) : null}
      <span className="paper-button__label">{children}</span>
      {!loading && trailingIcon ? (
        <span className="paper-button__icon" aria-hidden="true">
          {trailingIcon}
        </span>
      ) : null}
      {loading ? (
        <span className="paper-visually-hidden" id={loadingStatusId} role="status">
          {loadingLabel}
        </span>
      ) : null}
    </button>
  );
});
