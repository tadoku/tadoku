import { Menu } from "@base-ui/react/menu";
import { forwardRef, useId, useImperativeHandle, useRef, useState, type ReactElement, type ReactNode } from "react";
import { CheckIcon, ChevronDownIcon, iconClassName } from "../../../icons";
import { surface } from "../../../foundations/recipes";
import { cn, tv } from "../../../foundations/variants";
import { Button, buttonClassName } from "../../actions/Button";
import { Drawer } from "../../overlays/Drawer";

const filterMenu = tv({
  slots: {
    root: "paper-filter-menu inline-flex",
    trigger: "paper-filter-menu__trigger",
    selected: "bg-action-soft border-action not-disabled:hover:border-action data-popup-open:border-action",
    positioner: "paper-filter-menu__positioner z-menu outline-none",
    popup: [
      "paper-filter-menu__popup max-h-(--available-height) w-filter-menu overflow-y-auto origin-(--transform-origin)",
      "transition-enter duration-quick motion-reduce:transition-none",
      "data-starting-style:scale-enter data-starting-style:opacity-0",
      "data-ending-style:scale-enter data-ending-style:opacity-0",
    ],
    section: "paper-filter-menu__section border-b border-rule-subtle py-1.4",
    header: "paper-filter-menu__header flex items-baseline justify-between gap-inline px-field-x pb-1 pt-1.5",
    label: "paper-filter-menu__label m-0 text-xs font-bold uppercase tracking-label text-muted",
    clear: [
      "paper-filter-menu__clear cursor-pointer border-0 bg-transparent p-0 text-fine font-semibold text-link",
      "underline underline-offset-link outline-none hover:text-link-hover",
      "data-highlighted:text-link-hover data-highlighted:decoration-hover",
      "focus-visible:outline-ring focus-visible:outline-focus",
    ],
    item: [
      "paper-filter-menu__item flex min-h-control cursor-default select-none items-center gap-inline",
      "px-field-x py-1.5 text-ui outline-none data-highlighted:bg-action-soft",
    ],
    box: [
      "paper-filter-menu__box grid size-check flex-none place-items-center border border-field-edge bg-field",
      "text-on-action",
    ],
    boxChecked: "border-action bg-action",
    value: "paper-filter-menu__value min-w-0 flex-auto wrap-anywhere",
    points: "paper-filter-menu__points flex-none text-sm tabular-nums text-muted",
    more: "paper-filter-menu__more text-sm font-semibold text-link",
    footer: [
      "paper-filter-menu__footer flex items-center justify-between gap-inline px-field-x py-2",
      "text-sm text-muted",
    ],
    sheetSection: "paper-filter-menu__sheet-section border-b border-rule-subtle py-3 first:pt-0 last:border-b-0",
    sheetItem: [
      "paper-filter-menu__sheet-item flex min-h-12 cursor-pointer items-center gap-inline text-base",
    ],
    sheetInput: "size-check flex-none accent-action focus-visible:outline-ring focus-visible:outline-offset-ring",
    sheetFooter: "flex w-full items-center justify-between gap-inline",
  },
});

const slots = filterMenu();
const popupClassName = cn(surface({ tone: "overlay", elevation: "floating" }), slots.popup());
const pointsFormat = new Intl.NumberFormat("en", { maximumFractionDigits: 1 });

export interface FilterMenuValue {
  readonly value: string;
  readonly label: string;
  /** Points the value would still match under the other filters. Values without points are hidden unless chosen. */
  readonly points: number;
}

export interface FilterMenuSection {
  readonly id: string;
  readonly label: string;
  readonly values: readonly FilterMenuValue[];
  readonly selected: readonly string[];
}

export interface FilterMenuProps {
  /** One section per filter. Values within a section combine with "or"; sections combine with "and". */
  readonly sections: readonly FilterMenuSection[];
  readonly onChange: (sectionId: string, selected: readonly string[]) => void;
  /** Clears every section. The menu closes and focus returns to its trigger. */
  readonly onClear: () => void;
  /** The current result, such as "83 logs". The phone sheet's button reads "Show {resultLabel}". */
  readonly resultLabel: string;
  /** Values listed before "Show all"; defaults to 6 in the menu and 8 in the phone sheet. */
  readonly visibleCount?: number;
  readonly label?: string;
}

/** Focuses the visible trigger, for example after Clear all in AppliedFilters removes the focused control. */
export interface FilterMenuHandle {
  readonly focus: () => void;
}

interface ListedSection extends FilterMenuSection {
  readonly listed: readonly FilterMenuValue[];
}

function listed(sections: readonly FilterMenuSection[]): ListedSection[] {
  return sections
    .map((section) => ({
      ...section,
      listed: section.values.filter((value) => value.points > 0 || section.selected.includes(value.value)),
    }))
    .filter((section) => section.listed.length > 0);
}

function shown(section: ListedSection, count: number, expanded: boolean) {
  if (expanded) return section.listed;
  return section.listed.filter((value, index) => index < count || section.selected.includes(value.value));
}

function toggled(section: FilterMenuSection, value: string, checked: boolean) {
  const rest = section.selected.filter((item) => item !== value);
  return checked ? [...rest, value] : rest;
}

function useExpanded() {
  const [expanded, setExpanded] = useState<ReadonlySet<string>>(new Set());
  return {
    isExpanded: (id: string) => expanded.has(id),
    expand: (id: string) => setExpanded((current) => new Set(current).add(id)),
  };
}

function TriggerContent({ label, count }: { readonly label: string; readonly count: number }) {
  return (
    <>
      <span>{count ? `${label} (${count})` : label}</span>
      <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
    </>
  );
}

/**
 * A Filters button that opens checkbox sections: a menu from the tablet breakpoint up, and a full-screen sheet
 * below it. It reports selections; the screen filters, keeps the selection in the address and resets pagination.
 */
export const FilterMenu = forwardRef<FilterMenuHandle, FilterMenuProps>(function FilterMenu(
  { sections, onChange, onClear, resultLabel, visibleCount, label = "Filters" },
  ref,
) {
  const root = useRef<HTMLSpanElement>(null);
  useImperativeHandle(ref, () => ({
    focus: () => {
      const triggers = Array.from(root.current?.querySelectorAll<HTMLElement>(".paper-filter-menu__trigger") ?? []);
      (triggers.find((trigger) => trigger.getClientRects().length > 0) ?? triggers[0])?.focus();
    },
  }));
  const count = sections.reduce((total, section) => total + section.selected.length, 0);
  const triggerClassName = (presentation: "menu" | "sheet") =>
    buttonClassName({
      variant: "ghost",
      className: cn(
        slots.trigger(),
        `paper-filter-menu__trigger--${presentation}`,
        presentation === "menu" ? "max-tablet:hidden" : "tablet:hidden",
        count ? slots.selected() : undefined,
      ),
    });
  const shared = { sections: listed(sections), onChange, onClear, resultLabel, count };

  return (
    <span ref={root} className={slots.root()}>
      <DesktopMenu
        {...shared}
        visibleCount={visibleCount ?? 6}
        trigger={<TriggerContent label={label} count={count} />}
        triggerClassName={triggerClassName("menu")}
      />
      <PhoneSheet
        {...shared}
        visibleCount={visibleCount ?? 8}
        title={label}
        trigger={
          <button type="button" className={triggerClassName("sheet")}>
            <TriggerContent label={label} count={count} />
          </button>
        }
      />
    </span>
  );
});

interface PresentationProps {
  readonly sections: readonly ListedSection[];
  readonly onChange: FilterMenuProps["onChange"];
  readonly onClear: FilterMenuProps["onClear"];
  readonly resultLabel: string;
  readonly count: number;
  readonly visibleCount: number;
}

function DesktopMenu({
  sections,
  onChange,
  onClear,
  resultLabel,
  count,
  visibleCount,
  trigger,
  triggerClassName,
}: PresentationProps & { readonly trigger: ReactNode; readonly triggerClassName: string }) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const { isExpanded, expand } = useExpanded();

  return (
    <Menu.Root>
      <Menu.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
        }}
        className={triggerClassName}
      >
        {trigger}
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner className={slots.positioner()} sideOffset={6} align="end">
          <Menu.Popup className={popupClassName}>
            {sections.map((section) => {
              const values = shown(section, visibleCount, isExpanded(section.id));
              return (
                <Menu.Group key={section.id} className={slots.section()}>
                  <div className={slots.header()}>
                    <Menu.GroupLabel className={slots.label()}>{section.label}</Menu.GroupLabel>
                    {section.selected.length ? (
                      <Menu.Item
                        className={slots.clear()}
                        closeOnClick={false}
                        aria-label={`Clear ${section.label}`}
                        onClick={() => onChange(section.id, [])}
                      >
                        Clear
                      </Menu.Item>
                    ) : null}
                  </div>
                  {values.map((value) => {
                    const checked = section.selected.includes(value.value);
                    return (
                      <Menu.CheckboxItem
                        key={value.value}
                        className={slots.item()}
                        checked={checked}
                        onCheckedChange={(next) => onChange(section.id, toggled(section, value.value, next))}
                      >
                        <span className={slots.box({ className: checked ? slots.boxChecked() : undefined })}>
                          <Menu.CheckboxItemIndicator className="grid place-items-center">
                            <CheckIcon className="size-3.5" aria-hidden="true" />
                          </Menu.CheckboxItemIndicator>
                        </span>
                        <span className={slots.value()}>{value.label}</span>
                        <span className={slots.points()}>{pointsFormat.format(value.points)}</span>
                      </Menu.CheckboxItem>
                    );
                  })}
                  {values.length < section.listed.length ? (
                    <Menu.Item
                      className={slots.item({ className: slots.more() })}
                      closeOnClick={false}
                      onClick={() => expand(section.id)}
                    >
                      Show all {section.listed.length}
                    </Menu.Item>
                  ) : null}
                </Menu.Group>
              );
            })}
            <div className={slots.footer()}>
              <span role="status">{resultLabel}</span>
              {count ? (
                <Menu.Item className={slots.clear()} onClick={onClear}>
                  Clear all
                </Menu.Item>
              ) : null}
            </div>
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}

function PhoneSheet({
  sections,
  onChange,
  onClear,
  resultLabel,
  count,
  visibleCount,
  title,
  trigger,
}: PresentationProps & { readonly title: string; readonly trigger: ReactElement }) {
  const [open, setOpen] = useState(false);
  const { isExpanded, expand } = useExpanded();
  const id = useId();

  return (
    <Drawer
      size="full"
      open={open}
      onOpenChange={setOpen}
      trigger={trigger}
      title={title}
      closeLabel={`Close ${title.toLocaleLowerCase()}`}
      footer={
        <div className={slots.sheetFooter()}>
          {count ? (
            <Button
              variant="link"
              onClick={() => {
                onClear();
                setOpen(false);
              }}
            >
              Clear all
            </Button>
          ) : <span />}
          <Button onClick={() => setOpen(false)}>Show {resultLabel}</Button>
        </div>
      }
    >
      {sections.map((section) => {
        const values = shown(section, visibleCount, isExpanded(section.id));
        const labelId = `${id}-${section.id}`;
        return (
          <div key={section.id} role="group" aria-labelledby={labelId} className={slots.sheetSection()}>
            <div className={slots.header({ className: "px-0" })}>
              <p id={labelId} className={slots.label()}>{section.label}</p>
              {section.selected.length ? (
                <button
                  type="button"
                  className={slots.clear({ className: "min-h-target" })}
                  aria-label={`Clear ${section.label}`}
                  onClick={() => onChange(section.id, [])}
                >
                  Clear
                </button>
              ) : null}
            </div>
            {values.map((value) => (
              <label key={value.value} className={slots.sheetItem()}>
                <input
                  type="checkbox"
                  className={slots.sheetInput()}
                  checked={section.selected.includes(value.value)}
                  onChange={(event) => onChange(section.id, toggled(section, value.value, event.target.checked))}
                />
                <span className={slots.value()}>{value.label}</span>
                <span className={slots.points()}>{pointsFormat.format(value.points)}</span>
              </label>
            ))}
            {values.length < section.listed.length ? (
              <Button variant="link" className="min-h-12" onClick={() => expand(section.id)}>
                Show all {section.listed.length}
              </Button>
            ) : null}
          </div>
        );
      })}
    </Drawer>
  );
}
