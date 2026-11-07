import { tv } from "../../foundations/variants";

/** A record's title link in a list: ink with a quiet underline that takes the link colour on hover and focus. */
export const entryLink = tv({
  base: [
    "font-semibold leading-150 text-ink underline decoration-muted decoration-hairline underline-offset-link wrap-anywhere",
    "hover:text-link hover:decoration-current focus-visible:text-link focus-visible:decoration-current",
  ],
  variants: {
    /** Long descriptions preview two lines; the record's own page shows them whole. */
    clamp: { true: "line-clamp-2" },
  },
});

/** The muted line of facts under a title, joined with middle dots. */
export const metaLine = tv({ base: "mb-0 mt-1.5 text-metadata text-muted" });

/** Rows separated by rules, with no rule above the first or below the last. */
export const listRows = tv({
  slots: {
    root: "m-0 list-none p-0",
    item: "grid border-b border-rule py-3 wrap-anywhere first:pt-0 last:border-b-0 last:pb-0",
  },
});

/**
 * Dated records: a date gutter, the record and its score. Records on the same day as the one above share its
 * date, and their separator starts after the gutter, so each day reads as one group. Phones put the date above.
 */
export const entryList = tv({
  slots: {
    root: "m-0 list-none border-b border-rule p-0",
    item: [
      "relative grid grid-cols-entry items-start gap-6 border-t border-rule py-4 first:border-t-transparent",
      "max-sm:grid-cols-content-end max-sm:gap-x-4 max-sm:gap-y-1.5",
    ],
    sameDay: [
      "border-t-transparent before:absolute before:-top-px before:end-0 before:start-entry-rule before:border-t",
      "before:border-rule-subtle max-sm:before:start-0",
    ],
    date: "pt-0.75 text-metadata text-muted max-sm:col-span-full max-sm:p-0",
    body: "col-start-2 min-w-0 max-sm:col-start-1",
    score: [
      "col-start-3 row-start-1 whitespace-nowrap text-end font-bold tabular-nums max-sm:col-start-2 max-sm:row-auto",
      "[&_small]:text-metadata [&_small]:font-regular [&_small]:text-muted",
    ],
  },
});

/** A list's closing row: its count beside its pagination. */
export const listFooter = tv({
  slots: {
    root: "flex flex-wrap items-center justify-between gap-x-4 gap-y-2",
    summary: "m-0 text-metadata text-muted",
  },
});
