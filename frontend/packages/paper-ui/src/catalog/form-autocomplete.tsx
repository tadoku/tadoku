import Example23 from "./examples/autocomplete.language";
import Example24 from "./examples/autocomplete.language.empty";
import Example24Source from "./examples/autocomplete.language.empty.tsx?raw";
import Example23Source from "./examples/autocomplete.language.tsx?raw";
import Example25 from "./examples/multi-autocomplete.languages";
import Example26 from "./examples/multi-autocomplete.languages.empty";
import Example26Source from "./examples/multi-autocomplete.languages.empty.tsx?raw";
import Example25Source from "./examples/multi-autocomplete.languages.tsx?raw";
import Example27 from "./examples/tags.entry";
import Example28 from "./examples/tags.entry.empty";
import Example28Source from "./examples/tags.entry.empty.tsx?raw";
import Example27Source from "./examples/tags.entry.tsx?raw";
import { defineCatalogFixture } from "./schema";
import {
  disabledProp,
  fieldProps,
  formSetup,
  validationProp,
  type ComponentDetails,
  type DocSpec,
  type Prop,
} from "./forms-feedback-spec";

const autocompleteProps: readonly Prop[] = [...fieldProps, validationProp,
  { name: "options", type: "readonly Value[]", required: true, description: "Known objects supplied by the application. Fetch remote suggestions outside this component." },
  { name: "format", type: "(option: Value) => string", required: true, description: "Visible option label, also used for case-insensitive substring matching." },
  { name: "getId", type: "(option: Value) => string", required: true, description: "Stable unique identifier used to compare refreshed option objects with selected values." },
  { name: "match", type: "(option: Value, query: string) => boolean", description: "Optional custom matching, for example a label or language-code search." },
  { name: "maxResults", type: "number", defaultValue: "50", description: "Maximum filtered results rendered in the popup; this is not a selection limit." },
  { name: "placeholder", type: "string", description: "A short search prompt shown when the query is empty." }, disabledProp,
];
const selectionLimit: Prop = { name: "maxSelections", type: "number", description: "Maximum selected values. At the limit, remove a chip before adding another value." };

export const autocompleteDetails: ComponentDetails = {
  "component.autocomplete": {
    example: "Search Japanese, Chinese, or Korean, select a result, then save. The empty example requires a known option: typed text alone is not a submitted selection.",
    anatomy: "A labelled combobox has an input, open-list trigger, active option, and empty-result message. The popup remains in the input's document.",
    variants: "One object selection, initialized with an object or null. This small dataset makes keyboard behavior easy to inspect; use a native Select for three actual production choices.",
    states: "Search an unmatched word to see No matching options. required validates the selected object; rules accepts custom validation and messages. Blur reaches React Hook Form validation modes. disabled prevents both input and trigger interaction.",
    content: "Labels should explain the selected value; placeholders explain searching. Stable IDs must be unique even when visible labels repeat.", implementation: formSetup,
    props: autocompleteProps,
  },
  "component.multi-autocomplete": {
    example: "Choose two languages for a bilingual entry, then save. The consumer-supplied rules.validate requires two; maxSelections prevents a third. Remove a named chip to see the recovery message on submission.",
    anatomy: "Selected-object chips and a combobox input share a border; each chip exposes a named removal button.",
    variants: "Initialize an array of objects, including [] for empty. maxSelections limits the chosen values; maxResults separately limits visible search results.",
    states: "At the selection limit the input and trigger stop accepting additions while removal stays available. Removing one chip re-enables search. Escape dismisses the popup without clearing selected chips. The component does not create unknown objects or fetch remote data.",
    content: "State a selection limit in the hint before the user reaches it. Use rules.validate for eligibility or a minimum selection count; validation messages stay associated with the field. Use labels that still make sense in Remove Japanese removal actions.", implementation: formSetup,
    props: [...autocompleteProps, selectionLimit],
  },
  "component.tags-input": {
    example: "Choose a reading tag, type book club and press Enter to create it, then remove a chip. The input keeps focus so another tag can be added immediately.",
    anatomy: "A text input and removable chips share one field. The popup offers unselected suggestions and an Add action for a new tag.",
    variants: "Accepts string[] instead of objects, so format and getId are not needed. Use MultiAutocomplete when tags have separate identifiers and labels.",
    states: "Selected tags are absent from suggestions. New tags are trimmed and duplicates are ignored regardless of case. required means at least one tag. At the limit the input keeps focus but accepts no additions; removal stays available.",
    content: "Offer common tags as suggestions without making them a fixed vocabulary. State any selection limit in the hint. An empty options array still allows free-text tags.", implementation: formSetup,
    props: [...autocompleteProps.filter((prop) => !["getId", "format", "options"].includes(prop.name)), { name: "options", type: "readonly string[]", required: true, description: "Suggested tags. Users can also create their own strings; selected tags are hidden from this list." }, selectionLimit],
  },
};

export const autocompleteCommonMistakes: Record<string, string> = {
  "component.autocomplete": "Do not persist the typed query as the selected option. Submission stores the chosen object, not arbitrary search text.",
  "component.multi-autocomplete": "Do not use maxResults to limit selection; maxSelections limits the chosen array and should be explained in the hint.",
  "component.tags-input": "Do not intercept Enter on the parent form to save while the tags field is focused. TagsInput uses Enter to add a tag or select the highlighted suggestion.",
};

export const autocompleteFixtures = [
  defineCatalogFixture({ ...{
  "id": "autocomplete.language",
  "name": "Language search",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "autocomplete",
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
}, code: Example23Source, render: () => <Example23 /> }),
  defineCatalogFixture({ ...{
  "id": "autocomplete.language.empty",
  "name": "Language search — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "autocomplete",
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
}, code: Example24Source, render: () => <Example24 /> }),
] as const;

export const autocompleteSpec: DocSpec = { id: "component.autocomplete", route: "/components/forms/autocomplete", name: "Autocomplete", category: "forms", summary: "Filters a long option collection while Base UI manages active-descendant selection.", fixtureId: "autocomplete.language", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteInput"], types: ["AutocompleteInputProps"], when: "Use for a long known collection where searching is materially faster than scanning.", avoid: "Use native Select for short sets and Input when free-form values are valid.", choose: "Autocomplete selects one known object; Input accepts text and Select exposes a short known list.", behavior: "Typing filters, Arrow keys highlight, Enter selects, and Escape dismisses the popup and preserves the selected value, including when the popup is already closed.", accessibility: "Keep the combobox label, expanded state, active option, empty result, and error relationships programmatic.", migration: "Replace Headless UI Combobox and pass stable format/getId functions to Paper." };

export const multiAutocompleteFixtures = [
  defineCatalogFixture({ ...{
  "id": "multi-autocomplete.languages",
  "name": "Multiple languages",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "multi",
    "autocomplete",
    "languages"
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
}, code: Example25Source, render: () => <Example25 /> }),
  defineCatalogFixture({ ...{
  "id": "multi-autocomplete.languages.empty",
  "name": "Multiple languages — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "multi",
    "autocomplete",
    "languages",
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
}, code: Example26Source, render: () => <Example26 /> }),
] as const;

export const multiAutocompleteSpec: DocSpec = { id: "component.multi-autocomplete", route: "/components/forms/multi-autocomplete", name: "MultiAutocomplete", category: "forms", summary: "Selects several known objects from a searchable collection and represents them as removable chips.", fixtureId: "multi-autocomplete.languages", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteMultiInput"], types: ["AutocompleteMultiInputProps"], when: "Use when several values may be selected from one long known collection.", avoid: "Use Checkbox lists for a short stable set and TagsInput for simple string labels.", choose: "MultiAutocomplete preserves object identity; TagsInput is optimized for string tags.", behavior: "Base UI manages filtering and selection; named remove buttons update the React Hook Form array.", accessibility: "Each chip has a named removal action and the input exposes combobox state and results.", migration: "Replace Headless UI multi-combobox and supply stable object identifiers." };

export const tagsInputFixtures = [
  defineCatalogFixture({ ...{
  "id": "tags.entry",
  "name": "Suggested and custom entry tags",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "tags",
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
}, code: Example27Source, render: () => <Example27 /> }),
  defineCatalogFixture({ ...{
  "id": "tags.entry.empty",
  "name": "Entry tags — empty and validation",
  "description": "Edit, save, reset, and inspect the current form values.",
  "tags": [
    "tags",
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
}, code: Example28Source, render: () => <Example28 /> }),
] as const;

export const tagsInputSpec: DocSpec = { id: "component.tags-input", route: "/components/forms/tags-input", name: "TagsInput", category: "forms", summary: "Adds suggested or custom text tags as removable chips.", fixtureId: "tags.entry", sourcePath: "src/components/forms/controls.tsx", react: ["TagsInput"], types: ["TagsInputProps"], when: "Use for short user-defined labels such as reading-entry categories.", avoid: "Use MultiAutocomplete for a fixed vocabulary or objects with separate identifiers and labels. TagsInput does not fetch or persist suggestions.", choose: "TagsInput accepts free text and suggested strings. MultiAutocomplete limits selection to known values.", behavior: "Enter adds the typed tag or chooses the highlighted suggestion without submitting the form. The Add suggestion also supports pointer selection. Adding clears the query and keeps input focus; IME composition never commits a tag. Escape dismisses without clearing tags.", accessibility: "Suggestions expose combobox semantics; selected tags have named removal actions. The input retains focus after adding or removing, including when the selection limit is reached.", migration: "Pass existing tags as suggestions; custom tags are included in the React Hook Form string array and saved by the application." };
