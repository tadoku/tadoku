import type { ReactNode } from "react";
import { Button } from "../components/actions/Button";
import { Drawer } from "../components/overlays/Drawer";
import { Tabs } from "../components/navigation/Tabs";
import {
  COMPONENT_PAGE_SECTION_KEYS,
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type ComponentDocumentationSections,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

function sections(required: RequiredComponentSections): ComponentDocumentationSections {
  return { required, pageSections: COMPONENT_PAGE_SECTION_KEYS };
}

function document(
  input: Omit<
    CatalogDocument,
    "kind" | "aliases" | "lifecycle" | "reviewDate" | "packageVersion" |
      "dependencies" | "changelog"
  >,
): CatalogDocument {
  return defineCatalogDocument({
    ...input,
    kind: "component",
    aliases: [],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    packageVersion: "0.1.0",
    dependencies: { documents: [], packages: ["paper-ui"] },
    changelog: [{ date: REVIEW_DATE, note: "Published the Paper primitive." }],
  });
}

function fixture(
  input: {
    readonly id: string;
    readonly name: string;
    readonly description: string;
    readonly tags: readonly string[];
    readonly code: string;
    readonly render: () => ReactNode;
  },
) {
  return defineCatalogFixture({
    ...input,
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
  });
}

export const publishedPrimitiveFixtures = [
  fixture({
    id: "tabs.content",
    name: "Reading log views",
    description: "Peer views of one reading log with automatic keyboard selection.",
    tags: ["tabs", "content", "keyboard", "disabled"],
    code: `import { Tabs } from "paper-ui";

<Tabs.Root defaultValue="summary">
  <Tabs.List aria-label="Reading log views">
    <Tabs.Tab value="summary">Summary</Tabs.Tab>
    <Tabs.Tab value="entries">Entries</Tabs.Tab>
    <Tabs.Tab value="moderation" disabled>Moderation</Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel value="summary">1,240 pages read in Japanese.</Tabs.Panel>
  <Tabs.Panel value="entries">12 reading entries.</Tabs.Panel>
  <Tabs.Panel value="moderation">No moderation notes.</Tabs.Panel>
</Tabs.Root>`,
    render: () => (
      <Tabs.Root defaultValue="summary">
        <Tabs.List aria-label="Reading log views">
          <Tabs.Tab value="summary">Summary</Tabs.Tab>
          <Tabs.Tab value="entries">Entries</Tabs.Tab>
          <Tabs.Tab value="moderation" disabled>Moderation</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="summary">1,240 pages read in Japanese.</Tabs.Panel>
        <Tabs.Panel value="entries">12 reading entries.</Tabs.Panel>
        <Tabs.Panel value="moderation">No moderation notes.</Tabs.Panel>
      </Tabs.Root>
    ),
  }),
  fixture({
    id: "drawer.filters",
    name: "Reading-entry filters",
    description: "A supporting filter task that preserves the entries page context.",
    tags: ["drawer", "filters", "footer", "end placement"],
    code: `import { Button, Drawer } from "paper-ui";

<Drawer
  trigger={<Button variant="outline">Review filters</Button>}
  title="Entry filters"
  description="Narrow the reading entries shown on this page."
  footer={<Button>Apply filters</Button>}
>
  <p>Choose a language, contest, or date range.</p>
</Drawer>`,
    render: () => (
      <Drawer
        trigger={<Button variant="outline">Review filters</Button>}
        title="Entry filters"
        description="Narrow the reading entries shown on this page."
        footer={<Button>Apply filters</Button>}
      >
        <p>Choose a language, contest, or date range.</p>
      </Drawer>
    ),
  }),
] as const;

const tabsSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["Tabs switches among peer content views while keeping one view active."] },
  whenToUse: { heading: "When to use", content: ["Use Tabs to switch between peer views of the same local context, such as the summary and entries for one reading log."] },
  whenNotToUse: { heading: "When not to use", content: ["Use Tabbar for linked destinations with their own URLs, and use ordinary headings when readers should see every section together."] },
  choosingBetween: { heading: "Choose between", content: ["Tabs changes content in place. Tabbar and VerticalTabbar navigate to linked destinations and preserve native link behavior."] },
  anatomy: { heading: "Anatomy", content: ["A Tabs root contains one labelled list, two or more tabs, and one panel for each value."] },
  recommendedExample: { heading: "Recommended example", content: ["Reading log views keeps each tab next to its matching panel and includes a real disabled state."] },
  variants: { heading: "Variants", content: ["Use horizontal orientation by default; reserve vertical orientation for layouts with enough width for a stable side rail."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Selected, focused, hovered, and disabled tabs retain distinct non-color cues. Horizontal lists scroll at narrow widths."] },
  behavior: { heading: "Behavior", content: ["Arrow keys move focus and select automatically; Home and End move to the first and last enabled tabs. Inactive panels stay mounted by default so local state is preserved."] },
  contentGuidance: { heading: "Content guidance", content: ["Use short parallel nouns such as Summary and Entries. The tab-list label should name the set, not repeat the word tabs."] },
  accessibility: { heading: "Accessibility", content: ["Give Tabs.List an accessible label, keep a panel for every tab value, and do not place linked page navigation in a tab role."] },
  implementation: { heading: "Implementation", content: ["Import the compound Tabs API or the named TabsRoot, TabsList, TabsTab, and TabsPanel exports from paper-ui."] },
  apiReference: { heading: "API reference", content: ["TabsRootProps controls value and orientation; TabsListProps controls activation and focus looping; TabsTabProps and TabsPanelProps share string values."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use Tabbar for URL-backed destinations and Drawer for supporting work that temporarily covers the current view."] },
  migration: { heading: "Migration", content: ["Replace application-owned role, aria-selected, roving-tabindex, and panel visibility logic with the matching Tabs parts."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0 with horizontal, vertical, controlled, disabled, and mounted-panel behavior covered."] },
};

const drawerSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["Drawer presents a modal side sheet without discarding the page behind it."] },
  whenToUse: { heading: "When to use", content: ["Use Drawer for a supporting task such as filters, compact editing, or contextual navigation that should preserve the current page."] },
  whenNotToUse: { heading: "When not to use", content: ["Use Modal for a small focused decision and a dedicated page for long, multi-step, or linkable work."] },
  choosingBetween: { heading: "Choose between", content: ["Drawer allows more vertical content and page context than Modal; Sidebar remains persistent navigation rather than a modal task surface."] },
  anatomy: { heading: "Anatomy", content: ["An application-owned trigger opens a titled sheet with an optional description, scrolling body, footer, and labelled close action."] },
  recommendedExample: { heading: "Recommended example", content: ["Reading-entry filters keeps the page visible, explains the filter scope, and places the commit action in the footer."] },
  variants: { heading: "Variants", content: ["Placement is start or end in logical reading direction. Prefer end for supporting tasks unless navigation hierarchy calls for start."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["The sheet is viewport-bounded, its body scrolls independently, and reduced-motion preferences remove transitions."] },
  behavior: { heading: "Behavior", content: ["Base UI contains focus, blocks background interaction, closes on Escape or outside press, and restores focus to the trigger. Controlled and uncontrolled open state are supported."] },
  contentGuidance: { heading: "Content guidance", content: ["Use a brief noun title, add a description only when it clarifies scope, and keep persistent commit or cancel actions in the footer."] },
  accessibility: { heading: "Accessibility", content: ["Keep a visible title, supply an accurate trigger name, and preserve the close button as an additional dismissal path."] },
  implementation: { heading: "Implementation", content: ["Import Drawer and Button from paper-ui; pass application content and controls without importing Base UI dialog parts."] },
  apiReference: { heading: "API reference", content: ["DrawerProps accepts trigger, title, description, children, footer, placement, closeLabel, and controlled or uncontrolled open state. DrawerPlacement is start or end."] },
  relatedPatterns: { heading: "Related patterns", content: ["Use Modal for compact decisions, Sidebar for persistent local navigation, and a page for durable workflows."] },
  migration: { heading: "Migration", content: ["Replace application-owned portal, backdrop, focus trap, and slide animation code while retaining the application trigger and body content."] },
  lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0 with start/end placement, focus containment, dismissal, restoration, and owner-document behavior covered."] },
};

export const publishedPrimitiveDocuments = [
  document({
    id: "component.tabs",
    route: "/components/navigation/tabs",
    name: "Tabs",
    category: "navigation",
    summary: "Switches among related content views with one keyboard-operable selection model.",
    keywords: ["tabs", "tabpanel", "content", "selection", "keyboard"],
    sourcePath: "src/components/navigation/Tabs/Tabs.tsx",
    guidance: {
      whenToUse: tabsSections.whenToUse.content,
      whenNotToUse: tabsSections.whenNotToUse.content,
      content: tabsSections.contentGuidance.content,
      commonMistakes: tabsSections.choosingBetween.content,
    },
    accessibility: {
      requirements: tabsSections.accessibility.content,
      keyboard: tabsSections.behavior.content,
      knownConstraints: tabsSections.statesAndAdaptation.content,
    },
    api: {
      react: ["Tabs", "TabsRoot", "TabsList", "TabsTab", "TabsPanel"],
      cssClasses: ["paper-tabs", "paper-tabs__list", "paper-tabs__tab", "paper-tabs__panel"],
      publicTypes: ["TabsValue", "TabsOrientation", "TabsRootProps", "TabsListProps", "TabsTabProps", "TabsPanelProps"],
      defaults: ["orientation=horizontal", "activateOnFocus=true", "loopFocus=true", "keepMounted=true"],
      invalidCombinations: ["Do not use Tabs for navigation to another URL."],
    },
    fixtureIds: ["tabs.content"],
    migration: { legacy: ["application-owned tab roles"], notes: tabsSections.migration.content },
    behaviorTestIds: ["tabs.semantics", "tabs.keyboard", "tabs.controlled", "tabs.mounting"],
    sections: sections(tabsSections),
  }),
  document({
    id: "component.drawer",
    route: "/components/overlays/drawer",
    name: "Drawer",
    category: "overlays",
    summary: "Opens a modal side sheet for supporting work while preserving page context.",
    keywords: ["drawer", "sheet", "overlay", "focus", "filters"],
    sourcePath: "src/components/overlays/Drawer/Drawer.tsx",
    guidance: {
      whenToUse: drawerSections.whenToUse.content,
      whenNotToUse: drawerSections.whenNotToUse.content,
      content: drawerSections.contentGuidance.content,
      commonMistakes: drawerSections.choosingBetween.content,
    },
    accessibility: {
      requirements: drawerSections.accessibility.content,
      keyboard: drawerSections.behavior.content,
      knownConstraints: drawerSections.statesAndAdaptation.content,
    },
    api: {
      react: ["Drawer", "DRAWER_PLACEMENTS"],
      cssClasses: ["paper-drawer", "paper-drawer__body", "paper-drawer__footer"],
      publicTypes: ["DrawerProps", "DrawerPlacement"],
      defaults: ["placement=end", "closeLabel=Close", "modal=true"],
      invalidCombinations: ["Do not nest Drawer or Modal instances."],
    },
    fixtureIds: ["drawer.filters"],
    migration: { legacy: ["application-owned side sheet"], notes: drawerSections.migration.content },
    behaviorTestIds: ["drawer.semantics", "drawer.focus", "drawer.dismissal", "drawer.owner-document"],
    sections: sections(drawerSections),
  }),
] as const;
