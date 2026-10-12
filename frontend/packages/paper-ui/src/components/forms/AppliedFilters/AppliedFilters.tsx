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
      className={className ? `paper-applied-filters ${className}` : "paper-applied-filters"}
      role="group"
      aria-label={label}
    >
      <p className="paper-applied-filters__list">{filters.join(" · ")}</p>
      <Button variant="link" onClick={onClear}>{clearLabel}</Button>
    </div>
  );
}
