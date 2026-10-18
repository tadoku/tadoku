import DescriptionListExample from "./examples/description-list.contest-facts";
import DescriptionListExampleSource from "./examples/description-list.contest-facts.tsx?raw";
import StatGridExample from "./examples/stat-grid.reading-year";
import StatGridExampleSource from "./examples/stat-grid.reading-year.tsx?raw";
import { type RequiredComponentSections } from "./schema";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture } from "./page-composition-document";

export const statGridAndDescriptionListFixtures = [
  fixture({ id: "stat-grid.reading-year", name: "Reading year figures", description: "Four key figures for one profile year, two with supporting detail.", tags: ["statistics", "profile", "figures"], code: StatGridExampleSource, render: () => <StatGridExample /> }),
  fixture({ id: "description-list.contest-facts", name: "Contest facts", description: "Stacked dates beside an inline summary of the current contest.", tags: ["facts", "description list", "contest"], code: DescriptionListExampleSource, render: () => <DescriptionListExample /> }),
] as const;

const statGridSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["StatGrid presents a few key figures as labelled values in a responsive grid."] },
  whenToUse: { heading: "When to use", content: ["Use StatGrid for two to six headline figures that summarize a page or section, such as a profile year’s personal score, entries, languages and contests."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for facts whose values are sentences or dates; use DescriptionList. Do not use it to compare many records; use Table."] },
  choosingBetween: { heading: "Choosing between", content: ["StatGrid emphasises the value. DescriptionList emphasises the relationship between a label and a longer value."] },
  anatomy: { heading: "Anatomy", content: ["A description list. Each figure has a muted label, a large tabular value and optional detail read with the value, separated from its neighbour by a quiet leading rule."] },
  recommendedExample: { heading: "Recommended example", content: ["Reading year figures shows four profile figures; Entries and Languages add detail that explains them."] },
  variants: { heading: "Variants", content: ["There is one appearance. Detail is optional per figure."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Columns are at least 9rem wide and fill the available width, so four figures become two or one column on narrow screens without breakpoints. Values wrap anywhere rather than overflow."] },
  behavior: { heading: "Behavior", content: ["StatGrid is static and has no keyboard behavior."] },
  contentGuidance: { heading: "Content guidance", content: ["Use short noun labels and formatted values with their units in the label or detail. Show zero rather than hiding a figure when zero is meaningful."] },
  accessibility: { heading: "Accessibility", content: ["Each label is a term and each value, with its detail, is its definition, so assistive technology reads them as pairs."] },
  implementation: { heading: "Implementation", content: ["Import StatGrid from paper-ui and pass formatted values; StatGrid does not format numbers."] },
  apiReference: { heading: "API reference", content: ["StatGridProps accepts stats (an array of { label, value, detail? }) and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Pair StatGrid with SectionHeading, and with HeatmapChart or Table for the detail behind a figure."] },
  migration: { heading: "Migration", content: ["Replace application stat-grid and stat markup, including divs that paired labels and values without list semantics."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

const descriptionListSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["DescriptionList presents named facts about one subject as term and description pairs."] },
  whenToUse: { heading: "When to use", content: ["Use DescriptionList for the facts of one contest, log or service: dates, eligibility, moderators or status."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for headline figures; use StatGrid. Do not use it for records that share columns; use Table. Do not use it for unlabelled metadata lines such as Japanese · Reading · 48 pages."] },
  choosingBetween: { heading: "Choosing between", content: ["Use stacked when values are long or the list sits in a narrow aside. Use inline for short values that readers scan down a column."] },
  anatomy: { heading: "Anatomy", content: ["Each pair is grouped. Stacked places a muted label above its value. Inline places the label at the start and the value at the end of one ruled row."] },
  recommendedExample: { heading: "Recommended example", content: ["Contest facts shows round dates stacked and a current-contest summary inline."] },
  variants: { heading: "Variants", content: ["layout is stacked (default) or inline. Arrange stacked pairs in columns with an application grid on className when the page needs it."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Labels and values wrap anywhere. Inline rows wrap the value below the label when both do not fit."] },
  behavior: { heading: "Behavior", content: ["DescriptionList has no keyboard behavior; links and StatusTags in values keep their own behavior."] },
  contentGuidance: { heading: "Content guidance", content: ["Use short, parallel labels. Keep each value to the answer, such as Japanese, Korean only, and put explanations in nearby text."] },
  accessibility: { heading: "Accessibility", content: ["Pairs render as dt and dd inside a grouping div, which assistive technology reads as term and definition."] },
  implementation: { heading: "Implementation", content: ["Import DescriptionList from paper-ui and pass items in display order."] },
  apiReference: { heading: "API reference", content: ["DescriptionListProps accepts items (an array of { label, value }), layout (stacked or inline) and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use StatusTag in a value for a service or contest state, and SectionHeading to name the list."] },
  migration: { heading: "Migration", content: ["Replace application dl styles for contest facts, eligibility and summary rows."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

export const statGridAndDescriptionListDocuments = [
  componentDocument({
    id: "component.stat-grid",
    route: "/components/data-display/stat-grid",
    name: "StatGrid",
    category: "data-display",
    summary: "Presents a few key figures as labelled values in a responsive grid.",
    keywords: ["statistics", "stats", "figures", "metrics", "summary", "numbers"],
    sourcePath: "src/components/data-display/StatGrid/StatGrid.tsx",
    fixtureIds: ["stat-grid.reading-year"],
    behaviorTestIds: ["stat-grid.description-list", "stat-grid.detail"],
    sections: statGridSections,
    commonMistakes: ["Using StatGrid for long values or dates.", "Hiding units the reader needs to interpret a value."],
    api: {
      react: ["StatGrid"],
      props: [
        { name: "stats", type: "readonly StatGridItem[]", required: true, description: "Figures in display order: { label, value, detail? }." },
        { name: "className", type: "string", description: "Application spacing or rules around the grid." },
      ],
      cssClasses: ["paper-stat-grid", "paper-stat-grid__*"],
      publicTypes: ["StatGridProps", "StatGridItem"],
      defaults: ["Columns are at least 9rem wide."],
      invalidCombinations: ["Values are not formatted by StatGrid."],
    },
    dependencies: ["component.description-list"],
    legacy: ["paper-playground stat-grid, home-ledger-stats, records-profile-stats and records-admin-stats"],
  }),
  componentDocument({
    id: "component.description-list",
    route: "/components/data-display/description-list",
    name: "DescriptionList",
    category: "data-display",
    summary: "Presents named facts about one subject as grouped term and description pairs.",
    keywords: ["description list", "facts", "key value", "dl", "details"],
    sourcePath: "src/components/data-display/DescriptionList/DescriptionList.tsx",
    fixtureIds: ["description-list.contest-facts"],
    behaviorTestIds: ["description-list.structure", "description-list.linked-values"],
    sections: descriptionListSections,
    commonMistakes: ["Using a description list for unlabelled metadata.", "Choosing inline for long values that then wrap awkwardly."],
    api: {
      react: ["DescriptionList"],
      props: [
        { name: "items", type: "readonly DescriptionListItem[]", required: true, description: "Facts in display order: { label, value }." },
        { name: "layout", type: "\"stacked\" | \"inline\"", defaultValue: "stacked", description: "Label above value, or label and value on one ruled row." },
        { name: "className", type: "string", description: "Application grid or spacing, such as two stacked columns." },
      ],
      cssClasses: ["paper-description-list", "paper-description-list--stacked", "paper-description-list--inline", "paper-description-list__*"],
      publicTypes: ["DescriptionListProps", "DescriptionListItem"],
      defaults: ["layout=stacked"],
      invalidCombinations: [],
    },
    dependencies: ["component.stat-grid", "component.table"],
    legacy: ["paper-playground contest-facts, contest-eligibility, records-admin-contest and records-activity-mix"],
  }),
] as const;
