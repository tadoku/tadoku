import { tv } from "../../../foundations/variants";

export type StatusTagTone = "neutral" | "information" | "success" | "warning" | "danger";

export interface StatusTagProps {
  readonly tone?: StatusTagTone;
  /** Visible status text; tone colour only reinforces it. */
  readonly children: string;
  readonly className?: string;
}

const statusTag = tv({
  base: [
    "paper-status-tag inline-flex items-center border border-current px-2 py-0.5",
    "font-body text-fine leading-150 font-semibold whitespace-nowrap",
  ],
  variants: {
    tone: {
      neutral: "paper-status-tag--neutral border-field-edge text-ink",
      information: "paper-status-tag--information text-information",
      success: "paper-status-tag--success text-success",
      warning: "paper-status-tag--warning text-warning",
      danger: "paper-status-tag--danger text-danger",
    },
  },
});

export function StatusTag({ tone = "neutral", children, className }: StatusTagProps) {
  return <span className={statusTag({ tone, className })}>{children}</span>;
}
