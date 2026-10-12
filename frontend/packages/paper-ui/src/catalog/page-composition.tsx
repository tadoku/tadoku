import { type ReactNode } from "react";
import DescriptionListExample from "./examples/description-list.contest-facts";
import DescriptionListExampleSource from "./examples/description-list.contest-facts.tsx?raw";
import EmptyStateFirstActivity from "./examples/empty-state.first-activity";
import EmptyStateFirstActivitySource from "./examples/empty-state.first-activity.tsx?raw";
import EmptyStateNoResults from "./examples/empty-state.no-results";
import EmptyStateNoResultsSource from "./examples/empty-state.no-results.tsx?raw";
import PageHeaderExample from "./examples/page-header.contest-detail";
import PageHeaderExampleSource from "./examples/page-header.contest-detail.tsx?raw";
import FilterBarExample from "./examples/pattern.filter-bar";
import FilterBarExampleSource from "./examples/pattern.filter-bar.tsx?raw";
import SectionHeadingExample from "./examples/section-heading.profile";
import SectionHeadingExampleSource from "./examples/section-heading.profile.tsx?raw";
import StatGridExample from "./examples/stat-grid.reading-year";
import StatGridExampleSource from "./examples/stat-grid.reading-year.tsx?raw";
import StatusTagExample from "./examples/status-tag.tones";
import StatusTagExampleSource from "./examples/status-tag.tones.tsx?raw";
import {
  COMPONENT_PAGE_SECTION_KEYS,
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-10-12";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;
const EXPERIMENTAL_LIFECYCLE =
  "Experimental in Paper 0.1.0. Extracted from repeated paper-playground screens; no production application consumes it yet.";

function fixture(input: {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly tags: readonly string[];
  readonly code: string;
  readonly render: () => ReactNode;
}) {
  return defineCatalogFixture({
    ...input,
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
  });
}

type ComponentInput = Pick<
  CatalogDocument,
  "id" | "route" | "name" | "category" | "summary" | "keywords" | "sourcePath" | "fixtureIds" | "behaviorTestIds"
> & {
  readonly sections: RequiredComponentSections;
  readonly commonMistakes: readonly string[];
  readonly api: CatalogDocument["api"];
  readonly dependencies?: readonly string[];
  readonly legacy: readonly string[];
};

function componentDocument({
  sections,
  commonMistakes,
  dependencies = [],
  legacy,
  ...input
}: ComponentInput): CatalogDocument {
  return defineCatalogDocument({
    ...input,
    kind: "component",
    aliases: [],
    lifecycle: "Experimental",
    reviewDate: REVIEW_DATE,
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: sections.whenToUse.content,
      whenNotToUse: sections.whenNotToUse.content,
      content: sections.contentGuidance.content,
      commonMistakes,
    },
    accessibility: {
      requirements: sections.accessibility.content,
      keyboard: sections.behavior.content,
      knownConstraints: sections.statesAndAdaptation.content,
    },
    dependencies: { documents: dependencies, packages: ["paper-ui"] },
    migration: { legacy, notes: sections.migration.content },
    changelog: [{ date: REVIEW_DATE, note: "Added as an Experimental page-composition control." }],
    sections: { required: sections, pageSections: COMPONENT_PAGE_SECTION_KEYS },
  });
}

export const pageCompositionFixtures = [
  fixture({ id: "page-header.contest-detail", name: "Contest page header", description: "Back link, title, status tags, lead and page actions for one contest.", tags: ["page header", "status", "actions"], code: PageHeaderExampleSource, render: () => <PageHeaderExample /> }),
  fixture({ id: "section-heading.profile", name: "Profile sections", description: "A level-two section with a description and action, and a nested level-three subsection.", tags: ["section heading", "outline", "actions"], code: SectionHeadingExampleSource, render: () => <SectionHeadingExample /> }),
  fixture({ id: "status-tag.tones", name: "Status tones", description: "Every tone with the visible text that carries its meaning.", tags: ["status", "tag", "tones"], code: StatusTagExampleSource, render: () => <StatusTagExample /> }),
  fixture({ id: "empty-state.no-results", name: "Search without results", description: "A standalone empty state whose action clears the search and returns focus to it.", tags: ["empty", "search", "recovery"], code: EmptyStateNoResultsSource, render: () => <EmptyStateNoResults /> }),
  fixture({ id: "empty-state.first-activity", name: "Empty year", description: "A plain empty state inside an existing section, offering the first entry.", tags: ["empty", "plain", "first use"], code: EmptyStateFirstActivitySource, render: () => <EmptyStateFirstActivity /> }),
  fixture({ id: "stat-grid.reading-year", name: "Reading year figures", description: "Four key figures for one profile year, two with supporting detail.", tags: ["statistics", "profile", "figures"], code: StatGridExampleSource, render: () => <StatGridExample /> }),
  fixture({ id: "description-list.contest-facts", name: "Contest facts", description: "Stacked dates beside an inline summary of the current contest.", tags: ["facts", "description list", "contest"], code: DescriptionListExampleSource, render: () => <DescriptionListExample /> }),
  fixture({ id: "pattern.filter-bar", name: "Standings filter bar", description: "Collapsible filters, applied-filter summary, result count and a filtered empty state.", tags: ["filters", "pattern", "standings"], code: FilterBarExampleSource, render: () => <FilterBarExample /> }),
] as const;

const pageHeaderSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["PageHeader names a page with its only level-one heading and gathers the context and actions that apply to the whole page."] },
  whenToUse: { heading: "When to use", content: ["Use PageHeader at the top of every application page: a title, an optional lead sentence, an optional way back, status for the page’s subject and the page’s main actions."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for sections inside a page; use SectionHeading. Do not use it for a promotional hero whose layout is part of the page design, or for a page that is only an empty or error state; use EmptyState with headingLevel 1."] },
  choosingBetween: { heading: "Choosing between", content: ["PageHeader owns the page’s h1. SectionHeading starts at h2 for regions within the page. Breadcrumb shows a hierarchy; the back link returns to the single place the reader most likely came from."] },
  anatomy: { heading: "Anatomy", content: ["In reading order: an optional back link with a leading chevron, the title, optional status tags beside it, an optional lead, optional supporting content such as a step indicator, then actions. Actions sit at the end of the title row on wide screens and wrap below the text when space runs out."] },
  recommendedExample: { heading: "Recommended example", content: ["The contest header returns to the official contests, names the round, shows that it is live and open for registration, and offers joining as the primary action with standings as the secondary one."] },
  variants: { heading: "Variants", content: ["Every part except the title is optional. Pass a lead when the title alone does not explain the page. Pass children for page-level supporting content that must stay inside the header."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Long titles wrap anywhere rather than overflow. Status tags wrap beside or below the title. Actions wrap below the text at narrow widths. The title keeps the page type role in both densities; density changes only the actions’ control height."] },
  behavior: { heading: "Behavior", content: ["The back link, actions and any links in the content follow native link and button behavior in reading order. PageHeader has no interactive state of its own."] },
  contentGuidance: { heading: "Content guidance", content: ["Use the subject’s name as the title, such as a contest or log title. Keep the lead to one or two sentences. Label the back link with its destination, such as Back to official contests. Put one default-variant action first when there is a primary task."] },
  accessibility: { heading: "Accessibility", content: ["Render exactly one PageHeader per page so the page has one h1. Status tags render outside the heading so the heading’s name stays the subject’s name. The chevron is decorative; the back link’s text is its name."] },
  implementation: { heading: "Implementation", content: ["Import PageHeader from paper-ui. Pass renderLink to render the back link with your router, for example renderLink={({ href, ...props }) => <Link to={href} {...props} />}; the renderer must forward className and children. Without it a native link is rendered."] },
  apiReference: { heading: "API reference", content: ["PageHeaderProps accepts title, lead, status, actions, backLink ({ href, label }), renderLink, children and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use StatusTag inside status, Button or buttonClassName links inside actions, and SectionHeading for each region that follows."] },
  migration: { heading: "Migration", content: ["Replace application page-header, page-lead and back-link markup and styles. Keep page-specific title typography, such as a display-size guide title, as an application class on PageHeader."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

const sectionHeadingSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["SectionHeading titles a region of a page at an explicit heading level, with an optional description, status and actions."] },
  whenToUse: { heading: "When to use", content: ["Use SectionHeading at the start of each named region of a page, such as Contest history with an All contests link or Activity log with a Filters toggle."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for the page title; PageHeader owns the h1. Do not use it for list-item titles inside a row or card, where heading spacing and actions would crowd the item."] },
  choosingBetween: { heading: "Choosing between", content: ["Choose the level from the document outline, not the visual size. Level 2 uses the section type role; levels 3 to 6 use the component type role."] },
  anatomy: { heading: "Anatomy", content: ["A heading with optional status tags beside it and an optional description below, followed by actions at the end of the row. Actions wrap below the text when the row is too narrow. The heading ends with 1rem of space; inside paper-stack the stack gap replaces it."] },
  recommendedExample: { heading: "Recommended example", content: ["Anton’s activity is a level-two section with a description and an action. Immersion rhythm is a level-three subsection inside it, with a link to the scoring guide."] },
  variants: { heading: "Variants", content: ["The level selects the element and type role. Description, status and actions are independent and optional."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Long titles wrap anywhere. Actions wrap below the text before they overflow. The heading’s hierarchy does not change with density."] },
  behavior: { heading: "Behavior", content: ["SectionHeading has no interactive state. Actions keep native link and button behavior and follow the heading in the tab order."] },
  contentGuidance: { heading: "Content guidance", content: ["Use a short noun phrase for the title. Use the description for scope, such as a date range, not for instructions. Keep actions to the one or two tasks that apply to the whole section."] },
  accessibility: { heading: "Accessibility", content: ["Pass id and label the enclosing section with aria-labelledby so it becomes a named region. Do not skip levels. Status tags stay outside the heading’s accessible name."] },
  implementation: { heading: "Implementation", content: ["Import SectionHeading from paper-ui and pass level explicitly; TypeScript rejects level 1."] },
  apiReference: { heading: "API reference", content: ["SectionHeadingProps accepts level (2–6), title, id, description, status, actions and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use PageHeader above it, StatusTag for status, and the Filter bar pattern for a Filters toggle in actions."] },
  migration: { heading: "Migration", content: ["Replace application section-heading wrappers that pair a heading with a link or buttons. Remove heading font overrides that only reproduce the section or component type roles."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

const statusTagSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["StatusTag labels the current state of a contest, record or service with short visible text and a tone."] },
  whenToUse: { heading: "When to use", content: ["Use StatusTag beside a title or in a row to show a state that is true right now, such as Live now, Unlisted, Draft or Published."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for messages that need explanation or action; use Flash. Do not use it for counts, categories or filters, and do not make it interactive."] },
  choosingBetween: { heading: "Choosing between", content: ["StatusTag is a static label. Flash explains a change or problem. Toast confirms an action that just happened."] },
  anatomy: { heading: "Anatomy", content: ["A single line of text inside a quiet rule. Tone colours the text and, except for neutral, the rule."] },
  recommendedExample: { heading: "Recommended example", content: ["The example shows every tone with a realistic Tadoku state. Read each label without its colour: the text alone must still carry the meaning."] },
  variants: { heading: "Variants", content: ["Neutral is the default for descriptive states such as Unlisted. Information marks scheduled or upcoming states, success marks a live or healthy state, warning marks a state that needs attention soon, and danger marks an unavailable or failed state."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Tags do not wrap internally; a row of tags wraps between tags. Tones use the semantic status colours of the active light or dark theme."] },
  behavior: { heading: "Behavior", content: ["StatusTag has no keyboard behavior and is not a live region. Announce a change in status through the screen’s own status message."] },
  contentGuidance: { heading: "Content guidance", content: ["Use one to three words in sentence case, such as Live now or Registration closed. Prefer the reader’s perspective over system terms."] },
  accessibility: { heading: "Accessibility", content: ["Visible text is required by the API, so status never relies on colour alone. Do not add role=\"status\" to a static tag."] },
  implementation: { heading: "Implementation", content: ["Import StatusTag from paper-ui and pass the label as its text child. Non-React consumers can use the paper-status-tag class with one paper-status-tag--{tone} modifier."] },
  apiReference: { heading: "API reference", content: ["StatusTagProps accepts tone (neutral, information, success, warning or danger; default neutral), the required string children and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Pass StatusTag to the status slot of PageHeader or SectionHeading, or place it in a Table cell or DescriptionList value."] },
  migration: { heading: "Migration", content: ["Replace application status-tag spans and the legacy admin StatusBadge colour classes with StatusTag tones: Draft as warning, Scheduled as information and Published as success."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

const emptyStateSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["EmptyState explains why there is nothing to show and offers the next useful step."] },
  whenToUse: { heading: "When to use", content: ["Use EmptyState when a collection has no items yet, when filters or a search exclude every item, when data could not be loaded, or when a page cannot be shown, such as a missing contest."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for loading; use Loading. Do not use it for a failure that leaves the rest of the content usable; use Flash. Inside Table, prefer its emptyMessage while the table structure still helps."] },
  choosingBetween: { heading: "Choosing between", content: ["Use the surface variant when the state stands alone in the page. Use plain inside an existing Surface, sheet or dialog that already provides a boundary."] },
  anatomy: { heading: "Anatomy", content: ["A heading, an optional description and optional actions. The heading level is required because the state replaces content at a particular place in the outline."] },
  recommendedExample: { heading: "Recommended example", content: ["Search without results clears the search and returns focus to the search field. Empty year places a plain state inside a profile section and offers the first entry."] },
  variants: { heading: "Variants", content: ["surface (default) adds paper, a rule and padding. plain keeps only vertical spacing. headingLevel 1 uses the page type role, 2 the section role and 3 to 6 the component role."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Content stacks and actions wrap at narrow widths; the description keeps a readable measure. Pass role=\"alert\" for a failure the reader caused or must notice, and aria-live=\"polite\" when results can become empty while the reader types."] },
  behavior: { heading: "Behavior", content: ["Actions follow native link and button behavior. After an action removes the content, pass headingRef and focus the heading so the reader learns what happened; the heading is focusable only when headingRef is supplied."] },
  contentGuidance: { heading: "Content guidance", content: ["Title the state plainly, such as No activity yet or This log couldn’t be loaded. Use the description for the reason and what will make content appear. Label the action with the outcome, such as Clear filters or Log activity."] },
  accessibility: { heading: "Accessibility", content: ["Choose headingLevel from the surrounding outline. Use role or aria-live only for states that appear in response to the reader. Return focus to the control that can change the result after clearing."] },
  implementation: { heading: "Implementation", content: ["Import EmptyState from paper-ui. The screen decides which state applies and supplies the actions; EmptyState does not fetch, retry or check permissions."] },
  apiReference: { heading: "API reference", content: ["EmptyStateProps accepts title, headingLevel (1–6), description, actions, variant (surface or plain), headingRef and native div attributes such as role and aria-live."] },
  relatedPatterns: { heading: "Related patterns", content: ["The Filter bar pattern uses a plain EmptyState for filtered results. Use PageHeader instead when the page still has content."] },
  migration: { heading: "Migration", content: ["Replace application empty-state boxes and their heading and paragraph overrides. Move any focus target to headingRef."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

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
  componentDocument({
    id: "component.page-header",
    route: "/components/page-structure/page-header",
    name: "PageHeader",
    category: "page-structure",
    summary: "Names a page with its only h1, with an optional back link, lead, status and page actions.",
    keywords: ["page header", "title", "h1", "back link", "lead", "actions"],
    sourcePath: "src/components/page-structure/PageHeader/PageHeader.tsx",
    fixtureIds: ["page-header.contest-detail"],
    behaviorTestIds: ["page-header.single-h1", "page-header.status-outside-name", "page-header.back-link-renderer", "page-header.actions-keyboard"],
    sections: pageHeaderSections,
    commonMistakes: ["Rendering a second h1 for a page section.", "Placing status text inside the heading, which changes its accessible name."],
    api: {
      react: ["PageHeader"],
      props: [
        { name: "title", type: "ReactNode", required: true, description: "The page’s only level-one heading." },
        { name: "lead", type: "ReactNode", description: "One or two sentences that explain the page." },
        { name: "status", type: "ReactNode", description: "StatusTags beside the title, outside its accessible name." },
        { name: "actions", type: "ReactNode", description: "Page-level links or buttons at the end of the title row." },
        { name: "backLink", type: "{ href: string; label: string }", description: "The destination the reader most likely came from." },
        { name: "renderLink", type: "NavigationLinkRenderer", defaultValue: "native <a>", description: "Renders the back link with the application router; forward className and children." },
        { name: "children", type: "ReactNode", description: "Supporting content after the lead, such as a step indicator." },
      ],
      cssClasses: ["paper-page-header", "paper-page-header__*"],
      publicTypes: ["PageHeaderProps", "NavigationLinkRenderer"],
      defaults: ["Renders an h1 with the page type role and 2rem of space after the header."],
      invalidCombinations: ["Use one PageHeader per page."],
    },
    dependencies: ["component.section-heading", "component.status-tag", "component.breadcrumb"],
    legacy: ["paper-playground page-header, page-lead and page-back-link"],
  }),
  componentDocument({
    id: "component.section-heading",
    route: "/components/page-structure/section-heading",
    name: "SectionHeading",
    category: "page-structure",
    summary: "Titles a page region at an explicit heading level, with optional description, status and actions.",
    keywords: ["section heading", "heading level", "h2", "h3", "outline", "actions"],
    sourcePath: "src/components/page-structure/SectionHeading/SectionHeading.tsx",
    fixtureIds: ["section-heading.profile", "description-list.contest-facts"],
    behaviorTestIds: ["section-heading.level", "section-heading.region-label", "section-heading.status-outside-name", "section-heading.actions-keyboard"],
    sections: sectionHeadingSections,
    commonMistakes: ["Choosing a level for its size instead of its place in the outline.", "Forgetting aria-labelledby, which leaves the section unnamed."],
    api: {
      react: ["SectionHeading"],
      props: [
        { name: "level", type: "2 | 3 | 4 | 5 | 6", required: true, description: "Heading element. Level 2 uses the section type role; 3–6 use the component role." },
        { name: "title", type: "ReactNode", required: true, description: "Short name of the region." },
        { name: "id", type: "string", description: "Applied to the heading so the section can use aria-labelledby." },
        { name: "description", type: "ReactNode", description: "Muted scope line below the title." },
        { name: "status", type: "ReactNode", description: "StatusTags beside the title, outside its accessible name." },
        { name: "actions", type: "ReactNode", description: "Links or buttons that apply to the whole section." },
      ],
      cssClasses: ["paper-section-heading", "paper-section-heading__*"],
      publicTypes: ["SectionHeadingProps"],
      defaults: ["1rem of space after the heading; paper-stack gaps replace it."],
      invalidCombinations: ["level 1 is rejected; use PageHeader."],
    },
    dependencies: ["component.page-header", "component.status-tag"],
    legacy: ["paper-playground section-heading and records-section-heading"],
  }),
  componentDocument({
    id: "component.status-tag",
    route: "/components/feedback/status-tag",
    name: "StatusTag",
    category: "feedback",
    summary: "Labels the current state of a contest, record or service with visible text and a status tone.",
    keywords: ["status", "tag", "badge", "label", "tone", "live"],
    sourcePath: "src/components/feedback/StatusTag/StatusTag.tsx",
    fixtureIds: ["status-tag.tones"],
    behaviorTestIds: ["status-tag.visible-text", "status-tag.not-live-region", "status-tag.recipe"],
    sections: statusTagSections,
    commonMistakes: ["Using colour without a meaningful label.", "Turning a tag into a filter or button."],
    api: {
      react: ["StatusTag"],
      props: [
        { name: "children", type: "string", required: true, description: "Visible status text. Required so status never relies on colour alone." },
        { name: "tone", type: "\"neutral\" | \"information\" | \"success\" | \"warning\" | \"danger\"", defaultValue: "neutral", description: "Maps to the semantic status colours of the active theme." },
      ],
      cssClasses: ["paper-status-tag", "paper-status-tag--neutral", "paper-status-tag--information", "paper-status-tag--success", "paper-status-tag--warning", "paper-status-tag--danger"],
      publicTypes: ["StatusTagProps", "StatusTagTone"],
      defaults: ["tone=neutral"],
      invalidCombinations: ["A tag without text does not compile."],
    },
    dependencies: ["component.flash"],
    legacy: ["paper-playground status-tag", "admin StatusBadge"],
  }),
  componentDocument({
    id: "component.empty-state",
    route: "/components/feedback/empty-state",
    name: "EmptyState",
    category: "feedback",
    summary: "Explains why there is nothing to show and offers the next useful step.",
    keywords: ["empty", "no results", "not found", "error", "zero state", "first use"],
    sourcePath: "src/components/feedback/EmptyState/EmptyState.tsx",
    fixtureIds: ["empty-state.no-results", "empty-state.first-activity", "pattern.filter-bar"],
    behaviorTestIds: ["empty-state.heading-level", "empty-state.actions", "empty-state.alert", "empty-state.focus-heading"],
    sections: emptyStateSections,
    commonMistakes: ["Using a fixed heading level that breaks the page outline.", "Leaving focus on an action that disappears after clearing."],
    api: {
      react: ["EmptyState"],
      props: [
        { name: "title", type: "ReactNode", required: true, description: "Plain statement of the state." },
        { name: "headingLevel", type: "1 | 2 | 3 | 4 | 5 | 6", required: true, description: "Place in the outline; 1 when the state replaces the page." },
        { name: "description", type: "ReactNode", description: "Reason and what will make content appear." },
        { name: "actions", type: "ReactNode", description: "Recovery or first-use actions." },
        { name: "variant", type: "\"surface\" | \"plain\"", defaultValue: "surface", description: "Standalone paper surface, or plain inside an existing container." },
        { name: "headingRef", type: "Ref<HTMLHeadingElement>", description: "Makes the heading programmatically focusable for focus moves." },
        { name: "role / aria-live / className", type: "HTMLAttributes<HTMLDivElement>", description: "Announce failures with role=\"alert\" or result changes with aria-live." },
      ],
      cssClasses: ["paper-empty-state", "paper-empty-state--surface", "paper-empty-state--plain", "paper-empty-state__*"],
      publicTypes: ["EmptyStateProps"],
      defaults: ["variant=surface"],
      invalidCombinations: ["headingLevel is required."],
    },
    dependencies: ["component.loading", "component.flash", "component.table"],
    legacy: ["paper-playground empty-state, records-empty-year and records-submission-empty"],
  }),
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
    dependencies: { documents: ["component.applied-filters", "component.section-heading", "component.empty-state", "component.button", "component.select", "component.input", "component.table"], packages: ["paper-ui", "react-hook-form"] },
    migration: { legacy: ["paper-playground standings and activity-log filters"], notes: ["Keep each screen’s filter state and field layout; replace hand-built summaries with AppliedFilters and empty results with EmptyState."] },
    changelog: [{ date: REVIEW_DATE, note: "Documented from the paper-playground standings and activity-log filters." }],
    behaviorTestIds: ["pattern.filter-bar.summary", "pattern.filter-bar.empty"],
  }),
] as const;
