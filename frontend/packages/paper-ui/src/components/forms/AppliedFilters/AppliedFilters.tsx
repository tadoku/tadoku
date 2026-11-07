import * as stylex from "@stylexjs/stylex";
import { Button } from "../../actions/Button";
import { sx, type StyleXProp } from "../../../foundations/props";
import { fontSize, lineHeight, space } from "../../../tokens/tokens.stylex";

export interface AppliedFiltersProps {
  /** Readable labels for the filters currently narrowing the results. */
  readonly filters: readonly string[];
  /** Resets every filter. Return focus to the filter toggle or first field. */
  readonly onClear: () => void;
  readonly label?: string;
  readonly clearLabel?: string;
  readonly className?: string;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  root: {
    alignItems: "center",
    columnGap: space[16],
    display: "flex",
    flexWrap: "wrap",
    fontSize: fontSize[13],
    lineHeight: lineHeight.normal,
    minInlineSize: 0,
    rowGap: space[8],
  },
  list: { margin: 0, minInlineSize: 0, overflowWrap: "anywhere" },
  clear: { flex: "none", fontSize: "inherit" },
});

export function AppliedFilters({
  filters,
  onClear,
  label = "Applied filters",
  clearLabel = "Clear all",
  className,
  xstyle,
}: AppliedFiltersProps) {
  if (filters.length === 0) return null;
  return (
    <div {...sx(className, styles.root, xstyle)} role="group" aria-label={label}>
      <p {...stylex.props(styles.list)}>{filters.join(" · ")}</p>
      <Button variant="link" onClick={onClear} xstyle={styles.clear}>{clearLabel}</Button>
    </div>
  );
}
