import Example10 from "./examples/action-menu.recommended";
import Example10Source from "./examples/action-menu.recommended.tsx?raw";
import Example4 from "./examples/button.classes";
import Example4Source from "./examples/button.classes.tsx?raw";
import Example3 from "./examples/button.icons-disabled";
import Example3Source from "./examples/button.icons-disabled.tsx?raw";
import Example1 from "./examples/button.loading";
import Example1Source from "./examples/button.loading.tsx?raw";
import Example2 from "./examples/button.narrow";
import Example2Source from "./examples/button.narrow.tsx?raw";
import Example0 from "./examples/button.variants";
import Example0Source from "./examples/button.variants.tsx?raw";
import Example7 from "./examples/input.error";
import Example7Source from "./examples/input.error.tsx?raw";
import Example5 from "./examples/input.recommended";
import Example5Source from "./examples/input.recommended.tsx?raw";
import Example6 from "./examples/input.states";
import Example6Source from "./examples/input.states.tsx?raw";
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
  defineCatalogFixture({
  "id": "button.variants",
  "name": "Button variants",
  "description": "The complete action hierarchy beside a semantic anchor.",
  "tags": [
    "button",
    "variants",
    "anchor"
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
    code: Example0Source, render: () => <Example0 />
  }),
  defineCatalogFixture({
  "id": "button.loading",
  "name": "Loading action",
  "description": "A busy action keeps its visible and accessible name stable.",
  "tags": [
    "button",
    "loading",
    "disabled"
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
    code: Example1Source, render: () => <Example1 />
  }),
  defineCatalogFixture({
  "id": "button.narrow",
  "name": "Long, full-width action",
  "description": "Long localized content at a narrow viewport.",
  "tags": [
    "button",
    "full width",
    "long content",
    "narrow"
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
    code: Example2Source, render: () => <Example2 />
  }),
  defineCatalogFixture({
  "id": "button.icons-disabled",
  "name": "Icons and unavailable actions",
  "description": "A labelled action with a decorative icon, an explicitly named icon-only control, and an unavailable action with its reason.",
  "tags": [
    "button",
    "icons",
    "disabled"
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
    code: Example3Source, render: () => <Example3 />
  }),
  defineCatalogFixture({
  "id": "button.classes",
  "name": "Native elements with public classes",
  "description": "Use the public CSS recipe directly when a React wrapper is unnecessary. Native roles, disabled state and busy semantics stay explicit.",
  "tags": [
    "button",
    "classes",
    "native"
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
    code: Example4Source, render: () => <Example4 />
  }),
  defineCatalogFixture({
  "id": "input.recommended",
  "name": "Reading-log title",
  "description": "Label, hint, required rule, value, and submission in one field.",
  "tags": [
    "input",
    "form",
    "required",
    "hint"
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
    code: Example5Source, render: () => <Example5 />
  }),
  defineCatalogFixture({
  "id": "input.states",
  "name": "Read-only and disabled",
  "description": "Two non-editable states with distinct native semantics.",
  "tags": [
    "input",
    "readonly",
    "disabled"
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
    code: Example6Source, render: () => <Example6 />
  }),
  defineCatalogFixture({
  "id": "input.error",
  "name": "Validation error",
  "description": "Hint and error remain associated with the invalid input.",
  "tags": [
    "input",
    "error",
    "accessibility"
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
    code: Example7Source, render: () => <Example7 />
  }),
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
  defineCatalogDocument({
  "id": "component.button",
  "route": "/components/actions/button",
  "name": "Button",
  "category": "actions",
  "summary": "Triggers an immediate action with explicit hierarchy and safe defaults.",
  "keywords": [
    "button",
    "action",
    "loading",
    "anchor",
    "submit"
  ],
  "sourcePath": "src/components/actions/Button/Button.tsx",
  "fixtureIds": [
    "button.variants",
    "button.loading",
    "button.narrow",
    "button.icons-disabled",
    "button.classes"
  ],
  "behaviorTestIds": [
    "button.semantics",
    "button.recipe-parity",
    "button.loading"
  ],
  "guidance": {
    "whenToUse": [
      "Trigger an immediate operation such as saving or deleting."
    ],
    "whenNotToUse": [
      "Use an anchor for navigation to another resource."
    ],
    "content": [
      "Start with a specific verb and keep labels stable while loading."
    ],
    "commonMistakes": [
      "Do not use the link variant to turn navigation into a button."
    ]
  },
  "accessibility": {
    "requirements": [
      "Provide a stable accessible name and visible focus."
    ],
    "keyboard": [
      "Enter and Space activate a button; anchors retain native Enter behavior."
    ],
    "knownConstraints": [
      "Icon-only actions need an explicit aria-label."
    ]
  },
  "api": {
    "react": [
      "Button"
    ],
    "props": [
      {
        "name": "variant",
        "type": "\"default\" | \"outline\" | \"ghost\" | \"link\" | \"destructive\"",
        "description": "Choose emphasis by the action’s role. Link styling still renders a button.",
        "defaultValue": "default"
      },
      {
        "name": "type",
        "type": "\"button\" | \"submit\" | \"reset\"",
        "description": "Use submit explicitly for form submission; call methods.reset() for RHF resets.",
        "defaultValue": "button"
      },
      {
        "name": "loading",
        "type": "boolean",
        "description": "Disables repeat activation and announces progress without replacing the label.",
        "defaultValue": "false"
      },
      {
        "name": "loadingLabel",
        "type": "string",
        "description": "Accessible progress description; name the pending operation.",
        "defaultValue": "Working"
      },
      {
        "name": "disabled",
        "type": "boolean",
        "description": "Makes an unavailable action inert. Keep its reason visible nearby.",
        "defaultValue": "false"
      },
      {
        "name": "fullWidth",
        "type": "boolean",
        "description": "Fills the available inline width; long labels may wrap.",
        "defaultValue": "false"
      },
      {
        "name": "leadingIcon / trailingIcon",
        "type": "ReactNode",
        "description": "Decorative icons; provide aria-label on icon-only actions."
      },
      {
        "name": "children",
        "type": "ReactNode",
        "description": "A concise visible action label."
      },
      {
        "name": "onClick",
        "type": "MouseEventHandler<HTMLButtonElement>",
        "description": "Application-owned action handler. Native button attributes and refs are forwarded."
      }
    ],
    "cssClasses": [
      "paper-button",
      "buttonClassName()"
    ],
    "publicTypes": [
      "ButtonProps",
      "ButtonVariant",
      "ButtonRecipeOptions"
    ],
    "defaults": [
      "variant=default",
      "type=button",
      "loading=false"
    ],
    "invalidCombinations": [
      "Do not put href on Button; style a real anchor with buttonClassName()."
    ]
  },
  "migration": {
    "legacy": [
      "ui Button",
      "btn primary",
      "btn secondary",
      "btn danger"
    ],
    "notes": [
      "Map intent: primary to default, secondary to outline, danger to destructive."
    ]
  },
  "sections": {
    "required": {
      "overview": {
        "heading": "Overview",
        "content": [
          "Button expresses action hierarchy without changing native semantics."
        ]
      },
      "whenToUse": {
        "heading": "When to use",
        "content": [
          "Use Button for an immediate user-initiated operation such as saving a log, applying a filter, or deleting an entry. Use one default button for the primary action in a local task; give supporting actions less emphasis."
        ]
      },
      "whenNotToUse": {
        "heading": "When not to use",
        "content": [
          "Use an anchor when activation navigates to a URL, including links styled with buttonClassName(). Do not use a disabled button as an explanation; keep the reason visible near the unavailable action."
        ]
      },
      "choosingBetween": {
        "heading": "Choose between",
        "content": [
          "Default is the emphasized action, outline is a neutral alternative, ghost is a low-emphasis toolbar action, and link is an action embedded in prose. Destructive communicates irreversible intent; it does not replace a confirmation when the consequence is difficult to undo."
        ]
      },
      "anatomy": {
        "heading": "Anatomy",
        "content": [
          "A control contains an optional icon, a stable text label, and a lower interactive edge."
        ]
      },
      "recommendedExample": {
        "heading": "Recommended example",
        "content": [
          "Compare action emphasis, then select Loading, Icons and unavailable actions, or Native elements with public classes. The native example uses the same recipe without requiring the Button wrapper; supply native type, disabled and aria-busy attributes yourself."
        ]
      },
      "variants": {
        "heading": "Variants",
        "content": [
          "Use default for the main action in a task, outline for a visible alternative, ghost for compact secondary actions, link for button behavior that belongs inline with text, and destructive for a destructive operation. buttonClassName() gives a real anchor the same visual hierarchy without changing its navigation semantics."
        ]
      },
      "statesAndAdaptation": {
        "heading": "States and adaptation",
        "content": [
          "loading prevents repeat activation, sets aria-busy, keeps the original accessible name, and announces loadingLabel as a description. fullWidth is intended for constrained layouts; density changes sizing without changing hierarchy."
        ]
      },
      "behavior": {
        "heading": "Behavior",
        "content": [
          "Buttons activate on Enter or Space and default to type=button, preventing accidental form submission. Explicitly set type=submit inside a React Hook Form and submit through methods.handleSubmit(); use methods.reset() only for a deliberate reset workflow."
        ]
      },
      "contentGuidance": {
        "heading": "Content guidance",
        "content": [
          "Start with a specific verb and name the object when context is not obvious: Save log, Delete entry, or Clear filters. Keep the visible label stable while loading and avoid vague labels such as OK or Submit."
        ]
      },
      "accessibility": {
        "heading": "Accessibility",
        "content": [
          "Keep an accessible name, do not convey danger through color alone, and preserve native anchor roles for navigation."
        ]
      },
      "implementation": {
        "heading": "Implementation",
        "content": [
          "Button and buttonClassName() call the same recipe; load paper-ui/styles.css once at the application root."
        ]
      },
      "apiReference": {
        "heading": "API reference",
        "content": [
          "ButtonProps adds variant, loading, loadingLabel, icons, and fullWidth to native button attributes."
        ]
      },
      "relatedPatterns": {
        "heading": "Related patterns",
        "content": [
          "Use Modal for focused confirmation and ActionMenu when several contextual actions compete."
        ]
      },
      "migration": {
        "heading": "Migration",
        "content": [
          "Review old primary, secondary, and danger classes by semantic job instead of renaming mechanically."
        ]
      },
      "lifecycle": {
        "heading": "Lifecycle",
        "content": [
          "Stable in Paper 0.1.0; behavior and variant vocabulary are migration-ready."
        ]
      }
    },
    "pageSections": [
      "usage",
      "examples",
      "variantsAndStates",
      "behavior",
      "contentGuidance",
      "accessibility"
    ]
  },
  "kind": "component",
  "aliases": [],
  "lifecycle": "Stable",
  "reviewDate": "2026-09-05",
  "packageVersion": "0.1.0",
  "dependencies": {
    "documents": [],
    "packages": [
      "paper-ui"
    ]
  },
  "changelog": [
    {
      "date": "2026-09-05",
      "note": "Published the Stable Phase 2 contract."
    }
  ]
}),
  defineCatalogDocument({
  "id": "component.input",
  "route": "/components/forms/input",
  "name": "Input",
  "category": "forms",
  "summary": "Collects one line of text with complete field anatomy and form state.",
  "keywords": [
    "input",
    "field",
    "form",
    "error",
    "hint"
  ],
  "sourcePath": "src/components/forms/Input/Input.tsx",
  "fixtureIds": [
    "input.recommended",
    "input.states",
    "input.error"
  ],
  "behaviorTestIds": [
    "input.associations",
    "input.validation",
    "input.native-states"
  ],
  "guidance": {
    "whenToUse": [
      "Collect a short, free-form value in a React Hook Form."
    ],
    "whenNotToUse": [
      "Use a choice control when the set of valid values is known."
    ],
    "content": [
      "Use a visible noun label and a hint only when it adds constraints or context.",
      "Fields fill their parent column. Set a readable width on the form, such as w-full max-w-2xl, then group related fields with grid gap-4 sm:grid-cols-2. A full-width field and a two-field row share the same outside edges."
    ],
    "commonMistakes": [
      "Do not use placeholder text as the only label."
    ]
  },
  "accessibility": {
    "requirements": [
      "Associate label, hint, and error IDs with the native input."
    ],
    "keyboard": [
      "The native input follows platform text-editing behavior."
    ],
    "knownConstraints": [
      "Input requires a react-hook-form FormProvider."
    ]
  },
  "api": {
    "react": [
      "Input"
    ],
    "props": [
      {
        "name": "name",
        "type": "string",
        "description": "React Hook Form field path.",
        "required": true
      },
      {
        "name": "label",
        "type": "string",
        "description": "Persistent visible label.",
        "required": true
      },
      {
        "name": "hint",
        "type": "string",
        "description": "Format, constraint or consequence; associated with the input."
      },
      {
        "name": "rules",
        "type": "RegisterOptions",
        "description": "Validation messages and constraints supplied to React Hook Form."
      },
      {
        "name": "required",
        "type": "boolean",
        "description": "Sets the native required attribute, label marker and a default RHF required rule.",
        "defaultValue": "false"
      },
      {
        "name": "type / inputMode",
        "type": "native input attributes",
        "description": "Use type for input semantics and inputMode for the platform keyboard.",
        "defaultValue": "text"
      },
      {
        "name": "readOnly",
        "type": "boolean",
        "description": "Allows focus and copying while preventing edits.",
        "defaultValue": "false"
      },
      {
        "name": "disabled",
        "type": "boolean",
        "description": "Makes the field unavailable.",
        "defaultValue": "false"
      },
      {
        "name": "id / ref",
        "type": "string / Ref<HTMLInputElement>",
        "description": "Optional explicit ID or imperative focus target."
      }
    ],
    "cssClasses": [
      "paper-field",
      "paper-input"
    ],
    "publicTypes": [
      "InputProps"
    ],
    "defaults": [
      "Native text input behavior",
      "Errors come from react-hook-form state"
    ],
    "invalidCombinations": [
      "Do not render Input outside FormProvider."
    ]
  },
  "migration": {
    "legacy": [
      "ui/components/Form Input"
    ],
    "notes": [
      "Keep names and validation rules in react-hook-form; replace implicit global field styling."
    ]
  },
  "sections": {
    "required": {
      "overview": {
        "heading": "Overview",
        "content": [
          "Input owns the complete visible and semantic field relationship."
        ]
      },
      "whenToUse": {
        "heading": "When to use",
        "content": [
          "Use Input for a short free-form value that fits on one line, such as a reading-log title, page count, email address, or password. Set the appropriate native type or inputMode so platform keyboards and validation can help."
        ]
      },
      "whenNotToUse": {
        "heading": "When not to use",
        "content": [
          "Use TextArea for prose and Select, radio controls, or Autocomplete when valid choices are known. Do not use placeholder text as the label: it disappears after entry and does not provide a persistent field name."
        ]
      },
      "choosingBetween": {
        "heading": "Choose between",
        "content": [
          "A read-only value remains focusable and selectable and is still submitted; a disabled value is unavailable and omitted from submission. Use readOnly when users may need to inspect or copy the value, and disabled only when the control does not currently participate in the form."
        ]
      },
      "anatomy": {
        "heading": "Anatomy",
        "content": [
          "The field contains a persistent label, native input, optional hint, and validation message. Hints sit below controls so fields with and without hints align in a row."
        ]
      },
      "recommendedExample": {
        "heading": "Recommended example",
        "content": [
          "Edit the full-width log title and the paired Pages read and Date fields, then save to see their submitted values. The row shares the title field’s outside edges, and both controls align even though only Date has a hint. Clear the title and save again to test error and focus behavior."
        ]
      },
      "variants": {
        "heading": "Variants",
        "content": [
          "Text, email, password, numeric-input-mode, and other native types share the field anatomy."
        ]
      },
      "statesAndAdaptation": {
        "heading": "States and adaptation",
        "content": [
          "required adds the native required state and a visible marker. An invalid field sets aria-invalid and connects its alert message alongside any hint. Read-only remains operable for selection; disabled is visually muted and removed from interaction and submission."
        ]
      },
      "behavior": {
        "heading": "Behavior",
        "content": [
          "Registration, value, blur, and validation state come from the nearest FormProvider."
        ]
      },
      "contentGuidance": {
        "heading": "Content guidance",
        "content": [
          "Use a persistent noun phrase for the label. Add a hint only for a format, constraint, or consequence the label cannot carry. An error should explain how to fix the value—Enter the number of pages read is more useful than Invalid input.",
          "Fields fill their parent column. Set a readable width on the form, such as w-full max-w-2xl, then group related fields with grid gap-4 sm:grid-cols-2. A full-width field and a two-field row share the same outside edges."
        ]
      },
      "accessibility": {
        "heading": "Accessibility",
        "content": [
          "htmlFor, aria-describedby, aria-invalid, and role=alert connect the field anatomy without relying on color."
        ]
      },
      "implementation": {
        "heading": "Implementation",
        "content": [
          "Create methods with useForm(), wrap fields in FormProvider, and submit through methods.handleSubmit()."
        ]
      },
      "apiReference": {
        "heading": "API reference",
        "content": [
          "InputProps requires name and label, accepts hint and register rules, and otherwise follows native input attributes."
        ]
      },
      "relatedPatterns": {
        "heading": "Related patterns",
        "content": [
          "Button submits the surrounding form; future choice controls reuse the same field anatomy."
        ]
      },
      "migration": {
        "heading": "Migration",
        "content": [
          "Move validation into register rules and remove application-owned label/error selectors."
        ]
      },
      "lifecycle": {
        "heading": "Lifecycle",
        "content": [
          "Stable in Paper 0.1.0 with deterministic associations and React Hook Form ownership."
        ]
      }
    },
    "pageSections": [
      "usage",
      "examples",
      "variantsAndStates",
      "behavior",
      "contentGuidance",
      "accessibility"
    ]
  },
  "kind": "component",
  "aliases": [],
  "lifecycle": "Stable",
  "reviewDate": "2026-09-05",
  "packageVersion": "0.1.0",
  "dependencies": {
    "documents": [],
    "packages": [
      "paper-ui"
    ]
  },
  "changelog": [
    {
      "date": "2026-09-05",
      "note": "Published the Stable Phase 2 contract."
    }
  ]
}),
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
