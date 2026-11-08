import { tv } from "../../foundations/variants";

export const navigationIcon = "paper-navigation__icon grid size-icon flex-none place-items-center *:size-full";

const barLink = [
  "inline-flex min-h-control cursor-pointer items-center gap-inline border-0 bg-transparent px-field-x",
  "font-semibold text-ink no-underline aria-disabled:cursor-default aria-disabled:opacity-faded",
  "aria-[current=page]:shadow-current-bar",
];

export const navbar = tv({
  slots: {
    root: "paper-navbar relative z-header border-b border-rule bg-paper",
    inner: "paper-navbar__inner flex min-h-control items-center justify-between gap-inline px-field-x",
    brand: "paper-navbar__brand flex items-center",
    brandLink: "inline-flex items-center font-bold text-ink no-underline",
    links: "paper-navbar__desktop-links ms-auto flex items-stretch gap-1 max-lg:hidden",
    link: ["paper-navbar__link", ...barLink, "not-aria-disabled:hover:bg-neutral-hover"],
    trigger: ["paper-navbar__dropdown-trigger", ...barLink, "hover:bg-neutral-hover"],
    actions: "paper-navbar__actions flex items-center gap-inline",
    mobileTrigger: [
      "paper-navbar__mobile-trigger hidden size-target min-h-target min-w-target flex-none cursor-pointer",
      "place-items-center border-0 bg-transparent p-0 text-ink disabled:cursor-progress max-lg:grid",
    ],
    positioner: "paper-navbar__positioner z-popover outline-none",
    dropdown: [
      "paper-navbar__dropdown min-w-menu origin-(--transform-origin) py-1.4",
      "transition-enter duration-quick motion-reduce:transition-none",
      "data-starting-style:scale-enter data-starting-style:opacity-0",
      "data-ending-style:scale-enter data-ending-style:opacity-0",
    ],
    item: [
      "paper-navbar__dropdown-item flex min-h-control cursor-default items-center gap-inline px-field-x py-field-y",
      "text-ink no-underline outline-none data-highlighted:bg-action data-highlighted:text-on-action",
      "data-disabled:cursor-default data-disabled:opacity-faded",
    ],
    panel: "paper-navbar__mobile-panel hidden border-t border-rule-subtle p-2",
    mobileLink: ["paper-navbar__mobile-link w-full", ...barLink, "not-aria-disabled:hover:bg-neutral-hover"],
    groupTitle: "mx-field-x mb-1 mt-4 text-metadata leading-metadata text-muted uppercase",
    loading: [
      "paper-navbar__loading absolute inset-x-0 -bottom-px h-0.5 overflow-hidden bg-rule-subtle",
      "after:block after:h-full after:w-2/5 after:animate-progress after:bg-action",
      "motion-reduce:after:w-full motion-reduce:after:animate-none",
    ],
  },
  variants: {
    open: { true: { panel: "max-lg:block" } },
    alone: { true: { actions: "ms-auto" } },
    /** Inside an application frame's header band: the bar keeps to the frame width and its gutter. */
    contained: {
      true: {
        root: "static mx-auto max-w-frame border-0 px-gutter",
        inner: "px-0 py-2 max-lg:gap-2",
        brand: "max-lg:shrink-0",
        brandLink: "gap-brand-gap",
        actions: "max-lg:ms-auto max-lg:[&_.paper-button]:min-h-target",
        panel: "px-0",
        mobileLink: "max-lg:min-h-target",
      },
    },
  },
});

export type NavbarSlot = keyof typeof navbar.slots;

export const sidebar = tv({
  slots: {
    root: "paper-sidebar grid gap-6 compact:gap-4",
    section: "paper-sidebar__section grid gap-1",
    title: "paper-sidebar__title mb-1 mt-0 font-body text-label leading-label uppercase",
    list: "paper-sidebar__list m-0 grid list-none gap-1 p-0",
    link: [
      "paper-sidebar__link relative flex min-h-control items-center gap-inline pe-field-x ps-rail-inset",
      "text-muted no-underline before:absolute before:inset-y-0 before:start-0 before:w-rail before:bg-transparent",
      "not-aria-disabled:hover:bg-raised not-aria-disabled:hover:text-ink",
      "aria-[current=page]:bg-paper aria-[current=page]:font-semibold aria-[current=page]:text-ink",
      "aria-[current=page]:before:bg-link aria-disabled:cursor-default aria-disabled:opacity-faded",
    ],
  },
});

export const breadcrumb = tv({
  slots: {
    list: "paper-breadcrumb__list m-0 flex min-w-0 list-none flex-wrap items-center gap-1 p-0",
    item: "paper-breadcrumb__item flex items-center gap-1.4",
    /** Phones keep the parent and the current page. */
    distant: "max-tablet:hidden",
    link: "paper-breadcrumb__link flex items-center gap-1.4 text-link underline underline-offset-link",
    current: "paper-breadcrumb__current flex items-center gap-1.4 text-muted",
    separator: "paper-breadcrumb__separator mx-1 text-rule-strong",
  },
});

export const tabbar = tv({
  slots: {
    root: "paper-tabbar",
    list: "paper-tabbar__list m-0 list-none p-0",
    link: [
      "paper-tabbar__link relative flex min-h-control cursor-pointer items-center gap-inline whitespace-nowrap",
      "border-0 bg-transparent px-field-x font-semibold text-muted no-underline",
      "not-aria-disabled:not-aria-[current=page]:hover:bg-tab-hover",
      "not-aria-disabled:not-aria-[current=page]:hover:text-ink",
      "aria-[current=page]:text-ink aria-disabled:cursor-default aria-disabled:opacity-faded",
    ],
  },
  variants: {
    orientation: {
      horizontal: {
        list: "paper-tabbar__list--horizontal flex overflow-x-auto shadow-rule-bottom",
        link: "aria-[current=page]:shadow-current-bar",
      },
      vertical: {
        root: "paper-tabbar--vertical",
        list: "paper-tabbar__list--vertical grid shadow-rule-start",
        link: "aria-[current=page]:shadow-current-bar-start",
      },
    },
  },
});

export type TabbarSlot = keyof typeof tabbar.slots;

const paginationItem = [
  "inline-flex min-h-control min-w-control cursor-pointer items-center justify-center gap-1.4",
  "border-0 bg-transparent px-2 text-ink no-underline",
];

export const pagination = tv({
  slots: {
    root: "paper-pagination",
    list: [
      "paper-pagination__list m-0 flex list-none items-center justify-end gap-1 p-0 max-tablet:justify-between",
    ],
    /** Phones keep only the current page, as "2 of 5", between Previous and Next. */
    otherPage: "max-tablet:hidden",
    control: [
      "paper-pagination__control",
      ...paginationItem,
      "not-aria-disabled:hover:bg-neutral-hover aria-disabled:cursor-default aria-disabled:text-muted",
    ],
    page: [
      "paper-pagination__page",
      ...paginationItem,
      "not-aria-[current=page]:hover:bg-neutral-hover",
      "aria-[current=page]:cursor-default aria-[current=page]:font-semibold aria-[current=page]:text-ink",
      "aria-[current=page]:shadow-current-bar",
    ],
    of: "paper-pagination__of hidden font-regular max-tablet:inline",
    gap: "paper-pagination__gap grid w-control place-items-center text-muted max-tablet:hidden",
  },
});

export type PaginationSlot = keyof typeof pagination.slots;
