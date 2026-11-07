import { cn } from "../../../foundations/variants";
import { Button } from "../../actions/Button";

export interface AppliedFiltersProps {
  /** Readable labels for the filters currently narrowing the results. */
  readonly filters: readonly string[];
  /** Resets every filter. Return focus to the filter toggle or first field. */
  readonly onClear: () => void;
  readonly label?: string;
  readonly clearLabel?: string;
  readonly className?: string;
}

export function AppliedFilters({
  filters,
  onClear,
  label = "Applied filters",
  clearLabel = "Clear all",
  className,
}: AppliedFiltersProps) {
  if (filters.length === 0) return null;
  return (
    <div
      className={cn(
        "paper-applied-filters flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2 text-fine leading-150",
        className,
      )}
      role="group"
      aria-label={label}
    >
      <p className="paper-applied-filters__list m-0 min-w-0 wrap-anywhere">{filters.join(" · ")}</p>
      <Button variant="link" className="flex-none text-em" onClick={onClear}>{clearLabel}</Button>
    </div>
  );
}
