import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { fillIcon } from "../../foundations/IconSlot";
import type { StyleXProp } from "../../foundations/props";
import { color, size } from "../../tokens/tokens.stylex";

const styles = stylex.create({
  tile: {
    placeItems: "center",
    backgroundColor: color.actionSoft,
    blockSize: size[44],
    color: color.textLink,
    display: "grid",
    inlineSize: size[44],
  },
  icon: { blockSize: size[20], display: "block", inlineSize: size[20] },
});

/** A decorative square that leads a step or feature with its icon on the soft accent. */
export function IconTile({ children, xstyle }: { readonly children: ReactNode; readonly xstyle?: StyleXProp }) {
  return (
    <span aria-hidden="true" {...stylex.props(styles.tile, xstyle)}>
      <span {...stylex.props(styles.icon)}>{fillIcon(children)}</span>
    </span>
  );
}
