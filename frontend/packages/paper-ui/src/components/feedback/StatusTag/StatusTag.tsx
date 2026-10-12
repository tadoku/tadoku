export type StatusTagTone = "neutral" | "information" | "success" | "warning" | "danger";

export interface StatusTagProps {
  readonly tone?: StatusTagTone;
  /** Visible status text; tone colour only reinforces it. */
  readonly children: string;
  readonly className?: string;
}

export function StatusTag({ tone = "neutral", children, className }: StatusTagProps) {
  return (
    <span className={`paper-status-tag paper-status-tag--${tone}${className ? ` ${className}` : ""}`}>
      {children}
    </span>
  );
}
