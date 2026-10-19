import Example19 from "./examples/radio-group.format";
import Example20 from "./examples/radio-group.format.empty";
import Example20Source from "./examples/radio-group.format.empty.tsx?raw";
import Example19Source from "./examples/radio-group.format.tsx?raw";
import Example29 from "./examples/radio-select.native";
import Example29Source from "./examples/radio-select.native.tsx?raw";
import Example17 from "./examples/radio-select.viewport";
import Example18 from "./examples/radio-select.viewport.empty";
import Example18Source from "./examples/radio-select.viewport.empty.tsx?raw";
import Example17Source from "./examples/radio-select.viewport.tsx?raw";
import ToggleModifiersExample from "./examples/toggle-select.modifiers";
import ToggleModifiersSource from "./examples/toggle-select.modifiers.tsx?raw";
import EmptyToggleModifiersExample from "./examples/toggle-select.modifiers.empty";
import EmptyToggleModifiersSource from "./examples/toggle-select.modifiers.empty.tsx?raw";
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

export const choiceGroupDetails: ComponentDetails = {
  "component.radio-select": {
    example: "The segmented example selects a preview size. Compare it with the native-radio example, then submit an empty group to see how required validation is attached to the decision.",
    anatomy: "A fieldset and legend contain same-name native radios; segmented changes the labels' presentation without changing selection semantics.",
    variants: "Default uses visible native radio indicators; segmented joins concise peer choices.", states: "One option can be selected. Use defaultValues to initialize a required decision; disabled may apply to the whole group or an option.", content: "Use short peer labels.", implementation: formSetup,
    props: [...fieldProps, validationProp, optionsProp, { name: "variant", type: '"default" | "segmented"', defaultValue: '"default"', description: "Default radios for ordinary forms; segmented for two to four short peer choices." }, disabledProp],
  },
  "component.toggle-select": {
    example: "Select a different reading modifier, or press the selected one again to clear it. Save selection shows the stored choice. The listening example starts with its one optional modifier off.",
    anatomy: "One labelled group joins equal-width choices inside a shared border. A short description sits below each label, and a check has reserved space so selection does not move the text.",
    variants: "One connected presentation for one to four concise options. Wide controls use equal-width segments with descriptions beneath the labels. Narrow controls stack connected rows with inline descriptions; a single option spans the control.",
    states: "Initialize the field with [] for no selection or [value] for one selected option. Disabled options cannot be pressed. reset restores the initial selection; validation errors describe the group and focus its first enabled choice.",
    content: "Keep labels parallel and short. Use description for a rate or another small fact that changes the decision. Explain that all options can be off when that is not obvious.",
    implementation: formSetup,
    props: [
      ...fieldProps.filter((prop) => prop.name !== "required"),
      { ...optionsProp, description: "Each option has value and label, with optional description and disabled. The form stores [] or [value]; choosing another option replaces the previous choice." },
      validationProp,
      disabledProp,
    ],
  },
  "component.radio-group": {
    example: "Choose Book or Audio and inspect the saved string value. Other is visibly unavailable. The empty example shows the required error for the whole decision.",
    anatomy: "A fieldset/legend contains native radios presented as cards, each with a label and supporting description.",
    variants: "One card presentation. Use RadioSelect for choices that do not need descriptions; adding cards to terse choices adds unnecessary visual weight.",
    states: "Required defaults to true. A disabled option remains visible for explanation but cannot be chosen. Cards stack on narrow screens; long descriptions wrap within their label.",
    content: "Use the label for the format name and the description for the difference that affects the decision. Do not repeat the label in the description.", implementation: formSetup,
    props: [...fieldProps, validationProp, { ...optionsProp, type: "readonly RadioGroupOption<string>[]", description: "Each option requires value, label, and description; disabled is optional." }],
  },
};

export const choiceGroupCommonMistakes: Record<string, string> = {
  "component.radio-select": "Do not use segmented radios for commands. A radio changes one stored value; a button runs an operation.",
  "component.toggle-select": "Do not use this for several simultaneous choices or a required decision that cannot be cleared. Use checkboxes or RadioSelect for those cases.",
  "component.radio-group": "Do not hide the distinction between options in a tooltip. Keep decision-making descriptions visible beside each radio.",
};

export const toggleSelectFixtures = [
  defineCatalogFixture({
    id: "toggle-select.modifiers",
    name: "Optional reading modifier",
    description: "Choose one modifier, clear it, save the selection, and restore the initial value.",
    tags: ["toggle", "selection", "modifiers"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: [
      { id: "phone", label: "Phone", width: 320, height: 720 },
      { id: "desktop", label: "Desktop", width: 1280, height: 800 },
    ],
    deterministic: true,
    code: ToggleModifiersSource,
    render: () => <ToggleModifiersExample />,
  }),
  defineCatalogFixture({
    id: "toggle-select.modifiers.empty",
    name: "Single optional modifier",
    description: "An empty selection is valid. Turn passive listening on or off before saving.",
    tags: ["toggle", "selection", "modifiers", "empty"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: [
      { id: "phone", label: "Phone", width: 320, height: 720 },
      { id: "desktop", label: "Desktop", width: 1280, height: 800 },
    ],
    deterministic: true,
    code: EmptyToggleModifiersSource,
    render: () => <EmptyToggleModifiersExample />,
  }),
] as const;

export const toggleSelectSpec: DocSpec = {
  id: "component.toggle-select",
  route: "/components/forms/toggle-select",
  name: "ToggleSelect",
  category: "forms",
  summary: "Choose one optional value, with a second press to clear it.",
  fixtureId: "toggle-select.modifiers",
  sourcePath: "src/components/forms/ToggleSelect.tsx",
  react: ["ToggleSelect"],
  types: ["ToggleSelectProps", "Option"],
  css: ["paper-toggle-select", "paper-toggle-select__options", "paper-toggle-select__option", "paper-toggle-select__text", "paper-toggle-select__description", "paper-toggle-select__check"],
  when: "Use for a small set of optional, mutually exclusive choices such as score modifiers.",
  avoid: "Use RadioSelect when one choice must always remain selected, or Checkbox for independent choices that can be combined.",
  choose: "ToggleSelect lets users return to no selection by pressing the active option again.",
  behavior: "Tab moves between enabled choices. Enter, Space, or a click toggles the focused choice without submitting the form. Selecting another choice clears the previous one; focus stays on the pressed option.",
  accessibility: "The visible legend names the group. Each button exposes its pressed state, and a check reinforces selection without relying on color. Disabled choices are skipped by Tab.",
  migration: "Replace separate form buttons and selection state with ToggleSelect inside form.AppForm. Initialize the named field with [] or [value].",
};

export const radioSelectFixtures = [
  defineCatalogFixture({ ...{
  "id": "radio-select.viewport",
  "name": "Preview size",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "radio",
    "select",
    "viewport"
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
}, code: Example17Source, render: () => <Example17 /> }),
  defineCatalogFixture({ ...{
  "id": "radio-select.viewport.empty",
  "name": "Preview size — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "radio",
    "select",
    "viewport",
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
}, code: Example18Source, render: () => <Example18 /> }),
] as const;

export const radioSelectNativeFixtures = [
  defineCatalogFixture({ ...{
  "id": "radio-select.native",
  "name": "Native radio presentation",
  "description": "Compare default radios with the segmented example, including a disabled option.",
  "tags": [
    "radio",
    "select",
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
  "deterministic": true
}, code: Example29Source, render: () => <Example29 /> }),
] as const;

export const radioSelectSpec: DocSpec = { id: "component.radio-select", route: "/components/forms/radio-select", name: "RadioSelect", category: "forms", summary: "Shows concise mutually exclusive options as native radios, with an optional connected segmented presentation.", fixtureId: "radio-select.viewport", sourcePath: "src/components/forms/choice-controls.tsx", react: ["RadioSelect"], types: ["RadioSelectProps", "RadioSelectVariant", "Option"], when: "Use for two to four short choices that benefit from remaining visible; use segmented for compact peer choices such as preview sizes.", avoid: "Use Select for longer sets, RadioGroup when each choice needs description copy, and Button when activation does not select a form value.", choose: "Default exposes familiar radio controls; segmented visually connects short peers without changing their native semantics.", variants: "Default presents inline native radios. Segmented presents the same inputs as a connected quiet-surface group with an accent edge and check for the selected value.", states: "Both variants cover selected, unselected, focus-visible, disabled, and invalid states in light, dark, comfortable, compact, narrow, and forced-color settings.", behavior: "Arrow keys move within the native same-name radio set and Space selects. Segmented remains a form control and never exposes button or aria-pressed semantics.", content: "Use two to four parallel, concise nouns such as Phone, Tablet, and Desktop; require and initialize one value when the decision cannot be empty.", accessibility: "A fieldset and legend name the group while each label names its native radio. The segmented check and accent edge reinforce selection without replacing checked state.", migration: 'Preserve submitted option values and replace button-like viewport selectors with variant="segmented" rather than recreating toggle-button semantics.', pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] };

export const radioGroupFixtures = [
  defineCatalogFixture({ ...{
  "id": "radio-group.format",
  "name": "Reading format",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "radio",
    "group",
    "format"
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
}, code: Example19Source, render: () => <Example19 /> }),
  defineCatalogFixture({ ...{
  "id": "radio-group.format.empty",
  "name": "Reading format — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "radio",
    "group",
    "format",
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
}, code: Example20Source, render: () => <Example20 /> }),
] as const;

export const radioGroupSpec: DocSpec = { id: "component.radio-group", route: "/components/forms/radio-group", name: "RadioGroup", category: "forms", summary: "Presents mutually exclusive choices that each require supporting description.", fixtureId: "radio-group.format", sourcePath: "src/components/forms/choice-controls.tsx", react: ["RadioGroup"], types: ["RadioGroupProps", "RadioGroupOption"], when: "Use when each mutually exclusive choice needs a label and explanatory description.", avoid: "Use RadioSelect for terse choices and Checkbox for independent choices.", choose: "Choose the smallest radio presentation that gives users enough context to decide.", behavior: "Native radio focus and arrow-key selection work without a custom composite widget.", accessibility: "Use a legend for the decision, preserve disabled semantics, and never rely on the selected border alone.", migration: "Remove Headless UI RadioGroup imports; Paper uses native radios and TanStack Form." };
