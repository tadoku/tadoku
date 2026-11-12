import { useEffect, useRef } from "react";
import { XMarkIcon } from "../../../icons";
import { tv } from "../../../foundations/variants";
import { Button } from "../../actions/Button";

const appliedFilters = tv({
  slots: {
    root: "paper-applied-filters flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2 text-fine leading-150",
    list: "paper-applied-filters__list m-0 min-w-0 wrap-anywhere",
    chips: "paper-applied-filters__chips m-0 flex min-w-0 list-none flex-wrap gap-1.5 p-0",
    chip: [
      "paper-applied-filters__chip inline-flex min-h-chip-target cursor-pointer items-center gap-1 border border-action",
      "bg-action-soft px-2 py-0.5 text-fine font-semibold text-ink wrap-anywhere hover:bg-current-row-hover",
      "focus-visible:outline-ring focus-visible:outline-offset-ring focus-visible:outline-focus",
      "max-tablet:min-h-8",
    ],
    remove: "paper-applied-filters__remove size-icon-sm flex-none text-muted",
    clear: "flex-none text-em",
    count: "paper-applied-filters__count m-0 ms-auto text-muted tabular-nums",
  },
});

const slots = appliedFilters();

export interface RemovableFilter {
  /** Unique among the applied filters, such as "tag:audiobook". */
  readonly id: string;
  readonly label: string;
  readonly onRemove: () => void;
}

interface SharedProps {
  /** Resets every filter. Return focus to the filter toggle or first field. */
  readonly onClear: () => void;
  readonly label?: string;
  readonly clearLabel?: string;
  readonly className?: string;
}

export interface TextAppliedFiltersProps extends SharedProps {
  readonly variant?: "text";
  /** Readable labels for the filters currently narrowing the results. */
  readonly filters: readonly string[];
}

export interface RemovableAppliedFiltersProps extends SharedProps {
  readonly variant: "removable";
  /** Each applied value, rendered as a chip that removes itself. */
  readonly filters: readonly RemovableFilter[];
  /** The narrowed result, such as "83 logs of 127", at the end of the row. */
  readonly resultLabel?: string;
}

export type AppliedFiltersProps = TextAppliedFiltersProps | RemovableAppliedFiltersProps;

export function AppliedFilters(props: AppliedFiltersProps) {
  const { onClear, label = "Applied filters", clearLabel = "Clear all", className } = props;
  if (props.filters.length === 0) return null;
  return (
    <div className={slots.root({ className })} role="group" aria-label={label}>
      {props.variant === "removable" ? (
        <RemovableChips filters={props.filters} />
      ) : (
        <p className={slots.list()}>{props.filters.join(" · ")}</p>
      )}
      <Button variant="link" className={slots.clear()} onClick={onClear}>{clearLabel}</Button>
      {props.variant === "removable" && props.resultLabel ? (
        <p className={slots.count()}>{props.resultLabel}</p>
      ) : null}
    </div>
  );
}

function RemovableChips({ filters }: { readonly filters: readonly RemovableFilter[] }) {
  const list = useRef<HTMLUListElement>(null);
  const removedAt = useRef<number | null>(null);

  useEffect(() => {
    if (removedAt.current === null) return;
    const chips = list.current?.querySelectorAll("button");
    chips?.[Math.min(removedAt.current, chips.length - 1)]?.focus();
    removedAt.current = null;
  }, [filters]);

  return (
    <ul ref={list} className={slots.chips()}>
      {filters.map((filter, index) => (
        <li key={filter.id}>
          <button
            type="button"
            className={slots.chip()}
            aria-label={`Remove ${filter.label}`}
            onClick={() => {
              removedAt.current = index;
              filter.onRemove();
            }}
          >
            {filter.label}
            <XMarkIcon className={slots.remove()} aria-hidden="true" />
          </button>
        </li>
      ))}
    </ul>
  );
}
