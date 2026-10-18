import EmptyStateFirstActivity from "./examples/empty-state.first-activity";
import EmptyStateFirstActivitySource from "./examples/empty-state.first-activity.tsx?raw";
import EmptyStateNoResults from "./examples/empty-state.no-results";
import EmptyStateNoResultsSource from "./examples/empty-state.no-results.tsx?raw";
import StatusTagExample from "./examples/status-tag.tones";
import StatusTagExampleSource from "./examples/status-tag.tones.tsx?raw";
import { type RequiredComponentSections } from "./schema";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture } from "./page-composition-document";

export const statusTagAndEmptyStateFixtures = [
  fixture({ id: "status-tag.tones", name: "Status tones", description: "Every tone with the visible text that carries its meaning.", tags: ["status", "tag", "tones"], code: StatusTagExampleSource, render: () => <StatusTagExample /> }),
  fixture({ id: "empty-state.no-results", name: "Search without results", description: "A standalone empty state whose action clears the search and returns focus to it.", tags: ["empty", "search", "recovery"], code: EmptyStateNoResultsSource, render: () => <EmptyStateNoResults /> }),
  fixture({ id: "empty-state.first-activity", name: "Empty year", description: "A plain empty state inside an existing section, offering the first entry.", tags: ["empty", "plain", "first use"], code: EmptyStateFirstActivitySource, render: () => <EmptyStateFirstActivity /> }),
] as const;

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

export const statusTagAndEmptyStateDocuments = [
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
] as const;
