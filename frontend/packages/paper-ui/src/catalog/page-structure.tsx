import PageHeaderExample from "./examples/page-header.contest-detail";
import PageHeaderExampleSource from "./examples/page-header.contest-detail.tsx?raw";
import SectionHeadingExample from "./examples/section-heading.profile";
import SectionHeadingExampleSource from "./examples/section-heading.profile.tsx?raw";
import { type RequiredComponentSections } from "./schema";
import { componentDocument, EXPERIMENTAL_LIFECYCLE, fixture } from "./page-composition-document";

export const pageStructureFixtures = [
  fixture({ id: "page-header.contest-detail", name: "Contest page header", description: "Back link, title, status tags, lead and page actions for one contest.", tags: ["page header", "status", "actions"], code: PageHeaderExampleSource, render: () => <PageHeaderExample /> }),
  fixture({ id: "section-heading.profile", name: "Profile sections", description: "A level-two section with a description and action, and a nested level-three subsection.", tags: ["section heading", "outline", "actions"], code: SectionHeadingExampleSource, render: () => <SectionHeadingExample /> }),
] as const;

const pageHeaderSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["PageHeader names a page with its only level-one heading and gathers the context and actions that apply to the whole page."] },
  whenToUse: { heading: "When to use", content: ["Use PageHeader at the top of every application page: a title, an optional lead sentence, an optional way back, status for the page’s subject and the page’s main actions."] },
  whenNotToUse: { heading: "When not to use", content: ["Do not use it for sections inside a page; use SectionHeading. Do not use it for a promotional hero whose layout is part of the page design, or for a page that is only an empty or error state; use EmptyState with headingLevel 1."] },
  choosingBetween: { heading: "Choosing between", content: ["PageHeader owns the page’s h1. SectionHeading starts at h2 for regions within the page. Breadcrumb shows a hierarchy; the back link returns to the single place the reader most likely came from."] },
  anatomy: { heading: "Anatomy", content: ["In reading order: an optional back link with a leading chevron, the title, optional status tags beside it, an optional lead, optional supporting content such as a step indicator, then actions. Actions sit at the end of the title row when the text and actions fit, and wrap below the text when space runs out."] },
  recommendedExample: { heading: "Recommended example", content: ["The contest header returns to the official contests, names the round, shows that it is live and open for registration, and offers joining as the primary action with standings as the secondary one."] },
  variants: { heading: "Variants", content: ["Every part except the title is optional. Pass a lead when the title alone does not explain the page. Pass children for page-level supporting content that must stay inside the header."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["At every width, the text keeps its natural width and actions wrap below it when the whole row cannot fit. Status tags stay beside the title when they fit; otherwise they wrap below it. Long titles wrap rather than overflow, with individual words breaking only when necessary. The title keeps the page type role in both densities; density changes only the actions’ control height."] },
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
  statesAndAdaptation: { heading: "States and adaptation", content: ["At every width, the text keeps its natural width and actions wrap below it when the whole row cannot fit. Status tags stay beside the title when they fit; otherwise they wrap below it. Long titles wrap rather than overflow, with individual words breaking only when necessary. The heading’s hierarchy does not change with density."] },
  behavior: { heading: "Behavior", content: ["SectionHeading has no interactive state. Actions keep native link and button behavior and follow the heading in the tab order."] },
  contentGuidance: { heading: "Content guidance", content: ["Use a short noun phrase for the title. Use the description for scope, such as a date range, not for instructions. Keep actions to the one or two tasks that apply to the whole section."] },
  accessibility: { heading: "Accessibility", content: ["Pass id and label the enclosing section with aria-labelledby so it becomes a named region. Do not skip levels. Status tags stay outside the heading’s accessible name."] },
  implementation: { heading: "Implementation", content: ["Import SectionHeading from paper-ui and pass level explicitly; TypeScript rejects level 1."] },
  apiReference: { heading: "API reference", content: ["SectionHeadingProps accepts level (2–6), title, id, description, status, actions and className."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use PageHeader above it, StatusTag for status, and the Filter bar pattern for a Filters toggle in actions."] },
  migration: { heading: "Migration", content: ["Replace application section-heading wrappers that pair a heading with a link or buttons. Remove heading font overrides that only reproduce the section or component type roles."] },
  lifecycle: { heading: "Lifecycle", content: [EXPERIMENTAL_LIFECYCLE] },
};

export const pageStructureDocuments = [
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
] as const;
