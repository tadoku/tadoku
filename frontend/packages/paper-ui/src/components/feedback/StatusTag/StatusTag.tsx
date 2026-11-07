import * as stylex from "@stylexjs/stylex";
import { sx, type StyleXProp } from "../../../foundations/props";
import { border, color, font, fontSize, fontWeight, lineHeight, space } from "../../../tokens/tokens.stylex";

export type StatusTagTone = "neutral" | "information" | "success" | "warning" | "danger";

export interface StatusTagProps {
  readonly tone?: StatusTagTone;
  /** Visible status text; tone colour only reinforces it. */
  readonly children: string;
  readonly className?: string;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  root: {
    borderColor: "currentColor",
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: space[2],
    paddingInline: space[8],
    alignItems: "center",
    display: "inline-flex",
    fontFamily: font.body,
    fontSize: fontSize[13],
    fontWeight: fontWeight.semibold,
    lineHeight: lineHeight.normal,
    whiteSpace: "nowrap",
  },
  neutral: { borderColor: color.ruleFieldEdge, color: color.textInk },
  information: { color: color.statusInformation },
  success: { color: color.statusSuccess },
  warning: { color: color.statusWarning },
  danger: { color: color.statusDanger },
});

export function StatusTag({ tone = "neutral", children, className, xstyle }: StatusTagProps) {
  return (
    <span data-tone={tone} {...sx(className, styles.root, styles[tone], xstyle)}>
      {children}
    </span>
  );
}
