import * as stylex from "@stylexjs/stylex";
import {
  forwardRef,
  useId,
  type ButtonHTMLAttributes,
  type ReactNode,
} from "react";
import { IconSlot } from "../../../foundations/IconSlot";
import { sx, type StyleXProp } from "../../../foundations/props";
import { a11y, control } from "../../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  decoration,
  density,
  fontWeight,
  lineHeight,
  motion,
  opacity,
  radius,
  size,
  space,
} from "../../../tokens/tokens.stylex";

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
}

const spin = stylex.keyframes({ to: { transform: "rotate(1turn)" } });

const styles = stylex.create({
  base: {
    borderColor: {
      default: color.actionDefault,
      [":active:not(:disabled)"]: color.actionActive,
      [":hover:not(:disabled)"]: color.actionHover,
    },
    borderRadius: radius.none,
    borderStyle: "solid",
    borderWidth: border.static,
    gap: space[8],
    paddingInline: density.buttonPaddingInline,
    textDecoration: "none",
    alignItems: "center",
    backgroundColor: {
      default: color.actionDefault,
      [":active:not(:disabled)"]: color.actionActive,
      [":hover:not(:disabled)"]: color.actionHover,
    },
    borderBlockEndColor: color.ruleActionEdge,
    borderBlockEndWidth: border.actionEdge,
    boxSizing: "border-box",
    color: color.actionText,
    cursor: { '[aria-disabled="true"]': "default", default: "pointer", ":disabled": "default", },
    display: "inline-flex",
    fontSize: density.buttonFontSize,
    fontWeight: fontWeight.semibold,
    justifyContent: "center",
    lineHeight: lineHeight.none,
    minBlockSize: density.controlHeight,
    opacity: { '[aria-disabled="true"]': opacity.disabled, default: null, ":disabled": opacity.disabled, },
    paddingBlockEnd: density.buttonPaddingBlockEnd,
    paddingBlockStart: density.buttonPaddingBlockStart,
    position: "relative",
    textAlign: "center",
    transform: { default: null, [":active:not(:disabled)"]: `translateY(${space[1]})` },
    transitionDuration: motion.quick,
    transitionProperty: { default: "background-color, border-color, color", [breakpoints.reducedMotion]: "none" },
    userSelect: "none",
  },
  outline: {
    borderColor: {
      "[data-popup-open]": color.ruleStrong,
      default: color.ruleFieldEdge,
      [":active:not(:disabled)"]: color.ruleStrong,
      [":hover:not(:disabled)"]: color.ruleStrong,
    },
    backgroundColor: {
      "[data-popup-open]": color.actionNeutralHover,
      default: color.surfacePaper,
      [":active:not(:disabled)"]: color.actionNeutralHover,
      [":hover:not(:disabled)"]: color.actionNeutralHover,
    },
    borderBlockEndColor: {
      "[data-popup-open]": color.ruleStrong,
      default: color.ruleFieldEdge,
      [":active:not(:disabled)"]: color.ruleStrong,
      [":hover:not(:disabled)"]: color.ruleStrong,
    },
    color: color.textInk,
  },
  ghost: {
    borderColor: "transparent",
    backgroundColor: {
      "[data-popup-open]": color.actionNeutralHover,
      default: "transparent",
      [":active:not(:disabled)"]: color.actionNeutralHover,
      [":hover:not(:disabled)"]: color.actionNeutralHover,
    },
    borderBlockEndColor: "transparent",
    color: color.textInk,
  },
  link: {
    borderWidth: 0,
    paddingInline: 0,
    textDecoration: "underline",
    backgroundColor: "transparent",
    borderBlockEndWidth: 0,
    color: color.textLink,
    minBlockSize: "auto",
    paddingBlockEnd: 0,
    paddingBlockStart: 0,
    textUnderlineOffset: decoration.offset,
  },
  destructive: {
    borderColor: {
      default: color.actionDestructive,
      [":active:not(:disabled)"]: color.actionDestructiveHover,
      [":hover:not(:disabled)"]: color.actionDestructiveHover,
    },
    backgroundColor: {
      default: color.actionDestructive,
      [":active:not(:disabled)"]: color.actionDestructiveHover,
      [":hover:not(:disabled)"]: color.actionDestructiveHover,
    },
    borderBlockEndColor: color.ruleDestructiveActionEdge,
    color: color.actionDestructiveText,
  },
  default: {},
  fullWidth: { inlineSize: "100%" },
  spinner: {
    borderColor: "currentColor",
    borderRadius: radius.round,
    borderStyle: "solid",
    borderWidth: border.spinner,
    flex: "none",
    animationDuration: motion.spin,
    animationIterationCount: "infinite",
    animationName: { default: spin, [breakpoints.reducedMotion]: "none" },
    animationTimingFunction: motion.linear,
    blockSize: size[20],
    borderInlineEndColor: "transparent",
    boxSizing: "border-box",
    display: "inline-grid",
    inlineSize: size[20],
  },
});

/** The button's appearance, for links and other elements that act as buttons: `stylex.props(...buttonStyles())`. */
export function buttonStyles({ variant = "default", fullWidth = false }: ButtonRecipeOptions = {}) {
  return [control.inherit, styles.base, styles[variant], fullWidth && styles.fullWidth, a11y.focusRing] as const;
}

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  readonly variant?: ButtonVariant;
  readonly fullWidth?: boolean;
  readonly loading?: boolean;
  /** Visually hidden status announced without replacing the accessible name. */
  readonly loadingLabel?: string;
  readonly leadingIcon?: ReactNode;
  readonly trailingIcon?: ReactNode;
  readonly xstyle?: StyleXProp;
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
    xstyle,
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
      data-variant={variant}
      {...sx(className, ...buttonStyles({ variant, fullWidth }), xstyle)}
    >
      {loading ? <span aria-hidden="true" {...stylex.props(styles.spinner)} /> : null}
      {!loading && leadingIcon ? <IconSlot>{leadingIcon}</IconSlot> : null}
      <span>{children}</span>
      {!loading && trailingIcon ? <IconSlot>{trailingIcon}</IconSlot> : null}
      {loading ? (
        <span {...stylex.props(a11y.visuallyHidden)} id={loadingStatusId} role="status">
          {loadingLabel}
        </span>
      ) : null}
    </button>
  );
});
