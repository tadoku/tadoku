import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { sx, type StyleXProp } from "../../../foundations/props";
import { border, color, fluid, fontSize, fontWeight, lineHeight, size, space, type } from "../../../tokens/tokens.stylex";

export interface StatGridItem {
  readonly label: ReactNode;
  readonly value: ReactNode;
  /** Supporting context read with the value, such as "12 active days". */
  readonly detail?: ReactNode;
}

export interface StatGridProps {
  readonly stats: readonly StatGridItem[];
  readonly className?: string;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  root: {
    margin: 0,
    columnGap: space[24],
    display: "grid",
    gridTemplateColumns: `repeat(auto-fit, minmax(min(100%, ${size[88]}), 1fr))`,
    minInlineSize: 0,
    rowGap: space[16],
  },
  /** Each stat spans two shared rows, so values align even when a label wraps. */
  item: {
    gridRow: "span 2",
    borderInlineStartColor: color.ruleDefault,
    borderInlineStartStyle: "solid",
    borderInlineStartWidth: border.static,
    display: "grid",
    gridTemplateRows: "subgrid",
    minInlineSize: 0,
    paddingInlineStart: `clamp(${space[8]}, ${fluid[2]}, ${space[16]})`,
    rowGap: space[4],
  },
  label: {
    alignSelf: "start",
    color: color.textMuted,
    fontSize: fontSize[14],
    lineHeight: lineHeight.normal,
    overflowWrap: "anywhere",
  },
  value: {
    margin: 0,
    color: color.textInk,
    fontSize: type.sectionSize,
    fontVariantNumeric: "tabular-nums",
    fontWeight: fontWeight.bold,
    lineHeight: lineHeight.statValue,
    overflowWrap: "anywhere",
  },
  detail: {
    color: color.textMuted,
    display: "block",
    fontSize: fontSize[13],
    fontWeight: fontWeight.regular,
    lineHeight: lineHeight.normal,
    marginBlockStart: space[4],
  },
});

export function StatGrid({ stats, className, xstyle }: StatGridProps) {
  return (
    <dl {...sx(className, styles.root, xstyle)}>
      {stats.map((stat, index) => (
        <div key={index} {...stylex.props(styles.item)} data-stat="">
          <dt {...stylex.props(styles.label)}>{stat.label}</dt>
          <dd {...stylex.props(styles.value)}>
            {stat.value}
            {stat.detail ? <span {...stylex.props(styles.detail)}>{stat.detail}</span> : null}
          </dd>
        </div>
      ))}
    </dl>
  );
}
