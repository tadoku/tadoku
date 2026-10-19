import FilterBarExample from "./examples/pattern.filter-bar";
import FilterBarExampleSource from "./examples/pattern.filter-bar.tsx?raw";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture, REVIEW_DATE } from "./page-composition-document";
import { pageStructureDocuments, pageStructureFixtures } from "./page-structure";
import { defineCatalogDocument, type RequiredComponentSections } from "./schema";
import { statGridAndDescriptionListDocuments, statGridAndDescriptionListFixtures } from "./stat-grid-and-description-list";
import { statusTagAndEmptyStateDocuments, statusTagAndEmptyStateFixtures } from "./status-tag-and-empty-state";

export const pageCompositionFixtures = [
  ...pageStructureFixtures,
  ...statusTagAndEmptyStateFixtures,
  ...statGridAndDescriptionListFixtures,
  fixture({ id: "pattern.filter-bar", name: "Standings filter bar", description: "Collapsible filters, applied-filter summary, result count and a filtered empty state.", tags: ["filters", "pattern", "standings"], code: FilterBarExampleSource, render: () => <FilterBarExample /> }),
] as const;

const appliedFiltersSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["AppliedFilters summarizes the filters narrowing a result and clears them in one action."] },
  whenToUse: { heading: "When to use", content: ["Use AppliedFilters below a collapsible filter panel so readers can see and reset what still narrows the results after the panel closes."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it when every filter field stays visible beside the results; a Clear filters button next to the fields is enough. Do not use it for navigation between collections; use Tabbar."] },
  choosingBetween: { heading: "Choosing between", content: ["AppliedFilters is a summary with one reset action. TagsInput edits a list of values."] },
  anatomy: { heading: "Anatomy", content: ["A named group with the applied filter labels joined in one line, followed by a link-style Clear all button. It renders nothing when no filter applies."] },
  recommendedExample: { heading: "Recommended example", content: ["Open Filters, choose Spanish and close the panel: the summary keeps Spanish visible and the count updates. Clear all resets the filters and returns focus to the Filters toggle."] },
  variants: { heading: "Variants", content: ["There is one appearance. label and clearLabel localize the group name and action."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Labels wrap anywhere; the action wraps below them at narrow widths."] },
  behavior: { heading: "Behavior", content: ["Clear all calls onClear. The screen resets its filter state and moves focus to the filter toggle or first field, because the summary disappears once nothing is applied."] },
  contentGuidance: { heading: "Content guidance", content: ["Use the value a reader chose, such as Japanese, and prefix free text with its field, such as Name: Anton."] },
  accessibility: { heading: "Accessibility", content: ["The group is named Applied filters. Announce result counts in a separate role=\"status\" element owned by the screen."] },
  implementation: { heading: "Implementation", content: ["Import AppliedFilters from paper-ui. Derive filters from the same form state that filters the results."] },
  apiReference: { heading: "API reference", content: ["AppliedFiltersProps accepts filters (readonly string[]), onClear, label (default Applied filters), clearLabel (default Clear all) and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["The Filter bar pattern composes AppliedFilters with a Filters toggle, the filter panel, a result count and EmptyState."] },
  migration: { heading: "Migration", content: ["Replace application applied-filter summaries that join labels and add a Clear all link button."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

export const pageCompositionDocuments = [
  ...pageStructureDocuments,
  ...statusTagAndEmptyStateDocuments,
  ...statGridAndDescriptionListDocuments,
  componentDocument({
    id: "component.applied-filters",
    route: "/components/forms/applied-filters",
    name: "AppliedFilters",
    category: "forms",
    summary: "Summarizes the filters narrowing a result and clears them in one action.",
    keywords: ["filters", "applied filters", "clear all", "summary", "search"],
    sourcePath: "src/components/forms/AppliedFilters/AppliedFilters.tsx",
    fixtureIds: ["pattern.filter-bar"],
    behaviorTestIds: ["applied-filters.empty", "applied-filters.group-name", "applied-filters.clear", "applied-filters.localized"],
    sections: appliedFiltersSections,
    commonMistakes: ["Leaving focus on Clear all, which disappears after clearing.", "Showing a summary next to filter fields that are already visible."],
    api: {
      react: ["AppliedFilters"],
      props: [
        { name: "filters", type: "readonly string[]", required: true, description: "Readable labels of the applied filters. Empty renders nothing." },
        { name: "onClear", type: "() => void", required: true, description: "Reset every filter and move focus to the toggle or first field." },
        { name: "label", type: "string", defaultValue: "Applied filters", description: "Accessible name of the group." },
        { name: "clearLabel", type: "string", defaultValue: "Clear all", description: "Text of the reset action." },
      ],
      cssClasses: ["paper-applied-filters", "paper-applied-filters__list"],
      publicTypes: ["AppliedFiltersProps"],
      defaults: ["label=Applied filters", "clearLabel=Clear all"],
      invalidCombinations: [],
    },
    dependencies: ["pattern.filter-bar", "component.button"],
    legacy: ["paper-playground records-applied-filters and standings-filter-summary"],
  }),
] as const;

export const pageCompositionPatternDocuments = [
  defineCatalogDocument({
    id: "pattern.filter-bar", route: "/patterns/filter-bar", name: "Filter bar", kind: "pattern", category: "patterns", aliases: [],
    summary: "Narrows a list or board with collapsible filters, a visible summary, a live result count and a filtered empty state.",
    keywords: ["filters", "filter bar", "search", "standings", "activity log"], lifecycle: "Experimental", reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/page-composition.tsx", packageVersion: "0.1.0",
    guidance: {
      whenToUse: ["Use when readers narrow a long list or board by a few fields, such as standings by language and activity or an activity log by period and title."],
      whenNotToUse: ["When one search field or two selects fit beside the list, keep them visible with a Clear filters button and skip the toggle and summary.", "Do not use filters to switch between collections with their own addresses; use Tabbar."],
      content: ["Order: a Filters toggle in the section heading’s actions, the filter panel, AppliedFilters with the result count, then the results or a plain EmptyState.", "The toggle reads Filters, or Filters (n) when n filters apply, and uses aria-expanded and aria-controls for the panel. Keep the panel mounted while collapsed so selections persist.", "Filters apply as fields change. AppliedFilters stays visible while the panel is collapsed; Clear all and the empty state’s Clear filters reset every field and return focus to the toggle.", "Announce the result count in a role=\"status\" element, such as 2 of 5 participants."],
      commonMistakes: ["Hiding which filters still apply after the panel closes.", "Losing focus when the Clear all button disappears.", "Treating a filtered empty result like an empty collection: explain the filters, not how to add content."],
    },
    accessibility: {
      requirements: ["The toggle exposes its expanded state and controls the panel by id.", "Every field has a visible label; the summary is a named group; the count is a polite status message."],
      keyboard: ["Tab reaches the toggle, then the panel fields when expanded, then Clear all and the results. Enter or Space toggles the panel."],
      knownConstraints: ["The screen owns the filter state, the query string if filters are linkable, and any default-open rule, such as opening yearly standings filters so the year is reachable."],
    },
    api: {
      react: ["SectionHeading", "Button", "Select", "Input", "AppliedFilters", "EmptyState", "Table"],
      cssClasses: [],
      publicTypes: [],
      defaults: ["The panel starts collapsed unless a required filter, such as a year, must be reachable."],
      invalidCombinations: ["The pattern is guidance; only AppliedFilters is an exported control."],
    },
    fixtureIds: ["pattern.filter-bar"],
    dependencies: { documents: ["component.applied-filters", "component.section-heading", "component.empty-state", "component.button", "component.select", "component.input", "component.table"], packages: ["paper-ui", "@tanstack/react-form"] },
    migration: { legacy: ["paper-playground standings and activity-log filters"], notes: ["Keep each screen’s filter state and field layout; replace hand-built summaries with AppliedFilters and empty results with EmptyState."] },
    changelog: [{ date: REVIEW_DATE, note: "Documented from the paper-playground standings and activity-log filters." }],
    behaviorTestIds: ["pattern.filter-bar.summary", "pattern.filter-bar.empty"],
  }),
] as const;
