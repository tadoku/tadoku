import {
  forwardRef,
  useId,
  type ButtonHTMLAttributes,
  type ReactNode,
} from "react";
import { tv } from "../../../foundations/variants";

export const BUTTON_VARIANTS = [
  "default",
  "outline",
  "ghost",
  "link",
  "destructive",
  "destructive-ghost",
] as const;

export type ButtonVariant = (typeof BUTTON_VARIANTS)[number];

export interface ButtonRecipeOptions {
  readonly variant?: ButtonVariant;
  readonly fullWidth?: boolean;
  readonly loading?: boolean;
  readonly className?: string;
}

export const button = tv({
  slots: {
    base: [
      "paper-button relative inline-flex min-h-control select-none items-center justify-center gap-2",
      "px-button-x pt-button-top pb-button-bottom text-button font-semibold leading-none text-center no-underline",
      "rounded-none border border-b-edge border-button-shade bg-action text-on-action",
      "cursor-pointer transition-colors duration-quick motion-reduce:transition-none",
      "not-disabled:hover:bg-action-hover not-disabled:active:bg-action-active",
      "not-disabled:active:translate-y-px",
      "disabled:cursor-default disabled:bg-rule disabled:text-muted aria-disabled:cursor-default aria-disabled:bg-rule aria-disabled:text-muted",
    ],
    icon: "paper-button__icon inline-grid size-icon flex-none place-items-center *:size-full",
    spinner: [
      "paper-button__spinner inline-grid size-icon flex-none place-items-center",
      "rounded-full border-edge border-current border-e-transparent animate-spin motion-reduce:animate-none",
    ],
    label: "paper-button__label",
    status: "paper-visually-hidden sr-only",
  },
  variants: {
    variant: {
      default: { base: "paper-button--default" },
      outline: {
        base: [
          "paper-button--outline border-button-edge bg-raised text-ink",
          "not-disabled:hover:border-button-edge-hover not-disabled:hover:bg-raised-hover",
          "not-disabled:active:border-button-edge-hover not-disabled:active:bg-raised-hover",
          "data-popup-open:border-button-edge-hover data-popup-open:bg-raised-hover",
        ],
      },
      ghost: {
        base: [
          "paper-button--ghost border-transparent bg-transparent text-ink",
          "not-disabled:hover:border-transparent not-disabled:hover:bg-neutral-hover not-disabled:hover:text-ink",
          "not-disabled:active:border-transparent not-disabled:active:bg-neutral-hover not-disabled:active:text-ink",
          "data-popup-open:border-transparent data-popup-open:bg-neutral-hover data-popup-open:text-ink",
        ],
      },
      link: {
        base: [
          "paper-button--link min-h-auto p-0 border-0 bg-transparent text-link underline underline-offset-link",
          "not-disabled:hover:bg-transparent not-disabled:active:bg-transparent",
        ],
      },
      destructive: {
        base: [
          "paper-button--destructive bg-destructive text-on-destructive",
          "not-disabled:hover:bg-destructive-hover not-disabled:active:bg-destructive-hover",
        ],
      },
      "destructive-ghost": {
        base: [
          "paper-button--destructive-ghost border-transparent bg-transparent text-danger",
          "not-disabled:hover:border-transparent not-disabled:hover:bg-destructive-soft",
          "not-disabled:active:border-transparent not-disabled:active:bg-destructive-soft",
        ],
      },
    },
    fullWidth: { true: { base: "paper-button--full-width w-full" } },
    loading: { true: { base: "paper-button--loading" } },
  },
  defaultVariants: { variant: "default" },
});

export function buttonClassName({
  variant = "default",
  fullWidth = false,
  loading = false,
  className,
}: ButtonRecipeOptions = {}): string {
  return button({ variant, fullWidth, loading }).base({ className });
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
  const slots = button();
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
      {loading ? <span className={slots.spinner()} aria-hidden="true" /> : null}
      {!loading && leadingIcon ? (
        <span className={slots.icon()} aria-hidden="true">
          {leadingIcon}
        </span>
      ) : null}
      <span className={slots.label()}>{children}</span>
      {!loading && trailingIcon ? (
        <span className={slots.icon()} aria-hidden="true">
          {trailingIcon}
        </span>
      ) : null}
      {loading ? (
        <span className={slots.status()} id={loadingStatusId} role="status">
          {loadingLabel}
        </span>
      ) : null}
    </button>
  );
});
