import FilterBarExample from "./examples/pattern.filter-bar";
import FilterBarExampleSource from "./examples/pattern.filter-bar.tsx?raw";
import { defineCatalogDocument, type RequiredComponentSections } from "./schema";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture, REVIEW_DATE } from "./page-composition-document";
import { pageStructureDocuments, pageStructureFixtures } from "./page-structure";
import { statusTagAndEmptyStateDocuments, statusTagAndEmptyStateFixtures } from "./status-tag-and-empty-state";
import { statGridAndDescriptionListDocuments, statGridAndDescriptionListFixtures } from "./stat-grid-and-description-list";
import { filtersAndChartsDocuments, filtersAndChartsFixtures } from "./filters-and-charts";

export const pageCompositionFixtures = [
  ...pageStructureFixtures,
  ...statusTagAndEmptyStateFixtures,
  ...statGridAndDescriptionListFixtures,
  ...filtersAndChartsFixtures,
  fixture({ id: "pattern.filter-bar", name: "Filtered logs", description: "A Filters menu beside the Logs heading, removable chips with the result count, and the filtered list.", tags: ["filters", "pattern", "logs"], code: FilterBarExampleSource, render: () => <FilterBarExample /> }),
] as const;

const appliedFiltersSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["AppliedFilters shows the filters narrowing a result, as text or as chips that remove themselves, and clears them in one action."] },
  whenToUse: { heading: "When to use", content: ["Use AppliedFilters below the heading whose FilterMenu narrows the results, so readers can see and remove what still applies while the menu is closed."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it when every filter field stays visible beside the results; a Clear filters button next to the fields is enough. Do not use it for navigation between collections; use Tabbar."] },
  choosingBetween: { heading: "Choosing between", content: ["Use removable when each value can be dropped on its own, as with FilterMenu. Use the text variant for a summary with one reset action. TagsInput edits a list of values in a form."] },
  anatomy: { heading: "Anatomy", content: ["A named group. The text variant joins the labels in one line; the removable variant lists each value as a chip button with a ×. A link-style Clear all follows, then the removable variant's optional result count at the end of the row. It renders nothing when no filter applies."] },
  recommendedExample: { heading: "Recommended example", content: ["Removable filter chips: remove Listening and focus moves to the next chip; Clear all removes the rest and returns focus to the Filters button."] },
  variants: { heading: "Variants", content: ["variant is text (default) or removable. label and clearLabel localize the group name and action."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Labels and chips wrap; Clear all and the count wrap below them at narrow widths. Chips are at least 32px tall on phones."] },
  behavior: { heading: "Behavior", content: ["A chip calls its onRemove and focus moves to the next chip. Clear all calls onClear. The screen resets its filter state and moves focus to the Filters trigger, because the group disappears once nothing is applied; it does the same when the last chip is removed."] },
  contentGuidance: { heading: "Content guidance", content: ["Use the value a reader chose, such as Japanese, and prefix free text with its field, such as Name: Anton."] },
  accessibility: { heading: "Accessibility", content: ["The group is named Applied filters, and each chip is named Remove {value}. Announce result counts in a separate role=\"status\" element owned by the screen."] },
  implementation: { heading: "Implementation", content: ["Import AppliedFilters from paper-ui. Derive filters from the same form state that filters the results."] },
  apiReference: { heading: "API reference", content: ["AppliedFiltersProps is the text variant, with filters as readonly string[], or variant=\"removable\" with filters as { id, label, onRemove }[] and an optional resultLabel. Both take onClear, label (default Applied filters), clearLabel (default Clear all) and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["The Filter bar pattern composes FilterMenu, removable AppliedFilters, the result list and EmptyState."] },
  migration: { heading: "Migration", content: ["Replace application applied-filter summaries that join labels and add a Clear all link button."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

export const pageCompositionDocuments = [
  ...pageStructureDocuments,
  ...statusTagAndEmptyStateDocuments,
  ...statGridAndDescriptionListDocuments,
  ...filtersAndChartsDocuments,
  componentDocument({
    id: "component.applied-filters",
    route: "/components/forms/applied-filters",
    name: "AppliedFilters",
    category: "forms",
    summary: "Summarizes the filters narrowing a result and clears them in one action.",
    keywords: ["filters", "applied filters", "clear all", "summary", "search"],
    sourcePath: "src/components/forms/AppliedFilters/AppliedFilters.tsx",
    fixtureIds: ["applied-filters.removable", "pattern.filter-bar"],
    behaviorTestIds: ["applied-filters.empty", "applied-filters.group-name", "applied-filters.clear", "applied-filters.localized", "applied-filters.remove", "applied-filters.remove-focus"],
    sections: appliedFiltersSections,
    commonMistakes: ["Leaving focus on Clear all, which disappears after clearing.", "Showing a summary next to filter fields that are already visible."],
    api: {
      react: ["AppliedFilters"],
      props: [
        { name: "variant", type: "\"text\" | \"removable\"", defaultValue: "text", description: "A joined summary, or chips that remove themselves." },
        { name: "filters", type: "readonly string[] | readonly RemovableFilter[]", required: true, description: "Labels for text; { id, label, onRemove } for removable. Empty renders nothing." },
        { name: "resultLabel", type: "string", description: "Removable only: the result count at the end of the row, such as 83 logs of 127." },
        { name: "onClear", type: "() => void", required: true, description: "Reset every filter and move focus to the toggle or first field." },
        { name: "label", type: "string", defaultValue: "Applied filters", description: "Accessible name of the group." },
        { name: "clearLabel", type: "string", defaultValue: "Clear all", description: "Text of the reset action." },
      ],
      cssClasses: ["paper-applied-filters", "paper-applied-filters__list", "paper-applied-filters__chips", "paper-applied-filters__chip", "paper-applied-filters__count"],
      publicTypes: ["AppliedFiltersProps", "TextAppliedFiltersProps", "RemovableAppliedFiltersProps", "RemovableFilter"],
      defaults: ["variant=text", "label=Applied filters", "clearLabel=Clear all"],
      invalidCombinations: [],
    },
    dependencies: ["pattern.filter-bar", "component.filter-menu", "component.button"],
    legacy: ["paper-playground records-applied-filters and standings-filter-summary"],
  }),
] as const;

export const pageCompositionPatternDocuments = [
  defineCatalogDocument({
    id: "pattern.filter-bar", route: "/patterns/filter-bar", name: "Filter bar", kind: "pattern", category: "patterns", aliases: [],
    summary: "Narrows one list with a Filters menu beside its heading, removable chips for what applies, a result count and a filtered empty state.",
    keywords: ["filters", "filter bar", "filter menu", "chips", "logs", "standings"], lifecycle: "Experimental", reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/page-composition.tsx", packageVersion: "0.1.0",
    guidance: {
      whenToUse: ["Use when readers narrow one long list by a few filters that each take several values, such as logs by language, activity and tag, or standings by language."],
      whenNotToUse: ["When one search field or two selects fit beside the list, keep them visible with a Clear filters button.", "Do not use filters to switch between views with their own addresses; use Tabs or Tabbar."],
      content: ["Order: FilterMenu in the actions of the heading of the section it filters, AppliedFilters removable chips with the result count, then the results. With nothing applied, the Filters button is the only trace.", "When the filters match nothing, after removing a value or from a stale address, show a plain EmptyState that explains the filters, with a Clear filters action.", "Values within a filter combine with or, and filters combine with and. The menu lists only values that still have results under the other filters, with their points.", "Choices apply as they are ticked. Clear all, in the menu, the sheet or beside the chips, resets every filter and returns focus to the Filters trigger.", "Filters are URL search parameters, and pagination resets when one changes. Phones get a full-screen sheet with a Show {result} button."],
      commonMistakes: ["Placing filters away from the heading of the list they narrow.", "Offering values that would leave no results.", "Losing focus when the last chip or Clear all disappears.", "Treating a filtered empty result like an empty collection: explain the filters, not how to add content."],
    },
    accessibility: {
      requirements: ["The trigger's name includes the number of chosen values. Menu sections are named groups of menu checkboxes; the sheet is a modal dialog with native checkboxes.", "Each chip is named Remove {value}; the chips form a group named Applied filters; the count is a polite status message owned by the screen."],
      keyboard: ["Enter or Space on Filters opens the menu; arrow keys move between values and Space ticks one without closing. Escape closes and returns focus to Filters. Tab reaches the chips, Clear all and the results."],
      knownConstraints: ["The screen owns the filter state, its search parameters, the filtering, each value's points and pagination."],
    },
    api: {
      react: ["SectionHeading", "FilterMenu", "AppliedFilters", "EmptyState", "Button"],
      cssClasses: [],
      publicTypes: ["FilterMenuSection", "FilterMenuHandle", "RemovableFilter"],
      defaults: ["Nothing is applied until the reader chooses a value or opens a link with filters in its address."],
      invalidCombinations: ["The pattern is guidance; FilterMenu and AppliedFilters are the exported controls."],
    },
    fixtureIds: ["pattern.filter-bar"],
    dependencies: { documents: ["component.filter-menu", "component.applied-filters", "component.section-heading", "component.empty-state", "component.button", "component.drawer"], packages: ["paper-ui"] },
    migration: { legacy: ["paper-playground standings and activity-log filters", "the collapsible filter panel of selects"], notes: ["Move each filter into one FilterMenu beside the section heading, keep the selection in the address, and replace the text summary with removable chips."] },
    changelog: [{ date: REVIEW_DATE, note: "Documented from the paper-playground standings and activity-log filters." }, { date: "2026-11-10", note: "Replaced the collapsible panel with FilterMenu and removable chips." }],
    behaviorTestIds: ["pattern.filter-bar.narrow", "pattern.filter-bar.clear"],
  }),
] as const;
