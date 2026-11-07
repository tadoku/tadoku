import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { sx, type StyleXProp } from "../../../foundations/props";
import { border, color, fontSize, lineHeight, size, space, type } from "../../../tokens/tokens.stylex";

export interface DescriptionListItem {
  readonly label: ReactNode;
  readonly value: ReactNode;
}

export interface DescriptionListProps {
  readonly items: readonly DescriptionListItem[];
  /** `stacked` puts each label above its value; `inline` puts them on one ruled row. */
  readonly layout?: "stacked" | "inline";
  readonly className?: string;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  root: { margin: 0, fontSize: fontSize[14], lineHeight: lineHeight.description, minInlineSize: 0, },
  stacked: { gap: space[16], display: "grid", },
  inline: {},
  item: { minInlineSize: 0 },
  /** Inline rows keep each value near its label instead of across the full width. */
  inlineItem: {
    alignItems: "baseline",
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: { default: border.static, ":last-child": 0 },
    columnGap: space[16],
    display: "grid",
    gridTemplateColumns: `minmax(${size[96]}, ${size[160]}) minmax(0, 1fr)`,
    paddingBlockEnd: { default: space[12], ":last-child": 0 },
    paddingBlockStart: { default: space[12], ":first-child": 0 },
    rowGap: space[4],
  },
  label: { color: color.textMuted, overflowWrap: "anywhere" },
  stackedLabel: { fontSize: type.metadataSize, lineHeight: lineHeight.compact, marginBlockEnd: space[4] },
  value: { margin: 0, fontVariantNumeric: "tabular-nums", overflowWrap: "anywhere", textAlign: "start", },
});

export function DescriptionList({ items, layout = "stacked", className, xstyle }: DescriptionListProps) {
  const stacked = layout === "stacked";
  return (
    <dl {...sx(className, styles.root, styles[layout], xstyle)}>
      {items.map((item, index) => (
        <div key={index} {...stylex.props(styles.item, !stacked && styles.inlineItem)}>
          <dt {...stylex.props(styles.label, stacked && styles.stackedLabel)}>{item.label}</dt>
          <dd {...stylex.props(styles.value)}>{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
