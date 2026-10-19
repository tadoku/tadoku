import Example7 from "./examples/input.error";
import Example7Source from "./examples/input.error.tsx?raw";
import Example5 from "./examples/input.recommended";
import Example5Source from "./examples/input.recommended.tsx?raw";
import Example6 from "./examples/input.states";
import Example6Source from "./examples/input.states.tsx?raw";
import { defineCatalogDocument, defineCatalogFixture } from "./schema";

export const inputFixtures = [
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
] as const;

export const inputDocument = defineCatalogDocument({
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
      "Collect a short, free-form value in a TanStack Form."
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
      "Input requires a @tanstack/react-form form.AppForm."
    ]
  },
  "api": {
    "react": [
      "Input"
    ],
    "props": [
      {
        "name": "name (AppField)",
        "type": "string",
        "description": "TanStack Form field path.",
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
        "name": "validators (AppField)",
        "type": "FieldValidators",
        "description": "Validation messages and constraints supplied to TanStack Form."
      },
      {
        "name": "required",
        "type": "boolean",
        "description": "Sets the native required attribute and label marker; add the corresponding required validator.",
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
      "Errors come from @tanstack/react-form state"
    ],
    "invalidCombinations": [
      "Do not render Input outside form.AppForm."
    ]
  },
  "migration": {
    "legacy": [
      "ui/components/Form Input"
    ],
    "notes": [
      "Keep names and validation rules in @tanstack/react-form; replace implicit global field styling."
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
          "Value, blur and validation state come from the enclosing form.AppField."
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
          "Create form with usePaperForm(), wrap form.Form in form.AppForm, and bind each named form.AppField to field.Input."
        ]
      },
      "apiReference": {
        "heading": "API reference",
        "content": [
          "InputProps requires label and accepts hint and native input attributes. Put name and validators on form.AppField."
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
          "Move validation into AppField validators and remove application-owned label/error selectors."
        ]
      },
      "lifecycle": {
        "heading": "Lifecycle",
        "content": [
          "Stable in Paper 0.1.0 with deterministic associations and TanStack Form ownership."
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
