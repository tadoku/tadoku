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
import { defineCatalogDocument, defineCatalogFixture } from "./schema";

export const buttonFixtures = [
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
] as const;

export const buttonDocument = defineCatalogDocument({
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
          "Buttons activate on Enter or Space and default to type=button, preventing accidental form submission. Explicitly set type=submit inside a TanStack Form and submit through methods.handleSubmit(); use methods.reset() only for a deliberate reset workflow."
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
});
