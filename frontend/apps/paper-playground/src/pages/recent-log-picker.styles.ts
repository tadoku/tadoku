import * as stylex from '@stylexjs/stylex'
import { border, breakpoints, color, fontSize, lineHeight, size, space, fluid } from 'paper-ui/tokens.stylex'

export const recentLogPickerStyles = stylex.create({
  recentLogPicker__search: {
    marginBlockEnd: space[12],
    marginBlockStart: space[4],
    position: "relative",
  },
  recentLogPicker__list: {
    scrollbarGutter: "stable",
    margin: 0,
    padding: 0,
    listStyle: "none",
    overscrollBehavior: "contain",
    blockSize: { "[data-scrollable]": `min(${size[352]}, ${fluid.vh55})`, default: null, },
    fontSize: fontSize[14],
    maxBlockSize: `min(${size[352]}, ${fluid.vh55})`,
    overflowY: "auto",
  },
  recentLogPicker__copy: {
    minWidth: 0,
  },
  recentLogPicker__entryTitle: {
    alignItems: "baseline",
    columnGap: space[12],
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
    rowGap: space[3],
  },
  recentLogPicker__footer: {
    gap: { default: space[12], [breakpoints.phone]: space[8] },
    paddingBlock: space[8],
    paddingInline: 0,
    alignItems: "center",
    color: color.textMuted,
    display: "flex",
    fontSize: fontSize[12],
    justifyContent: "space-between",
    lineHeight: lineHeight.normal,
    borderTopColor: color.ruleSubtle,
    borderTopStyle: "solid",
    borderTopWidth: border.static,
  },
  recentLogPicker__empty: {
    margin: 0,
    paddingBlock: space[16],
    paddingInline: 0,
    color: color.textMuted,
    fontSize: fontSize[14],
  },
  recentLogPicker__row: {},
})

