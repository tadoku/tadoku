import { type ReactNode } from "react";
import Example48 from "./examples/drawer.filters";
import Example48Source from "./examples/drawer.filters.tsx?raw";
import Example47 from "./examples/tabs.content";
import Example47Source from "./examples/tabs.content.tsx?raw";
import {
COMPONENT_PAGE_SECTION_KEYS,
defineCatalogDocument,
defineCatalogFixture,
type CatalogDocument,
type ComponentDocumentationSections,
type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-09-05";
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
  fixture({ id: "tabs.content", name: "Reading log views", description: "Compare automatic horizontal selection with a controlled vertical list that waits for Enter or Space.", tags: ["tabs", "keyboard", "controlled"], code: Example47Source, render: () => <Example47 /> }),
  fixture({ id: "drawer.filters", name: "Reading-entry filters", description: "Apply a real language filter, cancel without saving, and compare end/start placement and optional footer.", tags: ["drawer", "keyboard", "controlled"], code: Example48Source, render: () => <Example48 /> }),
] as const;

const tabsSections: RequiredComponentSections = {
  overview: { heading: "Overview", content: ["Tabs switches among peer content views while keeping one view active."] },
  whenToUse: { heading: "When to use", content: ["Use Tabs to switch between peer views of the same local context, such as the summary and entries for one reading log."] },
  whenNotToUse: { heading: "When not to use", content: ["Use Tabbar for linked destinations with their own URLs, and use ordinary headings when readers should see every section together."] },
  choosingBetween: { heading: "Choose between", content: ["Tabs changes content in place. Tabbar and VerticalTabbar navigate to linked destinations and preserve native link behavior."] },
  anatomy: { heading: "Anatomy", content: ["A Tabs root contains one labelled list, two or more tabs, and one panel for each value."] },
  recommendedExample: { heading: "Recommended example", content: ["Compare automatic horizontal selection with controlled vertical selection. The horizontal example skips disabled Moderation. The vertical example uses activateOnFocus={false}: arrow keys move focus without changing the panel until Enter or Space."] },
  variants: { heading: "Variants", content: ["Use horizontal orientation by default. The second example uses vertical orientation and controlled value/onValueChange. Vertical orientation does not collapse automatically; choose horizontal when the available width cannot fit a side list and readable panel."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["Selected, focused, hovered, and disabled tabs retain distinct non-color cues. Horizontal lists scroll at narrow widths."] },
  behavior: { heading: "Behavior", content: ["Horizontal lists use Left/Right; vertical lists use Up/Down. Home and End reach the first/last enabled tab; loopFocus defaults to true. activateOnFocus defaults to true; set false when switching would cause expensive work. Inactive panels stay mounted by default, retaining local state. Set keepMounted={false} on a panel only when discarding that state is intentional."] },
  contentGuidance: { heading: "Content guidance", content: ["Use short parallel nouns such as Summary and Entries. The tab-list label should name the set, not repeat the word tabs."] },
  accessibility: { heading: "Accessibility", content: ["Give Tabs.List an accessible label, keep a panel for every tab value, and do not place linked page navigation in a tab role."] },
  implementation: { heading: "Implementation", content: ["Import Tabs and paper-ui/styles.css. Pair every Tab value with a Panel value, and supply defaultValue for an uncontrolled root or value plus onValueChange for controlled state. Do not mix these state models. Give List an aria-label or aria-labelledby."] },
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
  recommendedExample: { heading: "Recommended example", content: ["Open Review filters, choose a language, then Apply filters: the sheet closes and the result summary updates. Cancel, Escape and outside dismissal discard the draft on next open. Reading help demonstrates start placement and a footerless sheet."] },
  variants: { heading: "Variants", content: ["Placement is start or end in logical reading direction. Prefer end for supporting tasks unless navigation hierarchy calls for start."] },
  statesAndAdaptation: { heading: "States and adaptation", content: ["End placement opens from the reading-direction end; start opens from the beginning. The sheet is at most 28rem wide and leaves one control-width of backdrop visible on narrow screens. Its body scrolls independently, while header and footer stay visible. Reduced-motion removes transitions."] },
  behavior: { heading: "Behavior", content: ["Opening contains focus and blocks background interaction. Escape, outside press and Close dismiss and restore trigger focus. Use defaultOpen for an uncontrolled sheet or open/onOpenChange for a controlled one. Drawer does not save data or confirm unsaved changes; decide the draft policy in the application, as the example does."] },
  contentGuidance: { heading: "Content guidance", content: ["Use a brief noun title, add a description only when it clarifies scope, and keep persistent commit or cancel actions in the footer."] },
  accessibility: { heading: "Accessibility", content: ["Keep a visible title, supply an accurate trigger name, and preserve the close button. Tab and Shift+Tab stay within the sheet; Escape returns to the opener. Use the Phone preview to verify close and footer actions remain reachable while the body scrolls."] },
  implementation: { heading: "Implementation", content: ["Import Drawer, Button and the form controls from paper-ui; load paper-ui/styles.css once. Wrap the form in FormProvider. The footer lives outside the body form, so give the form a unique id from useId and the submit button a matching form attribute. Update controlled open state after a successful submit; retain the draft and display an error if saving fails."] },
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
      props: [{"name": "Root.defaultValue", "type": "string", "description": "Initial active Tab value for an uncontrolled root. Choose an enabled tab."}, {"name": "Root.value / onValueChange", "type": "string / (value: string) => void", "description": "Controlled selection. Update value from the callback; use instead of defaultValue."}, {"name": "Root.orientation", "type": "\"horizontal\" | \"vertical\"", "defaultValue": "horizontal", "description": "List layout and arrow-key axis. No automatic breakpoint switching."}, {"name": "List.aria-label / aria-labelledby", "type": "string", "required": true, "description": "Name the content-view group with one of these attributes."}, {"name": "List.activateOnFocus", "type": "boolean", "defaultValue": "true", "description": "Select as keyboard focus moves. Set false to require Enter/Space before changing panels."}, {"name": "List.loopFocus", "type": "boolean", "defaultValue": "true", "description": "Wrap focus at the first and last enabled tab."}, {"name": "Tab.value / Panel.value", "type": "string", "required": true, "description": "Matching unique value associates a tab and its panel."}, {"name": "Tab.disabled", "type": "boolean", "defaultValue": "false", "description": "Prevent selection and skip this tab during keyboard navigation. Explain unavailability nearby."}, {"name": "Panel.keepMounted", "type": "boolean", "defaultValue": "true", "description": "Keep hidden panel state in the DOM. False unmounts inactive panels."}, {"name": "className / ref / native attributes", "type": "per-part HTML props", "description": "Root/List/Panel forward div props and refs; Tab forwards button props and ref."}],
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
      props: [{"name": "trigger", "type": "ReactElement", "required": true, "description": "A focusable button that accepts merged props and a ref; Paper Button works directly."}, {"name": "title", "type": "ReactNode", "required": true, "description": "Visible heading that names the modal dialog."}, {"name": "children", "type": "ReactNode", "required": true, "description": "Scrollable body. The application owns content and form state."}, {"name": "description", "type": "ReactNode", "description": "Optional supporting copy associated with the dialog."}, {"name": "footer", "type": "ReactNode", "description": "Optional persistent actions outside the body. Link submit buttons to a form by id."}, {"name": "placement", "type": "\"start\" | \"end\"", "defaultValue": "end", "description": "Logical side of the viewport, following reading direction."}, {"name": "open / onOpenChange", "type": "boolean / (open: boolean) => void", "description": "Controlled open state. Apply the callback so Escape, outside press and close work."}, {"name": "defaultOpen", "type": "boolean", "defaultValue": "false", "description": "Initial open state for an uncontrolled drawer; do not combine with open."}, {"name": "closeLabel", "type": "string", "defaultValue": "Close", "description": "Accessible label for the built-in close button. Localize when needed."}],
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
