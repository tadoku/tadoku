import Example10 from "./examples/action-menu.recommended";
import Example10Source from "./examples/action-menu.recommended.tsx?raw";
import Example9 from "./examples/modal.composable-search";
import Example9Source from "./examples/modal.composable-search.tsx?raw";
import Example8 from "./examples/modal.recommended";
import Example8Source from "./examples/modal.recommended.tsx?raw";
import {
COMPONENT_PAGE_SECTION_KEYS,
defineCatalogDocument,
defineCatalogFixture,
type CatalogDocument,
type ComponentCategory,
type ComponentDocumentationSections,
type RequiredComponentSectionKey,
} from "./schema";
import { buttonDocument, buttonFixtures } from "./button";
import { inputDocument, inputFixtures } from "./input";

const REVIEW_DATE = "2026-09-05";
const PACKAGE_VERSION = "0.1.0";

function section(
  heading: string,
  ...content: readonly string[]
): { readonly heading: string; readonly content: readonly string[] } {
  return { heading, content };
}

function componentSections(
  name: string,
  details: Readonly<
    Record<
      RequiredComponentSectionKey,
      { readonly heading: string; readonly content: readonly string[] }
    >
  >,
): ComponentDocumentationSections {
  void name;
  return { required: details, pageSections: COMPONENT_PAGE_SECTION_KEYS };
}

interface ComponentDocumentOptions {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly category: ComponentCategory;
  readonly summary: string;
  readonly keywords: readonly string[];
  readonly sourcePath: string;
  readonly fixtureIds: readonly string[];
  readonly behaviorTestIds: readonly string[];
  readonly guidance: CatalogDocument["guidance"];
  readonly accessibility: CatalogDocument["accessibility"];
  readonly api: CatalogDocument["api"];
  readonly migration: CatalogDocument["migration"];
  readonly sections: ComponentDocumentationSections;
}

function componentDocument(options: ComponentDocumentOptions): CatalogDocument {
  return defineCatalogDocument({
    ...options,
    kind: "component",
    aliases: [],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    packageVersion: PACKAGE_VERSION,
    dependencies: { documents: [], packages: ["paper-ui"] },
    changelog: [
      { date: REVIEW_DATE, note: "Published the Stable Phase 2 contract." },
    ],
  });
}

export const phaseTwoFixtures = [
  ...buttonFixtures,
  ...inputFixtures,
  defineCatalogFixture({
  "id": "modal.recommended",
  "name": "Confirm log deletion",
  "description": "Confirm a local demo deletion, or keep the log with either dismissal path.",
  "tags": [
    "modal",
    "focus",
    "destructive"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true,
    code: Example8Source, render: () => <Example8 />
  }),
  defineCatalogFixture({
  "id": "modal.composable-search",
  "name": "Composable catalogue search",
  "description": "A controlled footerless dialog focuses a real form field and filters local examples.",
  "tags": [
    "modal",
    "controlled",
    "trigger",
    "initial focus",
    "footerless"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true,
    code: Example9Source, render: () => <Example9 />
  }),
  defineCatalogFixture({
  "id": "action-menu.recommended",
  "name": "Reading-log actions",
  "description": "Common, unavailable, and destructive actions in one menu.",
  "tags": [
    "action menu",
    "keyboard",
    "disabled",
    "destructive"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true,
    code: Example10Source, render: () => <Example10 />
  })
] as const;

export const phaseTwoDocuments = [
  buttonDocument,
  inputDocument,
  componentDocument({
    id: "component.modal",
    route: "/components/overlays/modal",
    name: "Modal",
    category: "overlays",
    summary: "Temporarily focuses attention while managing focus, dismissal, and return.",
    keywords: ["modal", "dialog", "overlay", "focus", "confirmation"],
    sourcePath: "src/components/overlays/Modal/Modal.tsx",
    fixtureIds: ["modal.recommended", "modal.composable-search"],
    behaviorTestIds: ["modal.keyboard", "modal.focus-containment", "modal.focus-return"],
    guidance: {
      whenToUse: ["Require a focused decision or a short task without losing page context."],
      whenNotToUse: ["Use a page when the task is long, linkable, or needs deep navigation."],
      content: ["Name the decision in the title and make the consequence concrete."],
      commonMistakes: ["Do not nest modals or hide the only escape action."],
    },
    accessibility: {
      requirements: ["Expose a labelled dialog, trap focus while open, and return focus on close."],
      keyboard: ["Escape closes; Tab remains within the dialog."],
      knownConstraints: ["Long workflows should move to a dedicated page."],
    },
    api: {
      react: ["Modal"],
      props: [
  {
    "name": "title",
    "type": "string",
    "description": "Visible title and accessible dialog name.",
    "required": true
  },
  {
    "name": "children",
    "type": "ReactNode",
    "description": "A short bounded task or confirmation content.",
    "required": true
  },
  {
    "name": "description",
    "type": "string",
    "description": "Explain the consequence or task; associated with the dialog."
  },
  {
    "name": "triggerLabel",
    "type": "string",
    "description": "Name for the standard trigger; required unless supplying trigger."
  },
  {
    "name": "trigger",
    "type": "ReactElement",
    "description": "Application-owned trigger receiving dialog behavior and ARIA state."
  },
  {
    "name": "triggerVariant",
    "type": "ButtonVariant",
    "description": "Appearance of the standard trigger.",
    "defaultValue": "default"
  },
  {
    "name": "open / onOpenChange",
    "type": "boolean / (open: boolean) => void",
    "description": "Provide both for controlled state; close after successful async work."
  },
  {
    "name": "defaultOpen",
    "type": "boolean",
    "description": "Initial open state for an uncontrolled dialog.",
    "defaultValue": "false"
  },
  {
    "name": "initialFocus",
    "type": "boolean | RefObject<HTMLElement> | callback",
    "description": "Choose the initial focus target; use a field ref for editing tasks."
  },
  {
    "name": "closeLabel",
    "type": "string",
    "description": "Label for header and standard footer dismissal.",
    "defaultValue": "Close"
  },
  {
    "name": "action",
    "type": "{ label, variant?, disabled?, onAction }",
    "description": "Optional synchronous action in the standard footer. Activation also closes the dialog."
  },
  {
    "name": "footer",
    "type": "ReactNode",
    "description": "Replace the standard footer, or pass null to omit it. Use for async actions.",
    "defaultValue": "standard action/close footer"
  }
],
      cssClasses: [],
      publicTypes: ["ModalProps"],
      defaults: ["modal=true", "triggerVariant=default", "closeLabel=Close"],
      invalidCombinations: ["Do not nest Modal instances."],
    },
    migration: {
      legacy: ["ui Modal", "Headless UI Dialog"],
      notes: ["Paper owns the Base UI wrapper; applications do not import Base UI directly."],
    },
    sections: componentSections("Modal", {
      overview: section("Overview", "Modal creates a focused layer above the current task and restores context after dismissal."),
      whenToUse: section("When to use", "Use it for concise confirmations and bounded editing tasks."),
      whenNotToUse: section("When not to use", "Avoid it for passive notices, multi-step flows, or content that deserves a URL."),
      choosingBetween: section("Choose between", "Use Flash for non-blocking feedback, ActionMenu for contextual choices, and a page for sustained work."),
      anatomy: section("Anatomy", "Backdrop, viewport, titled popup, optional description, content, close affordance, and footer."),
      recommendedExample: section("Recommended example", "Review deletion and choose Delete log to see a local confirmation. Keep log, Escape, and the header close leave the value unchanged. Search examples demonstrates controlled state, initial focus, and a footerless search field."),
      variants: section("Variants", "Use the standard triggerLabel and action contract for common confirmations. Pass an application-owned trigger and footer when the task needs a composed entry point or action layout; footer=null intentionally omits the footer."),
      statesAndAdaptation: section("States and adaptation", "Use open with onOpenChange when application state or another control owns visibility. The viewport scrolls long content and remains bounded at phone, tablet, and desktop widths."),
      behavior: section("Behavior", "Base UI manages opening, focus containment, Escape/outside dismissal, and focus return. initialFocus targets an application-owned field. The standard action closes immediately; for an asynchronous save, use controlled open state and a custom footer, keep errors inside the dialog, and close only after success."),
      contentGuidance: section("Content guidance", "Use a question for confirmation titles and put consequences in the description."),
      accessibility: section("Accessibility", "Title and description label the dialog; two close paths ensure touch-screen-reader escape."),
      implementation: section("Implementation", "Import Modal from paper-ui; Base UI remains a private interaction dependency."),
      apiReference: section("API reference", "ModalProps supports controlled or uncontrolled open state, a standard or application-owned trigger, initialFocus, footer replacement or omission, title, description, content, close, and optional primary action contracts."),
      relatedPatterns: section("Related patterns", "Pair with destructive Button intent only when the action is genuinely destructive."),
      migration: section("Migration", "Replace legacy Dialog imports with Modal and remove application-owned portal/focus logic."),
      lifecycle: section("Lifecycle", "Stable in Paper 0.1.0 after keyboard, containment, dismissal, and return tests."),
    }),
  }),
  componentDocument({
    id: "component.action-menu",
    route: "/components/actions/action-menu",
    name: "ActionMenu",
    category: "actions",
    summary: "Presents a compact keyboard-operable set of contextual actions.",
    keywords: ["menu", "actions", "keyboard", "disabled", "destructive"],
    sourcePath: "src/components/overlays/ActionMenu/ActionMenu.tsx",
    fixtureIds: ["action-menu.recommended"],
    behaviorTestIds: ["action-menu.keyboard", "action-menu.disabled", "action-menu.selection"],
    guidance: {
      whenToUse: ["Group three or more contextual actions for one object."],
      whenNotToUse: ["Keep a single important action visible as Button."],
      content: ["Use parallel verb phrases ordered by frequency, with destructive actions last."],
      commonMistakes: ["Do not use a menu to hide primary page navigation."],
    },
    accessibility: {
      requirements: ["Expose a named trigger, menu semantics, roving focus, and disabled state."],
      keyboard: ["Enter, Space, or ArrowDown opens; arrows move; Escape closes and returns focus."],
      knownConstraints: ["Menu items are actions; router-neutral link integration is a separate contract."],
    },
    api: {
      react: ["ActionMenu"],
      props: [
  { "name": "iconOnly", "type": "boolean", "defaultValue": "false", "description": "Show a named icon instead of text and chevron, with a 44px target in both densities. Defaults to an ellipsis; use triggerIcon for a recognizable account or context icon. Give label the full action context." },
  { "name": "triggerIcon", "type": "ReactNode", "description": "Optional decorative replacement for the default trigger icon. Paper supplies its size and hides it from assistive technology; label supplies the accessible name." },
  {
    "name": "label",
    "type": "string",
    "description": "Visible trigger name identifying which object the actions affect.",
    "required": true
  },
  {
    "name": "items",
    "type": "readonly ActionMenuItem[]",
    "description": "Stable id, visible label and onSelect callback for each action.",
    "required": true
  },
  {
    "name": "items[].disabled",
    "type": "boolean",
    "description": "Keeps an action visible but unavailable; explain why outside the menu.",
    "defaultValue": "false"
  },
  {
    "name": "items[].destructive",
    "type": "boolean",
    "description": "Highlights destructive intent; does not add confirmation.",
    "defaultValue": "false"
  },
  {
    "name": "items[].icon",
    "type": "ReactNode",
    "description": "Optional decorative icon beside the label."
  },
  {
    "name": "triggerVariant",
    "type": "ButtonVariant",
    "description": "Choose an appropriate trigger hierarchy.",
    "defaultValue": "outline"
  },
  {
    "name": "defaultOpen",
    "type": "boolean",
    "description": "Initial open state; normally leave closed.",
    "defaultValue": "false"
  }
],
      cssClasses: [],
      publicTypes: ["ActionMenuProps", "ActionMenuItem"],
      defaults: ["triggerVariant=outline", "loopFocus=true"],
      invalidCombinations: ["Do not treat disabled items as explanatory text."],
    },
    migration: {
      legacy: ["ui ActionMenu", "Headless UI Menu"],
      notes: ["Map each item to a stable id and explicit selection callback."],
    },
    sections: componentSections("ActionMenu", {
      overview: section("Overview", "ActionMenu keeps secondary contextual operations available without crowding the page."),
      whenToUse: section("When to use", "Use it for several actions that operate on the same reading log or table row."),
      whenNotToUse: section("When not to use", "Do not hide the main task or global navigation in a contextual action menu."),
      choosingBetween: section("Choose between", "Use Button for one visible action, Modal for a focused decision, and ActionMenu for a compact action set."),
      anatomy: section("Anatomy", "A named trigger anchors a floating menu containing text-labelled action items."),
      recommendedExample: section("Recommended example", "Open Log actions with the keyboard or pointer, select Edit log, and read the outcome below. Duplication stays disabled with its reason visible. Delete demonstrates intent only; real data deletion should request confirmation."),
      variants: section("Variants", "The trigger may use an appropriate Button variant; destructive styling belongs to the item, not the whole menu."),
      statesAndAdaptation: section("States and adaptation", "Highlighted, disabled, destructive, open, and closed states retain non-color cues and density sizing."),
      behavior: section("Behavior", "Base UI owns trigger activation, placement, roving focus, typeahead, dismissal, selection, and focus return."),
      contentGuidance: section("Content guidance", "Use short parallel verb phrases, order by likely use, and place destructive actions last."),
      accessibility: section("Accessibility", "The trigger names the action set; disabled items remain identifiable but cannot be selected."),
      implementation: section("Implementation", "Pass immutable item metadata and callbacks to ActionMenu; do not expose Base UI primitives to applications."),
      apiReference: section("API reference", "ActionMenuItem requires id, label, and onSelect; icon, disabled, and destructive are optional."),
      relatedPatterns: section("Related patterns", "Use Modal after selection only when the chosen action needs focused confirmation."),
      migration: section("Migration", "Replace legacy render-prop menus and implicit class ordering with explicit item intent."),
      lifecycle: section("Lifecycle", "Stable in Paper 0.1.0 with keyboard, disabled-item, selection, and real-browser pointer evidence."),
    }),
  }),
] as const;
