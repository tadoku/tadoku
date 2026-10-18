import Example21 from "./examples/amount.progress";
import Example22 from "./examples/amount.progress.empty";
import Example22Source from "./examples/amount.progress.empty.tsx?raw";
import Example21Source from "./examples/amount.progress.tsx?raw";
import Example15 from "./examples/checkbox.public-entry";
import Example16 from "./examples/checkbox.public-entry.empty";
import Example16Source from "./examples/checkbox.public-entry.empty.tsx?raw";
import Example15Source from "./examples/checkbox.public-entry.tsx?raw";
import Example13 from "./examples/select.language";
import Example14 from "./examples/select.language.empty";
import Example14Source from "./examples/select.language.empty.tsx?raw";
import Example13Source from "./examples/select.language.tsx?raw";
import Example11 from "./examples/textarea.reading-notes";
import Example12 from "./examples/textarea.reading-notes.empty";
import Example12Source from "./examples/textarea.reading-notes.empty.tsx?raw";
import Example11Source from "./examples/textarea.reading-notes.tsx?raw";
import { defineCatalogFixture } from "./schema";
import {
  disabledProp,
  fieldProps,
  formSetup,
  optionsProp,
  validationProp,
  type ComponentDetails,
  type DocSpec,
} from "./forms-feedback-spec";

export const formFieldDetails: ComponentDetails = {
  "component.textarea": {
    example: "Edit the reading note and save it. The empty example demonstrates required validation: submit blank, then enter a note and resubmit to clear the error.",
    anatomy: "A persistent label, resizable native textarea, optional hint, and inline error share one field relationship.",
    variants: "One multiline presentation. Set rows for the expected starting length; the user can resize vertically. Density changes padding, not the meaning of the field.",
    states: "The populated and empty examples share the same required rule. Invalid submission displays an associated recovery message and focuses the field. Native disabled and readOnly have different purposes: unavailable editing versus selectable text that cannot change.",
    content: "Label the value, such as Reading notes. Use the hint for visibility or spoiler policy. Put examples in a placeholder only when they remain useful after the hint.", implementation: formSetup,
    props: [...fieldProps, validationProp, { name: "rows", type: "number", description: "Initial visible text rows; use a larger value for long-form writing." }, { name: "readOnly", type: "boolean", description: "Keeps the text focusable and selectable without allowing edits." }, disabledProp],
  },
  "component.select": {
    example: "Choose a reading language. The second example starts with a placeholder so Save entry demonstrates the required-choice error before selection.",
    anatomy: "A label precedes a native select with a decorative chevron; hint and error text follow the control.",
    variants: "Use options for a flat list or groups for native optgroups. When groups is supplied it replaces options; pass options={[]} for a grouped-only select.",
    states: "An empty string selects the placeholder. Required rejects that value. Individual options and the entire control can be disabled; no custom menu is substituted on mobile.",
    content: "Use parallel labels and an instructional placeholder such as Choose a language. A placeholder is an empty option, not a meaningful saved choice.", implementation: formSetup,
    props: [...fieldProps, validationProp, optionsProp, { name: "groups", type: "readonly OptionGroup<string>[]", description: "Named groups containing label and options; takes precedence over options." }, { name: "placeholder", type: "string", description: "Adds an option with an empty-string value before real choices." }, disabledProp],
  },
  "component.checkbox": {
    example: "Toggle profile visibility, then save and inspect the boolean form value. The unchecked example remains valid because this is an optional preference.",
    anatomy: "A native checkbox and its clickable label form one choice, followed by optional hint and error text.",
    variants: "One binary presentation. Checked means the labelled proposition is enabled; this component does not expose a mixed/indeterminate group-selection state.",
    states: "Use true or false in defaultValues. Optional false is valid. required is appropriate for consent that must be checked, not for preferences. Native disabled prevents toggling.",
    content: "Write a proposition that reads naturally when enabled: Show this entry on my profile. Explain who can still access the entry in the hint.", implementation: formSetup,
    props: [...fieldProps, validationProp, disabledProp],
  },
  "component.amount-with-unit": {
    example: "Enter positive progress, then choose pages or minutes. Inspect current values to see progressValue as a number and progressUnit as a separate string. Submit the empty example to see required validation.",
    anatomy: "A labelled group joins one numeric input to one native unit picker. The shared hint and first error describe both controls.",
    variants: "One compound field. Use Input with a unit in its label when the unit is fixed; a disabled picker should not be used to fake a fixed suffix.",
    states: "required, min, and max validate the amount on submission. An empty optional amount becomes NaN through valueAsNumber; normalize it before sending JSON. Disabling the field disables both controls.",
    content: "Label the measurement, such as Progress. Use concise plural unit labels, and state any minimum or rounding policy in the hint.", implementation: formSetup,
    props: [{ name: "name", type: "string", required: true, description: "Prefix for two fields: nameValue (number) and nameUnit (string). Initialize both in defaultValues." }, ...fieldProps.filter((prop) => prop.name !== "name"), { name: "units", type: "readonly Option<string>[]", required: true, description: "Available units. Each has value, label, and optional disabled." }, { name: "unitsLabel", type: "string", description: "Accessible unit-picker label; defaults to Unit for plus the lowercased field label." }, { name: "min / max / step", type: "number | string", description: "Native number constraints. min/max also validate through React Hook Form; use step for the native stepping interval." }, disabledProp],
  },
};

export const formFieldCommonMistakes: Record<string, string> = {
  "component.textarea": "Do not set a fixed height with overflow hidden: users need to read and resize longer notes.",
  "component.select": "Do not store a placeholder as a real language value. Initialize the empty option with an empty string and validate required selections.",
  "component.checkbox": "Do not mark an optional preference required; required means the user must check it to continue.",
  "component.amount-with-unit": "Do not initialize a single progress object: this component registers progressValue and progressUnit as two sibling fields.",
};

export const textAreaFixtures = [
  defineCatalogFixture({ ...{
  "id": "textarea.reading-notes",
  "name": "Reading notes",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "textarea",
    "reading",
    "notes"
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
}, code: Example11Source, render: () => <Example11 /> }),
  defineCatalogFixture({ ...{
  "id": "textarea.reading-notes.empty",
  "name": "Reading notes — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "textarea",
    "reading",
    "notes",
    "empty"
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
}, code: Example12Source, render: () => <Example12 /> }),
] as const;

export const textAreaSpec: DocSpec = { id: "component.textarea", route: "/components/forms/textarea", name: "TextArea", category: "forms", summary: "Collects multiline text with the same deterministic field anatomy as Input.", fixtureId: "textarea.reading-notes", sourcePath: "src/components/forms/controls.tsx", react: ["TextArea"], types: ["TextAreaProps"], css: ["paper-textarea"], when: "Use for prose, notes, and other values that genuinely need multiple lines.", avoid: "Use Input for short single-line values and do not resize away the user's editing space.", choose: "Choose TextArea over Input by content length, not by visual preference.", behavior: "Native text editing, Tab focus, and React Hook Form validation remain intact.", accessibility: "Associate the persistent label, hint, and recoverable error with the native textarea.", migration: "Replace legacy TextArea while retaining the existing field name and register rules." };

export const selectFixtures = [
  defineCatalogFixture({ ...{
  "id": "select.language",
  "name": "Language select",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "select",
    "language"
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
}, code: Example13Source, render: () => <Example13 /> }),
  defineCatalogFixture({ ...{
  "id": "select.language.empty",
  "name": "Language select — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "select",
    "language",
    "empty"
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
}, code: Example14Source, render: () => <Example14 /> }),
] as const;

export const selectSpec: DocSpec = { id: "component.select", route: "/components/forms/select", name: "Select", category: "forms", summary: "Uses the native platform picker for a known set of concise choices.", fixtureId: "select.language", sourcePath: "src/components/forms/controls.tsx", react: ["Select"], types: ["SelectProps", "Option", "OptionGroup"], when: "Use when one value must be chosen from a stable, reasonably short option set.", avoid: "Use Autocomplete for long searchable collections and RadioSelect when every choice should remain visible.", choose: "Prefer native Select unless a documented requirement needs composite listbox behavior.", behavior: "Platform keyboard, touch, and form submission behavior is preserved.", accessibility: "Provide a visible label and meaningful option labels; disabled options remain programmatically unavailable.", migration: "Map legacy values and groups directly without copying global select selectors." };

export const checkboxFixtures = [
  defineCatalogFixture({ ...{
  "id": "checkbox.public-entry",
  "name": "Public entry",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "checkbox",
    "public",
    "entry"
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
}, code: Example15Source, render: () => <Example15 /> }),
  defineCatalogFixture({ ...{
  "id": "checkbox.public-entry.empty",
  "name": "Public entry — unchecked",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "checkbox",
    "public",
    "entry",
    "empty"
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
}, code: Example16Source, render: () => <Example16 /> }),
] as const;

export const checkboxSpec: DocSpec = { id: "component.checkbox", route: "/components/forms/checkbox", name: "Checkbox", category: "forms", summary: "Captures an independent boolean choice with native semantics.", fixtureId: "checkbox.public-entry", sourcePath: "src/components/forms/controls.tsx", react: ["Checkbox"], types: ["CheckboxProps"], when: "Use for an independent yes/no setting or acknowledgement.", avoid: "Use radio choices when exactly one option in a set must be selected.", choose: "A checkbox toggles one proposition; a radio group chooses among mutually exclusive propositions.", behavior: "Space toggles the focused native checkbox and React Hook Form owns its boolean value.", accessibility: "Put the consequence in the visible label or hint and retain the native checkbox input.", migration: "Replace legacy Checkbox without converting it into a custom switch role." };

export const amountWithUnitFixtures = [
  defineCatalogFixture({ ...{
  "id": "amount.progress",
  "name": "Progress with unit",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "amount",
    "progress"
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
}, code: Example21Source, render: () => <Example21 /> }),
  defineCatalogFixture({ ...{
  "id": "amount.progress.empty",
  "name": "Progress with unit — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "amount",
    "progress",
    "empty"
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
}, code: Example22Source, render: () => <Example22 /> }),
] as const;

export const amountWithUnitSpec: DocSpec = { id: "component.amount-with-unit", route: "/components/forms/amount-with-unit", name: "AmountWithUnit", category: "forms", summary: "Groups a numeric amount with the unit that gives it meaning.", fixtureId: "amount.progress", sourcePath: "src/components/forms/controls.tsx", react: ["AmountWithUnit"], types: ["AmountWithUnitProps", "Option"], when: "Use when a numeric value is invalid or ambiguous without one unit from a short set.", avoid: "Use Input alone when the unit is fixed and can be stated in the label.", choose: "Keep the unit fixed in content when users cannot change it; otherwise group amount and native Select.", behavior: "The amount and unit submit as nameValue and nameUnit for legacy-compatible migration.", accessibility: "Name the amount visibly and provide a generated accessible name for the unit picker.", migration: "Retain the legacy paired field names while replacing its implicit layout and error selectors." };
