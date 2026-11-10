import AppliedFiltersRemovableExample from "./examples/applied-filters.removable";
import AppliedFiltersRemovableExampleSource from "./examples/applied-filters.removable.tsx?raw";
import DrawerFullScreenExample from "./examples/drawer.full-screen";
import DrawerFullScreenExampleSource from "./examples/drawer.full-screen.tsx?raw";
import FilterMenuExample from "./examples/filter-menu.log-filters";
import FilterMenuExampleSource from "./examples/filter-menu.log-filters.tsx?raw";
import StackedBarChartExample from "./examples/stacked-bar-chart.contest-activity";
import StackedBarChartExampleSource from "./examples/stacked-bar-chart.contest-activity.tsx?raw";
import StatGridSheetExample from "./examples/stat-grid.contest-summary";
import StatGridSheetExampleSource from "./examples/stat-grid.contest-summary.tsx?raw";
import { type RequiredComponentSections } from "./schema";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture } from "./page-composition-document";

export const filtersAndChartsFixtures = [
  fixture({ id: "filter-menu.log-filters", name: "Log filters", description: "Activity and tag sections with points, one value chosen, and Show all for the twelve tags.", tags: ["filters", "menu", "checkbox", "sheet"], code: FilterMenuExampleSource, render: () => <FilterMenuExample /> }),
  fixture({ id: "applied-filters.removable", name: "Removable filter chips", description: "Three applied values as chips that remove themselves, Clear all and the result count.", tags: ["filters", "chips", "clear all"], code: AppliedFiltersRemovableExampleSource, render: () => <AppliedFiltersRemovableExample /> }),
  fixture({ id: "stat-grid.contest-summary", name: "Contest summary row", description: "Score, active days, logs and best day boxed in one ruled row, two by two on phones.", tags: ["statistics", "contest", "summary"], code: StatGridSheetExampleSource, render: () => <StatGridSheetExample /> }),
  fixture({ id: "stacked-bar-chart.contest-activity", name: "Contest activity", description: "Three months of points per day stacked by activity, with the running total.", tags: ["chart", "activity", "contest"], code: StackedBarChartExampleSource, render: () => <StackedBarChartExample /> }),
  fixture({ id: "drawer.full-screen", name: "Full-screen sheet", description: "A Drawer with size=\"full\" that covers a phone screen and keeps its footer in view.", tags: ["drawer", "phone", "sheet"], code: DrawerFullScreenExampleSource, render: () => <DrawerFullScreenExample /> }),
] as const;

const filterMenuSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["FilterMenu is a Filters button that opens one checkbox section per filter: a menu on desktop and a full-screen sheet on phones."] },
  whenToUse: { heading: "When to use", content: ["Use FilterMenu to narrow one list or board by a few filters with several values each, such as logs by language, activity and tag."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it to switch between views with their own addresses; use Tabs or Tabbar. When one or two selects fit beside the list, keep them visible instead."] },
  choosingBetween: { heading: "Choosing between", content: ["FilterMenu chooses several values per filter. ActionMenu runs commands. ToggleSelect and RadioSelect choose one value in a form."] },
  anatomy: { heading: "Anatomy", content: ["A ghost trigger reading Filters, or Filters (n) on the selected look. Each section has an uppercase label, a Clear action while it has choices, checkbox rows with the value and its points, and Show all N after the first values. The footer holds the result label and Clear all."] },
  recommendedExample: { heading: "Recommended example", content: ["Log filters starts with Reading chosen. Tick several tags: the menu stays open, the points of the other section and the result label update, and Show all 12 reveals the remaining tags."] },
  variants: { heading: "Variants", content: ["There is one control. From the tablet breakpoint up it is a menu; below it, the same trigger opens a full-screen Drawer with larger rows and a footer with Clear all and Show {resultLabel}."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Values without points are hidden unless chosen, so a chosen value stays listed and checked at 0. The menu lists 6 values per section and the sheet 8 before Show all; chosen values beyond that stay listed. Sections without listed values are left out."] },
  behavior: { heading: "Behavior", content: ["Ticking a value calls onChange with the section's new values and keeps the menu open. Clear calls onChange with no values for its section. Clear all calls onClear, closes the menu or sheet and returns focus to the trigger. Escape and an outside press close the menu and return focus to the trigger. In the sheet choices apply as they are ticked, and Show {resultLabel} closes it."] },
  contentGuidance: { heading: "Content guidance", content: ["Label sections with the filter's noun, such as Language or Tags. Give points the meaning they have under the other filters, and phrase resultLabel as a count with its noun, such as 83 logs."] },
  accessibility: { heading: "Accessibility", content: ["The menu items are menuitemcheckbox elements in groups named by their section labels. The sheet is a modal dialog named Filters with native checkboxes and rows at least 48px tall. The trigger's accessible name includes the count."] },
  implementation: { heading: "Implementation", content: ["Import FilterMenu from paper-ui. The screen owns the selection, keeps it in URL search parameters, filters the results with \"any of\" within a section and \"and\" across sections, computes each value's points under the other filters and resets pagination when a filter changes. Pass a ref to focus the visible trigger after clearing from AppliedFilters."] },
  apiReference: { heading: "API reference", content: ["FilterMenuProps accepts sections ({ id, label, values: { value, label, points }[], selected }), onChange(sectionId, selected), onClear, resultLabel, visibleCount and label. The ref is a FilterMenuHandle with focus()."] },
  relatedPatterns: { heading: "Related patterns", content: ["The Filter bar pattern places FilterMenu in a SectionHeading's actions with AppliedFilters removable chips below it."] },
  migration: { heading: "Migration", content: ["Replace collapsible filter panels of selects with FilterMenu when filters take several values."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

const stackedBarChartSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["StackedBarChart shows points per day across a date range, stacked by series, with an optional running total."] },
  whenToUse: { heading: "When to use", content: ["Use it for activity over a contest or a year when the mix of series per day matters, such as points by activity."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it as the only record of the numbers: a summary, list or table must carry them as text. Use HeatmapChart for which days had activity across a year."] },
  choosingBetween: { heading: "Choosing between", content: ["StackedBarChart compares amounts per day. HeatmapChart shows rhythm in calendar weeks."] },
  anatomy: { heading: "Anatomy", content: ["A points axis with gridlines at zero, half and a round maximum, one bar slot per calendar day, date labels, an optional running-total line on its own right axis, and a legend of the series with points."] },
  recommendedExample: { heading: "Recommended example", content: ["Contest activity draws three months of reading, listening and writing with the running total. Hover a day to read its date, total and series."] },
  variants: { heading: "Variants", content: ["runningTotal adds the cumulative line, its axis and a legend entry. Ranges over 45 days are labelled by month; shorter ranges every few days."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["The chart is two SVG drawings in the markup; the tablet breakpoint shows the desktop or the phone one, so labels stay readable at 390px and nothing is measured. Days without points keep their slot, and a range without points draws empty axes."] },
  behavior: { heading: "Behavior", content: ["StackedBarChart is static. Each day's group has a title the browser shows on hover."] },
  contentGuidance: { heading: "Content guidance", content: ["Name the chart for what it plots and describe the total and best day in description. Pass each series the colour readers know it by, such as the activity colours."] },
  accessibility: { heading: "Accessibility", content: ["The chart is one image named by label and described by description. Keep the same numbers available as text nearby."] },
  implementation: { heading: "Implementation", content: ["Import StackedBarChart from paper-ui and pass ISO dates, days with points per series id, and series with CSS colours such as var(--color-activity-reading)."] },
  apiReference: { heading: "API reference", content: ["StackedBarChartProps accepts start, end, days ({ date, values }[]), series ({ id, label, color }[]), runningTotal, label and description."] },
  relatedPatterns: { heading: "Related patterns", content: ["Pair it with a StatGrid sheet summary above and the breakdown lists or logs it summarizes below."] },
  migration: { heading: "Migration", content: ["Replace client-side chart libraries for daily points with this server-rendered chart."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

export const filtersAndChartsDocuments = [
  componentDocument({
    id: "component.filter-menu",
    route: "/components/forms/filter-menu",
    name: "FilterMenu",
    category: "forms",
    summary: "Checkbox sections that narrow one list: a menu on desktop and a full-screen sheet on phones.",
    keywords: ["filters", "filter menu", "checkbox menu", "facets", "sheet"],
    sourcePath: "src/components/forms/FilterMenu/FilterMenu.tsx",
    fixtureIds: ["filter-menu.log-filters", "pattern.filter-bar"],
    behaviorTestIds: ["filter-menu.multi-select", "filter-menu.dismissal", "filter-menu.zero-points", "filter-menu.show-all", "filter-menu.clear", "filter-menu.sheet"],
    sections: filterMenuSections,
    commonMistakes: ["Hiding a chosen value because its points fell to 0.", "Filtering inside the control instead of the screen.", "Using filters to switch between views."],
    api: {
      react: ["FilterMenu"],
      props: [
        { name: "sections", type: "readonly FilterMenuSection[]", required: true, description: "One section per filter with its values, points and chosen values." },
        { name: "onChange", type: "(sectionId: string, selected: readonly string[]) => void", required: true, description: "Reports a section's chosen values." },
        { name: "onClear", type: "() => void", required: true, description: "Clears every section." },
        { name: "resultLabel", type: "string", required: true, description: "The current result, such as 83 logs." },
        { name: "visibleCount", type: "number", description: "Values before Show all; 6 in the menu and 8 in the sheet by default." },
        { name: "label", type: "string", defaultValue: "Filters", description: "Trigger text and sheet title." },
      ],
      cssClasses: ["paper-filter-menu", "paper-filter-menu__*"],
      publicTypes: ["FilterMenuProps", "FilterMenuSection", "FilterMenuValue", "FilterMenuHandle"],
      defaults: ["label=Filters", "visibleCount=6 in the menu, 8 in the sheet"],
      invalidCombinations: ["Two FilterMenus for one list; put every filter in one menu."],
    },
    dependencies: ["component.drawer", "component.applied-filters", "pattern.filter-bar"],
    legacy: ["webv3 contest profile filter selects"],
  }),
  componentDocument({
    id: "component.stacked-bar-chart",
    route: "/components/data-display/stacked-bar-chart",
    name: "StackedBarChart",
    category: "data-display",
    summary: "Points per day across a date range, stacked by series, with an optional running total.",
    keywords: ["chart", "bar chart", "stacked", "activity", "running total"],
    sourcePath: "src/components/data-display/StackedBarChart/StackedBarChart.tsx",
    fixtureIds: ["stacked-bar-chart.contest-activity"],
    behaviorTestIds: ["stacked-bar-chart.accessible-name", "stacked-bar-chart.day-slots", "stacked-bar-chart.stacking", "stacked-bar-chart.running-total", "stacked-bar-chart.server"],
    sections: stackedBarChartSections,
    commonMistakes: ["Leaving out description, so screen reader users get no numbers.", "Passing series colours that are also link or action colours."],
    api: {
      react: ["StackedBarChart"],
      props: [
        { name: "start / end", type: "string", required: true, description: "ISO dates of the first and last day, inclusive." },
        { name: "days", type: "readonly StackedBarChartDay[]", required: true, description: "Points per series id for each date; missing days count as 0." },
        { name: "series", type: "readonly StackedBarChartSeries[]", required: true, description: "Stacking order, labels and CSS colours." },
        { name: "runningTotal", type: "boolean", defaultValue: "false", description: "Adds the cumulative line on its own axis." },
        { name: "label", type: "string", required: true, description: "Accessible name." },
        { name: "description", type: "string", description: "Text summary for assistive technology." },
      ],
      cssClasses: ["paper-stacked-bar-chart", "paper-stacked-bar-chart__*"],
      publicTypes: ["StackedBarChartProps", "StackedBarChartDay", "StackedBarChartSeries"],
      defaults: ["runningTotal=false"],
      invalidCombinations: [],
    },
    dependencies: ["component.heatmap-chart", "component.stat-grid"],
    legacy: ["webv2 contest profile activity chart"],
  }),
] as const;
