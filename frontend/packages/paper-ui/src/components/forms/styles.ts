import { tv } from "../../foundations/variants";

/** A labelled field: label, control, hint and error. */
export const field = tv({
  slots: {
    root: "paper-field grid w-full min-w-0 content-start gap-1.4",
    label: "paper-field__label text-label leading-label font-semibold text-ink",
    hint: "paper-field__hint m-0 text-metadata leading-metadata text-muted",
    error: "paper-field__error m-0 text-metadata leading-metadata font-semibold text-danger",
  },
  variants: {
    invalid: { true: { root: "paper-field--invalid" } },
    disabled: { true: { root: "paper-field--disabled text-muted" } },
    /** A hint under a checkbox lines up with the label text, not the box. */
    afterChoice: { true: { hint: "-mt-1 ps-check-hint" } },
  },
});

/** Text inputs, text areas and selects. */
export const control = tv({
  base: [
    "paper-input min-h-control w-full rounded-none border border-field-edge bg-field px-field-x py-field-y",
    "font-body text-body leading-125 text-ink",
    "[&::-webkit-datetime-edit-fields-wrapper]:p-0 placeholder:text-muted placeholder:opacity-full",
    "focus-visible:relative focus-visible:z-lift focus-visible:border-focus",
    "focus-visible:outline-halo focus-visible:outline-focus-halo focus-visible:outline-offset-ring",
    "forced-colors:focus-visible:border-b-forced-highlight forced-colors:focus-visible:outline-forced-highlight",
    "disabled:cursor-default disabled:border-rule disabled:bg-paper disabled:text-muted disabled:placeholder:opacity-faded",
    "[&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none",
  ],
  variants: {
    kind: {
      input: "",
      textarea: "paper-textarea min-h-textarea resize-y leading-body",
      select: "paper-select appearance-none pe-select-end has-[option[value='']:checked]:text-muted",
    },
    interactive: { true: "not-disabled:hover:border-field-edge-hover" },
    readOnly: { true: "border-b-rule bg-raised" },
    invalid: {
      true: "border-danger border-b-danger focus-visible:border-danger focus-visible:outline-danger-halo",
    },
  },
  defaultVariants: { kind: "input" },
});

/** The wrapper that positions a select's chevron. */
export const selectFrame = tv({
  slots: {
    root: "paper-select__control relative block w-full has-[select:disabled]:[&_.paper-select__icon]:text-muted",
    icon: "paper-select__icon pointer-events-none absolute end-field-x top-1/2 z-lift-2 -translate-y-1/2 text-ink",
  },
});

/** Fieldsets that group checkboxes, radios and toggles. */
export const choiceField = tv({
  slots: {
    root: "paper-choice-field m-0 grid gap-2 border-0 p-0",
    legend: "float-start w-full p-0",
    list: "paper-choice-list flex flex-wrap gap-x-5 gap-y-3",
  },
  variants: {
    invalid: { true: { root: "paper-field--invalid" } },
  },
});

/** A checkbox or radio beside its label. */
export const choice = tv({
  slots: {
    root: "paper-choice group/choice inline-flex cursor-pointer items-start gap-inline",
    input: [
      "peer size-check accent-action",
      "focus-visible:outline-ring focus-visible:outline-offset-ring focus-visible:inset-ring",
      "forced-colors:focus-visible:outline-halo forced-colors:focus-visible:outline-forced-highlight",
      "forced-colors:focus-visible:outline-offset-0",
      "[&[type=checkbox]:disabled]:rounded-chip [&[type=checkbox]:disabled]:opacity-faded [&[type=checkbox]:disabled]:outline",
      "[&[type=checkbox]:disabled]:outline-checkbox-disabled-edge [&[type=checkbox]:disabled]:-outline-offset-px",
    ],
    text: "peer-disabled:cursor-default peer-disabled:text-muted",
    body: "paper-choice__body grid min-w-0 gap-0.5 peer-disabled:cursor-default peer-disabled:text-muted",
    label: [
      "paper-choice__label font-semibold wrap-anywhere",
      "group-has-disabled/choice:font-regular group-has-disabled/choice:text-muted",
    ],
    description: "paper-choice__description text-metadata text-muted",
    aside: "paper-choice__aside text-end text-metadata text-muted tabular-nums max-sm:col-start-2 max-sm:text-start",
  },
  variants: {
    card: {
      true: {
        root: [
          "paper-choice--card grid w-full grid-cols-icon-content-end items-center gap-x-inline gap-y-1",
          "border border-field-edge bg-field px-4 py-3 max-sm:grid-cols-icon-content",
          "not-has-disabled:hover:bg-neutral-hover",
          "has-checked:border-link has-checked:bg-action-soft has-checked:shadow-underline-link",
          "has-disabled:cursor-default has-disabled:border-rule has-disabled:bg-paper",
        ],
        input: "m-0",
      },
      false: { input: "m-0 mt-0.8" },
    },
    invalid: {
      true: { input: "focus-visible:outline-danger focus-visible:inset-ring-danger" },
      false: { input: "focus-visible:outline-focus focus-visible:inset-ring-focus" },
    },
  },
  defaultVariants: { card: false, invalid: false },
});

/** Segmented radios: connected buttons that each show a check when selected. */
export const segment = tv({
  slots: {
    list: "inline-flex w-fit max-w-full flex-wrap items-stretch gap-0",
    root: [
      "paper-radio-select__segment group/segment relative inline-grid min-h-control min-w-radio-min grow basis-28",
      "grid-cols-label-check items-center -ms-px -mt-px first:ms-0 cursor-pointer",
      "border border-field-edge bg-field px-field-x py-field-y",
      "text-body leading-125 font-semibold text-ink",
      "not-has-disabled:not-has-checked:hover:border-field-edge-hover not-has-disabled:not-has-checked:hover:bg-neutral-hover",
      "has-checked:z-lift has-checked:border-link has-checked:bg-action-soft",
      "has-focus-visible:z-lift-2 has-focus-visible:outline-halo has-focus-visible:outline-offset-ring",
      "has-disabled:cursor-default has-disabled:border-rule has-disabled:bg-paper has-disabled:text-muted",
      "forced-colors:border-forced-ink forced-colors:bg-forced-canvas forced-colors:text-forced-ink",
      "forced-colors:forced-color-adjust-none",
      "forced-colors:has-checked:border-forced-highlight forced-colors:has-checked:bg-forced-canvas",
      "forced-colors:has-focus-visible:border-forced-highlight forced-colors:has-focus-visible:outline-forced-highlight",
      "forced-colors:has-disabled:text-forced-muted",
    ],
    input: "sr-only",
    check: [
      "paper-radio-select__check ms-inline size-icon-sm stroke-check text-link opacity-0",
      "group-has-checked/segment:opacity-full forced-colors:text-forced-highlight",
    ],
  },
  variants: {
    invalid: {
      true: {
        root: [
          "border-danger border-b-danger has-checked:border-danger has-focus-visible:border-danger",
          "has-focus-visible:outline-danger-halo",
          "forced-colors:border-forced-mark",
        ],
      },
      false: { root: "has-focus-visible:border-focus has-focus-visible:outline-focus-halo" },
    },
  },
  defaultVariants: { invalid: false },
});

/** Radio cards: each option with a description. */
export const radioCard = tv({
  slots: {
    list: "paper-radio-cards grid grid-cols-radio-cards gap-2.6",
    root: [
      "paper-radio-card grid cursor-pointer grid-cols-mark-label gap-3 border border-field-edge bg-field p-3.4",
      "has-checked:border-b-edge has-checked:border-link",
      "has-focus-visible:relative has-focus-visible:z-lift has-focus-visible:outline-halo has-focus-visible:outline-offset-ring",
      "has-disabled:cursor-default has-disabled:border-rule has-disabled:bg-paper has-disabled:text-muted",
      "forced-colors:has-focus-visible:border-b-forced-highlight forced-colors:has-focus-visible:outline-forced-highlight",
    ],
    input: "mt-1 accent-action focus-visible:outline-none",
    text: "grid gap-0.8",
    description: "text-muted",
  },
  variants: {
    invalid: {
      true: { root: "has-focus-visible:border-danger has-focus-visible:outline-danger-halo" },
      false: { root: "has-focus-visible:border-focus has-focus-visible:outline-focus-halo" },
    },
  },
  defaultVariants: { invalid: false },
});

/** An amount input joined to its unit select. Focus on either control rings the whole field. */
export const compoundField = tv({
  slots: {
    root: [
      "paper-compound-field grid min-h-control grid-cols-content-end border border-field-edge bg-field",
      "not-has-disabled:hover:border-field-edge-hover has-disabled:border-rule has-disabled:bg-paper",
      "forced-colors:border-forced-ink forced-colors:bg-forced-canvas",
    ],
    input: "min-h-compound-input border-0 bg-transparent",
    unit: [
      "paper-compound-field__unit h-compound-unit w-auto min-w-unit-select border-s border-rule",
      "forced-colors:border-s-forced-ink",
    ],
    select: "h-full min-h-0 border-0 bg-transparent",
  },
  variants: {
    invalid: {
      true: {
        root: "border-danger forced-colors:border-forced-mark",
      },
      false: {
        root: "",
      },
    },
  },
  defaultVariants: { invalid: false },
});

/** A connected set of pressable options with at most one selected. */
export const toggleSelect = tv({
  slots: {
    root: "paper-toggle-select @container min-w-0",
    legend: "mb-2",
    options: [
      "paper-toggle-select__options flex flex-wrap gap-px border border-field-edge bg-rule",
      "@max-toggle:flex-col forced-colors:border-forced-ink forced-colors:bg-forced-ink",
    ],
    option: [
      "paper-toggle-select__option flex min-h-target min-w-0 grow basis-34 cursor-pointer items-center justify-between gap-2",
      "@max-toggle:basis-auto rounded-none border-0 bg-field px-field-x py-field-y text-start",
      "font-body text-sm leading-125 font-semibold text-ink",
      "not-disabled:not-aria-pressed:hover:bg-toggle-hover aria-pressed:bg-action-soft",
      "focus-visible:relative focus-visible:z-lift focus-visible:outline-ring focus-visible:outline-focus",
      "focus-visible:-outline-offset-ring disabled:cursor-default disabled:text-muted",
      "forced-colors:bg-forced-canvas forced-colors:text-forced-ink",
      "forced-colors:aria-pressed:bg-forced-highlight forced-colors:aria-pressed:text-forced-on-highlight",
      "forced-colors:aria-pressed:forced-color-adjust-none forced-colors:focus-visible:outline-forced-highlight",
      "forced-colors:focus-visible:outline-offset-ring forced-colors:disabled:text-forced-muted",
      "group/toggle",
    ],
    text: [
      "paper-toggle-select__text grid min-w-0 gap-0.5 wrap-anywhere group-disabled/toggle:text-muted",
      "@max-toggle:flex @max-toggle:items-baseline @max-toggle:gap-2",
    ],
    description: [
      "paper-toggle-select__description text-metadata leading-metadata font-regular text-muted",
      "forced-colors:group-aria-pressed/toggle:text-inherit",
    ],
    check: "paper-toggle-select__check invisible shrink-0 group-aria-pressed/toggle:visible",
  },
  variants: {
    invalid: { true: { options: "border-danger" } },
  },
});

/** Autocomplete, multi-select and tag inputs, and their popups. */
export const combobox = tv({
  slots: {
    group: "paper-combobox__input-group relative",
    input: "pe-combobox-end",
    trigger: [
      "paper-combobox__trigger absolute end-0 top-1/2 z-lift inline-grid size-target -translate-y-1/2",
      "cursor-pointer place-items-center border-0 bg-transparent p-0 text-muted",
      "disabled:cursor-default disabled:opacity-disabled",
    ],
    positioner: "paper-combobox__positioner z-popover outline-none",
    popup: "paper-combobox__popup max-h-popup w-(--anchor-width) overflow-y-auto py-1.4",
    item: [
      "paper-combobox__item group/item grid min-h-control cursor-default grid-cols-item-indicator items-center",
      "px-field-x py-field-y outline-none data-highlighted:bg-action data-highlighted:text-on-action",
    ],
    indicator: "paper-combobox__indicator invisible group-data-selected/item:visible",
    // Base UI unmounts the indicator on unselected items, so the label must name its column.
    label: "col-start-2",
    empty: "paper-combobox__empty px-field-x py-field-y text-muted empty:hidden",
    chips: [
      "paper-combobox__chips relative flex min-h-control flex-wrap items-center gap-1.4 p-1 pe-combobox-end has-[input:disabled]:border-rule has-[input:disabled]:bg-paper",
      "border border-field-edge bg-field font-body text-body leading-125",
      "has-[input:focus-visible]:z-lift",
      "has-[input:focus-visible]:outline-halo",
      "has-[input:focus-visible]:outline-offset-ring",
      "has-[>button:focus-visible]:z-lift",
      "has-[>button:focus-visible]:outline-halo",
      "has-[>button:focus-visible]:outline-offset-ring",
      "forced-colors:has-[input:focus-visible]:border-b-forced-highlight",
      "forced-colors:has-[input:focus-visible]:outline-forced-highlight",
      "forced-colors:has-[>button:focus-visible]:border-b-forced-highlight",
      "forced-colors:has-[>button:focus-visible]:outline-forced-highlight",
    ],
    chipTrigger: "focus-visible:outline-none",
    chip: [
      "paper-combobox__chip inline-flex min-w-0 max-w-full items-center gap-0.5 rounded-chip bg-chip",
      "py-0.5 ps-2 pe-0.5 text-metadata leading-150 font-medium text-ink wrap-anywhere",
    ],
    chipRemove: [
      "inline-grid min-h-chip-target min-w-chip-target max-tablet:min-h-8 max-tablet:min-w-8 shrink-0 cursor-pointer place-items-center rounded-chip",
      "border-0 bg-transparent p-0 text-ink hover:bg-chip-remove-hover",
      "focus-visible:outline-ring focus-visible:outline-focus focus-visible:outline-offset-ring",
    ],
    chipInput: [
      "paper-combobox__chip-input min-w-chip-input grow basis-36 border-0 bg-transparent outline-none",
      "first:ps-chip-input-start placeholder:text-muted placeholder:opacity-full",
    ],
    option: "paper-tags-input__option block wrap-anywhere",
  },
  variants: {
    invalid: {
      true: {
        chips: [
          "has-[input:focus-visible]:border-danger",
          "has-[input:focus-visible]:outline-danger-halo",
          "has-[>button:focus-visible]:border-danger",
          "has-[>button:focus-visible]:outline-danger-halo",
        ],
      },
      false: {
        chips: [
          "has-[input:focus-visible]:border-focus",
          "has-[input:focus-visible]:outline-focus-halo",
          "has-[>button:focus-visible]:border-focus",
          "has-[>button:focus-visible]:outline-focus-halo",
        ],
      },
    },
  },
  defaultVariants: { invalid: false },
});

/** A form's closing row: actions at the start, an optional summary such as an estimate at the end. */
export const formFooter = tv({
  slots: {
    root: [
      "paper-form-footer mt-6 flex flex-wrap-reverse items-center justify-between gap-x-6 gap-y-3",
      "border-t border-rule pt-6",
    ],
    actions: "paper-form-footer__actions flex flex-wrap items-center gap-x-5 gap-y-3",
  },
});
