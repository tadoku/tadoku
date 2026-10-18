import Example30 from "./examples/button-group.entry-actions";
import Example30Source from "./examples/button-group.entry-actions.tsx?raw";
import Example34 from "./examples/surface.summary";
import Example34Source from "./examples/surface.summary.tsx?raw";
import { defineCatalogFixture } from "./schema";
import type { ComponentDetails, DocSpec } from "./forms-feedback-spec";

export const buttonGroupAndSurfaceDetails: ComponentDetails = {
  "component.button-group": {
    example: "View log follows an in-preview anchor, Edit log updates the status, and Delete log is disabled with a visible reason. These actions stay independently focusable.",
    anatomy: "A labelled role=group wraps real links and buttons. The wrapper has no toolbar keyboard behavior.",
    variants: "align=start or end positions the group. Individual actions use default, outline, ghost, link, or destructive; keep one primary action per task.",
    states: "Actions wrap when width is limited. visible=false removes an action. Disabled buttons use native disabled; disabled links retain link semantics but block activation with aria-disabled.",
    content: "Name the operation, such as Edit log. The group label describes the shared subject; explain why disabled destructive operations are unavailable.",
    implementation: "Import ButtonGroup from paper-ui. Provide href for navigation and onSelect for an operation; when both exist, onSelect also runs before navigation. The example defines the complete actions list and updates local status.",
    props: [{ name: "actions", type: "readonly ButtonGroupAction[]", required: true, description: "Each action requires id and label. Optional href, onSelect, variant, disabled, and visible determine native link/button behavior." }, { name: "label", type: "string", defaultValue: '"Actions"', description: "Accessible group name." }, { name: "align", type: '"start" | "end"', defaultValue: '"start"', description: "Horizontal alignment; actions wrap on small screens." }],
  },
  "component.surface": {
    example: "Compare the same reading summary at flat, floating, and showcase elevation, then compare a flat surface with an accent rail. Content and spacing stay constant so each treatment is visible.",
    anatomy: "A semantic container provides a background, rule, padding, optional shadow, and optional inline-start accent rail. Content headings and internal spacing remain explicit.",
    variants: "Flat is the default content group. Floating adds the lighter offset shadow. Showcase uses the stronger shadow for an intentional demonstration. accent adds a rail independently of elevation; it does not imply success or selection.",
    states: "Surface has no selected, disabled, hover, or interactive state. Content wraps with the available width. Theme changes surface/rule/shadow tokens; density inherits from the Paper root.",
    content: "Give articles or named sections a heading that describes their content. Avoid repeated nested cards: use spacing and rules within a surface before adding another container.",
    implementation: "Import Surface for div, article, or section containers. For another existing semantic element, use surfaceClassName with elevation, accent, and className. Load paper-ui/styles.css once; the recipe adds no click or keyboard behavior.",
    props: [{ name: "as", type: '"div" | "article" | "section"', defaultValue: '"div"', description: "HTML element. Use article for self-contained content; label a section when it should be a named region." }, { name: "elevation", type: '"flat" | "floating" | "showcase"', defaultValue: '"flat"', description: "Rule-only, light offset shadow, or stronger offset shadow." }, { name: "accent", type: "boolean", defaultValue: "false", description: "Adds the shared inline-start accent rail inside the surface boundary." }, { name: "className", type: "string", description: "Additional composition classes; paper-stack can space the children." }, { name: "children", type: "ReactNode", description: "Caller-owned content and heading hierarchy." }],
  },
};

export const buttonGroupAndSurfaceCommonMistakes: Record<string, string> = {
  "component.button-group": "Do not add toolbar arrow-key expectations: each link or button is reached independently with Tab.",
  "component.surface": "Do not treat an accent rail or shadow as an interactive affordance. Surface has no click or selection behavior.",
};

export const buttonGroupFixtures = [
  defineCatalogFixture({ ...{
  "id": "button-group.entry-actions",
  "name": "Entry actions",
  "description": "Run the actions and inspect their visible result.",
  "tags": [
    "button",
    "group",
    "entry",
    "actions"
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
  "deterministic": true
}, code: Example30Source, render: () => <Example30 /> }),
] as const;

export const buttonGroupSpec: DocSpec = { id: "component.button-group", route: "/components/actions/button-group", name: "ButtonGroup", category: "actions", summary: "Keeps related visible actions together without changing link or button semantics.", fixtureId: "button-group.entry-actions", sourcePath: "src/components/feedback/feedback.tsx", react: ["ButtonGroup"], types: ["ButtonGroupProps", "ButtonGroupAction"], when: "Use for a small, stable set of related actions that should remain visible.", avoid: "Use ActionMenu when the set is contextual or too large for the available width.", choose: "ButtonGroup prioritizes visibility; ActionMenu prioritizes compact contextual access.", behavior: "Each entry remains a real anchor or button and the wrapper contributes only a labelled group.", accessibility: "Preserve native roles, express disabled links with aria-disabled plus blocked activation, and label the group.", migration: "Replace responsive Next Link coupling with router-neutral href or onSelect actions." };

export const surfaceFixtures = [
  defineCatalogFixture({ ...{
  "id": "surface.summary",
  "name": "Elevation and accent comparison",
  "description": "Compare the supported presentations with consistent reading-log content.",
  "tags": [
    "surface",
    "summary"
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
  "deterministic": true
}, code: Example34Source, render: () => <Example34 /> })
] as const;

export const surfaceSpec: DocSpec = { id: "component.surface", route: "/components/data-display/surface", name: "Surface", category: "data-display", summary: "Composes flat, floating, or showcase content with optional Paper accent rail.", fixtureId: "surface.summary", sourcePath: "src/components/feedback/feedback.tsx", react: ["Surface", "surfaceClassName"], types: ["SurfaceProps"], css: ["paper-surface-card", "paper-elevation-*"], when: "Use to group related content when a semantic article, section, or neutral container also needs Paper surface treatment.", avoid: "Do not wrap every block in a card or use elevation as decoration without hierarchy.", choose: "Flat separates with rules; floating marks transient or interactive content; showcase is reserved for strong demonstrations.", behavior: "Surface adds no interaction and preserves the semantics selected with the as prop.", accessibility: "Choose article or section only when the content merits that landmark; elevation cannot replace headings.", migration: "Replace copied card and shadow classes with the shared recipe or Surface." };
