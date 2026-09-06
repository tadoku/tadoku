import Example21 from "./examples/amount.progress";
import Example22 from "./examples/amount.progress.empty";
import Example22Source from "./examples/amount.progress.empty.tsx?raw";
import Example21Source from "./examples/amount.progress.tsx?raw";
import Example23 from "./examples/autocomplete.language";
import Example24 from "./examples/autocomplete.language.empty";
import Example24Source from "./examples/autocomplete.language.empty.tsx?raw";
import Example23Source from "./examples/autocomplete.language.tsx?raw";
import Example30 from "./examples/button-group.entry-actions";
import Example30Source from "./examples/button-group.entry-actions.tsx?raw";
import Example15 from "./examples/checkbox.public-entry";
import Example16 from "./examples/checkbox.public-entry.empty";
import Example16Source from "./examples/checkbox.public-entry.empty.tsx?raw";
import Example15Source from "./examples/checkbox.public-entry.tsx?raw";
import Example32 from "./examples/flash.statuses";
import Example32Source from "./examples/flash.statuses.tsx?raw";
import Example33 from "./examples/loading.entries";
import Example33Source from "./examples/loading.entries.tsx?raw";
import Example25 from "./examples/multi-autocomplete.languages";
import Example26 from "./examples/multi-autocomplete.languages.empty";
import Example26Source from "./examples/multi-autocomplete.languages.empty.tsx?raw";
import Example25Source from "./examples/multi-autocomplete.languages.tsx?raw";
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
import Example13 from "./examples/select.language";
import Example14 from "./examples/select.language.empty";
import Example14Source from "./examples/select.language.empty.tsx?raw";
import Example13Source from "./examples/select.language.tsx?raw";
import Example34 from "./examples/surface.summary";
import Example34Source from "./examples/surface.summary.tsx?raw";
import Example27 from "./examples/tags.entry";
import Example28 from "./examples/tags.entry.empty";
import Example28Source from "./examples/tags.entry.empty.tsx?raw";
import Example27Source from "./examples/tags.entry.tsx?raw";
import Example11 from "./examples/textarea.reading-notes";
import Example12 from "./examples/textarea.reading-notes.empty";
import Example12Source from "./examples/textarea.reading-notes.empty.tsx?raw";
import Example11Source from "./examples/textarea.reading-notes.tsx?raw";
import Example31 from "./examples/toast.entry-saved";
import Example31Source from "./examples/toast.entry-saved.tsx?raw";
import {
defineCatalogDocument,
defineCatalogFixture,
type CatalogDocument,
type ComponentCategory,
type ComponentDocumentationSections,
type ComponentPageSectionKey,
type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-09-05";

interface DocSpec {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly category: ComponentCategory;
  readonly summary: string;
  readonly fixtureId: string;
  readonly sourcePath: string;
  readonly react: readonly string[];
  readonly types: readonly string[];
  readonly css?: readonly string[];
  readonly when: string;
  readonly avoid: string;
  readonly choose: string;
  readonly behavior: string;
  readonly accessibility: string;
  readonly migration: string;
  readonly variants?: string;
  readonly states?: string;
  readonly content?: string;
  readonly pageSections?: readonly ComponentPageSectionKey[];
}

type Prop = NonNullable<CatalogDocument["api"]["props"]>[number];
const fieldProps: readonly Prop[] = [
  { name: "name", type: "string", required: true, description: "Path in the enclosing useForm values. Initialize it through defaultValues, then use reset to load another record." },
  { name: "label", type: "string", required: true, description: "Persistent visible label; do not replace it with a placeholder." },
  { name: "hint", type: "string", description: "Help shown between the label and control and connected through aria-describedby." },
  { name: "required", type: "boolean", defaultValue: "false (RadioGroup: true)", description: "Marks the decision as required and validates it through React Hook Form." },
];
const validationProp: Prop = { name: "rules", type: "RegisterOptions", description: "React Hook Form validation, including custom recovery messages. Use noValidate on the form so inline errors handle submission." };
const optionsProp: Prop = { name: "options", type: "readonly Option<string>[]", required: true, description: "Objects with value, label, and optional disabled. Persist value; show label to the user." };
const disabledProp: Prop = { name: "disabled", type: "boolean", defaultValue: "false", description: "Prevents editing. Explain why the value is unavailable in nearby text." };
const autocompleteProps: readonly Prop[] = [...fieldProps, validationProp,
  { name: "options", type: "readonly Value[]", required: true, description: "Known objects supplied by the application. Fetch remote suggestions outside this component." },
  { name: "format", type: "(option: Value) => string", required: true, description: "Visible option label, also used for case-insensitive substring matching." },
  { name: "getId", type: "(option: Value) => string", required: true, description: "Stable unique identifier used to compare refreshed option objects with selected values." },
  { name: "match", type: "(option: Value, query: string) => boolean", description: "Optional custom matching, for example a label or language-code search." },
  { name: "maxResults", type: "number", defaultValue: "50", description: "Maximum filtered results rendered in the popup; this is not a selection limit." },
  { name: "placeholder", type: "string", description: "A short search prompt shown when the query is empty." }, disabledProp,
];
const selectionLimit: Prop = { name: "maxSelections", type: "number", description: "Maximum selected values. At the limit, remove a chip before adding another value." };
const formSetup = "Import from paper-ui and load paper-ui/styles.css once at the application root. Wrap the form in FormProvider, initialize useForm defaultValues, and submit with methods.handleSubmit. The examples show validation, reset, and locally observable submitted state; replace the local confirmation with your save handler.";
const details: Record<string, { example: string; anatomy: string; variants: string; states: string; content: string; implementation: string; props: readonly Prop[] }> = {
  "component.textarea": {
    example: "Edit the reading note and save it. The empty example demonstrates required validation: submit blank, then enter a note and resubmit to clear the error.",
    anatomy: "A persistent label, optional hint, resizable native textarea, and inline error share one field relationship.",
    variants: "One multiline presentation. Set rows for the expected starting length; the user can resize vertically. Density changes padding, not the meaning of the field.",
    states: "The populated and empty examples share the same required rule. Invalid submission displays an associated recovery message and focuses the field. Native disabled and readOnly have different purposes: unavailable editing versus selectable text that cannot change.",
    content: "Label the value, such as Reading notes. Use the hint for visibility or spoiler policy. Put examples in a placeholder only when they remain useful after the hint.", implementation: formSetup,
    props: [...fieldProps, validationProp, { name: "rows", type: "number", description: "Initial visible text rows; use a larger value for long-form writing." }, { name: "readOnly", type: "boolean", description: "Keeps the text focusable and selectable without allowing edits." }, disabledProp],
  },
  "component.select": {
    example: "Choose a reading language. The second example starts with a placeholder so Save entry demonstrates the required-choice error before selection.",
    anatomy: "Label and hint precede a native select with a decorative chevron; an error follows the control.",
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
  "component.radio-select": {
    example: "The segmented example selects a preview size. Compare it with the native-radio example, then submit an empty group to see how required validation is attached to the decision.",
    anatomy: "A fieldset and legend contain same-name native radios; segmented changes the labels' presentation without changing selection semantics.",
    variants: "Default uses visible native radio indicators; segmented joins concise peer choices.", states: "One option can be selected. Use defaultValues to initialize a required decision; disabled may apply to the whole group or an option.", content: "Use short peer labels.", implementation: formSetup,
    props: [...fieldProps, validationProp, optionsProp, { name: "variant", type: '"default" | "segmented"', defaultValue: '"default"', description: "Default radios for ordinary forms; segmented for two to four short peer choices." }, disabledProp],
  },
  "component.radio-group": {
    example: "Choose Book or Audio and inspect the saved string value. Other is visibly unavailable. The empty example shows the required error for the whole decision.",
    anatomy: "A fieldset/legend contains native radios presented as cards, each with a label and supporting description.",
    variants: "One card presentation. Use RadioSelect for choices that do not need descriptions; adding cards to terse choices adds unnecessary visual weight.",
    states: "Required defaults to true. A disabled option remains visible for explanation but cannot be chosen. Cards stack on narrow screens; long descriptions wrap within their label.",
    content: "Use the label for the format name and the description for the difference that affects the decision. Do not repeat the label in the description.", implementation: formSetup,
    props: [...fieldProps, validationProp, { ...optionsProp, type: "readonly RadioGroupOption<string>[]", description: "Each option requires value, label, and description; disabled is optional." }],
  },
  "component.amount-with-unit": {
    example: "Enter positive progress, then choose pages or minutes. Inspect current values to see progressValue as a number and progressUnit as a separate string. Submit the empty example to see required validation.",
    anatomy: "A labelled group joins one numeric input to one native unit picker. The shared hint and first error describe both controls.",
    variants: "One compound field. Use Input with a unit in its label when the unit is fixed; a disabled picker should not be used to fake a fixed suffix.",
    states: "required, min, and max validate the amount on submission. An empty optional amount becomes NaN through valueAsNumber; normalize it before sending JSON. Disabling the field disables both controls.",
    content: "Label the measurement, such as Progress. Use concise plural unit labels, and state any minimum or rounding policy in the hint.", implementation: formSetup,
    props: [{ name: "name", type: "string", required: true, description: "Prefix for two fields: nameValue (number) and nameUnit (string). Initialize both in defaultValues." }, ...fieldProps.filter((prop) => prop.name !== "name"), { name: "units", type: "readonly Option<string>[]", required: true, description: "Available units. Each has value, label, and optional disabled." }, { name: "unitsLabel", type: "string", description: "Accessible unit-picker label; defaults to Unit for plus the lowercased field label." }, { name: "min / max / step", type: "number | string", description: "Native number constraints. min/max also validate through React Hook Form; use step for the native stepping interval." }, disabledProp],
  },
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
    example: "Choose known reading tags, remove a chip, or select all four to inspect the limit. An unmatched query shows the empty result instead of creating a new tag.",
    anatomy: "A string specialization of MultiAutocomplete: known-string options become removable chips.",
    variants: "Accepts string[] instead of objects, so format and getId are not needed. Use MultiAutocomplete when tags have separate identifiers and labels.",
    states: "Empty, filtered, no-results, selected, and selection-limit states share the same combobox. required means at least one tag. The four-tag limit leaves removal available.",
    content: "Keep tags short and consistent in case. Tell users that only existing tags are available when creation might otherwise be expected.", implementation: formSetup,
    props: [...autocompleteProps.filter((prop) => !["getId", "format", "options"].includes(prop.name)), { name: "options", type: "readonly string[]", required: true, description: "Known tags; each string is both its visible label and identity. Free-text creation is not supported." }, selectionLimit],
  },
  "component.button-group": {
    example: "View log follows an in-preview anchor, Edit log updates the status, and Delete log is disabled with a visible reason. These actions stay independently focusable.",
    anatomy: "A labelled role=group wraps real links and buttons. The wrapper has no toolbar keyboard behavior.",
    variants: "align=start or end positions the group. Individual actions use default, outline, ghost, link, or destructive; keep one primary action per task.",
    states: "Actions wrap when width is limited. visible=false removes an action. Disabled buttons use native disabled; disabled links retain link semantics but block activation with aria-disabled.",
    content: "Name the operation, such as Edit log. The group label describes the shared subject; explain why disabled destructive operations are unavailable.",
    implementation: "Import ButtonGroup from paper-ui. Provide href for navigation and onSelect for an operation; when both exist, onSelect also runs before navigation. The example defines the complete actions list and updates local status.",
    props: [{ name: "actions", type: "readonly ButtonGroupAction[]", required: true, description: "Each action requires id and label. Optional href, onSelect, variant, disabled, and visible determine native link/button behavior." }, { name: "label", type: "string", defaultValue: '"Actions"', description: "Accessible group name." }, { name: "align", type: '"start" | "end"', defaultValue: '"start"', description: "Horizontal alignment; actions wrap on small screens." }],
  },
  "component.flash": {
    example: "Compare all four statuses with identical anatomy. The warning includes a real in-preview recovery link; title and text explain the status without relying on the rail color.",
    anatomy: "A colored square rail, optional decorative icon, title/body, and optional action form one persistent message.", variants: "Information, success, warning, and danger encode distinct outcomes.", states: "visible=false unmounts the message; caller-owned actions keep native semantics.", content: "Name the issue and its recovery.",
    implementation: "Import Flash from paper-ui. Pass body content as children and a real Paper button or link as action. Visibility is caller-controlled; Flash does not add its own dismiss state.",
    props: [{ name: "variant", type: '"information" | "success" | "warning" | "danger"', defaultValue: '"information"', description: "Danger uses role=alert; the other variants use role=status." }, { name: "title", type: "string", description: "Short outcome or problem statement above the body." }, { name: "children", type: "ReactNode", description: "Consequence and recovery copy." }, { name: "action", type: "ReactNode", description: "Caller-owned link or button; no action behavior is inferred." }, { name: "icon", type: "ReactNode", description: "Optional decorative icon, hidden from assistive technology." }, { name: "visible", type: "boolean", defaultValue: "true", description: "False returns null and removes the message from the document." }],
  },
  "component.loading": {
    example: "Compare all three spinner sizes with a visible description of the pending reading operation. The accessible label remains present even with reduced motion enabled.",
    anatomy: "A decorative spinner sits inside a role=status wrapper with a visually hidden label. Visible loading copy must be supplied by the application.",
    variants: "small fits inline operations; default fits a section; large gives a pending page or region stronger emphasis. Size does not indicate percentage progress.",
    states: "Indeterminate only. Reduced motion stops the animation while retaining the shape and label. Unmount Loading when content or an error replaces the pending state.",
    content: "Name what is loading: Loading reading entries. For long waits, also show visible explanatory text so sighted users know which operation is pending.",
    implementation: "Import Loading from paper-ui. Render conditionally while waiting; mark the region being updated aria-busy when appropriate. Keep one useful status announcement per operation.",
    props: [{ name: "label", type: "string", defaultValue: '"Loading"', description: "Visually hidden accessible status text; supply visible copy alongside the spinner when needed." }, { name: "size", type: '"small" | "default" | "large"', defaultValue: '"default"', description: "Visual spinner size, independent of the loading state." }],
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
  "component.toast": {
    example: "Trigger a confirmation or failure, add several notifications to inspect the queue, then dismiss one or all. This teaching preview uses timeout=0; choose a finite timeout in application setup unless persistence is intentional.",
    anatomy: "One provider owns the queue and viewport; each notification contains a title, optional description/action, and named dismiss button.",
    variants: "low priority announces politely; high interrupts and belongs to urgent failures. The type field is metadata and does not currently select a Paper color variant.",
    states: "ToastProvider defaults follow Base UI: a five-second timeout and three visible toasts. timeout=0 disables expiry. Hover/focus pauses expiry; closing removes a notification through the manager. Keep durable recovery beside the affected content.",
    content: "Use a completed outcome for confirmation, such as Entry saved. An error should explain what remains safe and where to retry. Do not put the only copy of a critical recovery action in a timed message.",
    implementation: "Mount ToastProvider once above components calling useToast. The hook must run inside that provider, not in the same component that creates it. Use add for a new message, update for an existing id, close for dismissal, and promise for loading/success/error transitions. ToastContainer is a migration alias for the provider.",
    props: [{ name: "ToastProvider.timeout", type: "number", defaultValue: "5000", description: "Milliseconds before dismissal; 0 disables automatic dismissal." }, { name: "ToastProvider.limit", type: "number", defaultValue: "3", description: "Maximum notifications visible at once." }, { name: "add(options)", type: "(ToastOptions) => string", description: "Returns an id. Options include title, description, priority, timeout, actionProps, onClose, and onRemove." }, { name: "close(id?)", type: "(string?) => void", description: "Dismiss one notification, or all when the id is omitted." }, { name: "update(id, options)", type: "(string, ToastUpdateOptions) => void", description: "Change an existing notification without adding another." }, { name: "promise(promise, options)", type: "<Value>(Promise<Value>, ToastPromiseOptions<Value>) => Promise<Value>", description: "Supply loading, success, and error messages or callbacks for one asynchronous operation." }],
  },
};

function sections(spec: DocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [details[spec.id].anatomy] },
    recommendedExample: { heading: "Recommended example", content: [details[spec.id].example] },
    variants: { heading: "Variants", content: [spec.variants ?? details[spec.id].variants] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states ?? details[spec.id].states] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content ?? details[spec.id].content] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [details[spec.id].implementation] },
    apiReference: { heading: "API reference", content: [`Public React APIs: ${spec.react.join(", ")}. Public types: ${spec.types.join(", ") || "none"}.`] },
    relatedPatterns: { heading: "Related patterns", content: [`Combine ${spec.name} with Button and the nearest documented form, feedback, or action pattern rather than inventing local structure.`] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: ["Published in Paper 0.1.0. Review the documented limits before application migration; Stable describes API availability, not exhaustive product coverage."] },
  };
  return {
    required,
    pageSections: spec.pageSections ?? ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"],
  };
}

const commonMistakes: Record<string, string> = {
  "component.textarea": "Do not set a fixed height with overflow hidden: users need to read and resize longer notes.",
  "component.select": "Do not store a placeholder as a real language value. Initialize the empty option with an empty string and validate required selections.",
  "component.checkbox": "Do not mark an optional preference required; required means the user must check it to continue.",
  "component.radio-select": "Do not use segmented radios for commands. A radio changes one stored value; a button runs an operation.",
  "component.radio-group": "Do not hide the distinction between options in a tooltip. Keep decision-making descriptions visible beside each radio.",
  "component.amount-with-unit": "Do not initialize a single progress object: this component registers progressValue and progressUnit as two sibling fields.",
  "component.autocomplete": "Do not persist the typed query as the selected option. Submission stores the chosen object, not arbitrary search text.",
  "component.multi-autocomplete": "Do not use maxResults to limit selection; maxSelections limits the chosen array and should be explained in the hint.",
  "component.tags-input": "Do not imply that Enter creates a new tag. Enter selects a known matching string only.",
  "component.button-group": "Do not add toolbar arrow-key expectations: each link or button is reached independently with Tab.",
  "component.flash": "Do not use a danger rail for neutral context; danger also interrupts assistive technology with an alert.",
  "component.loading": "Do not assume label is visible: it is visually hidden. Supply visible waiting copy when users need more than a spinner.",
  "component.surface": "Do not treat an accent rail or shadow as an interactive affordance. Surface has no click or selection behavior.",
  "component.toast": "Do not call useToast above its provider or duplicate providers for every action. Mount one provider around the application area."
};

function document(spec: DocSpec): CatalogDocument {
  return defineCatalogDocument({
    id: spec.id,
    route: spec.route,
    name: spec.name,
    kind: "component",
    category: spec.category,
    aliases: [],
    summary: spec.summary,
    keywords: [spec.name.toLocaleLowerCase(), spec.category, "paper"],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    sourcePath: spec.sourcePath,
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: [spec.when],
      whenNotToUse: [spec.avoid],
      content: [spec.content ?? details[spec.id].content],
      commonMistakes: [commonMistakes[spec.id]],
    },
    accessibility: {
      requirements: [spec.accessibility],
      keyboard: [spec.behavior],
      knownConstraints: [details[spec.id].states],
    },
    api: {
      react: spec.react,
      props: details[spec.id].props,
      cssClasses: spec.css ?? [],
      publicTypes: spec.types,
      defaults: ["Theme and density inherit from the nearest Paper root."],
      invalidCombinations: [spec.avoid],
    },
    fixtureIds: [spec.fixtureId, ...(spec.category === "forms" ? [`${spec.fixtureId}.empty`] : []), ...(spec.id === "component.radio-select" ? ["radio-select.native"] : [])],
    dependencies: { documents: ["component.button"], packages: spec.category === "forms" ? ["paper-ui", "react-hook-form"] : ["paper-ui"] },
    migration: { legacy: [`ui ${spec.name}`], notes: [spec.migration] },
    changelog: [{ date: REVIEW_DATE, note: `Revised ${spec.name} with complete runnable examples, state guidance, and concrete prop contracts.` }],
    behaviorTestIds: [`${spec.id}.semantics`, `${spec.id}.behavior`],
    sections: sections(spec),
  });
}

export const phaseThreeFormsFeedbackFixtures = [
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
  defineCatalogFixture({ ...{
  "id": "tags.entry",
  "name": "Known entry tags",
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
  "name": "Known entry tags — empty and validation",
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
  defineCatalogFixture({ ...{
  "id": "toast.entry-saved",
  "name": "Queued notifications",
  "description": "Run the actions and inspect their visible result.",
  "tags": [
    "toast",
    "entry",
    "saved"
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
}, code: Example31Source, render: () => <Example31 /> }),
  defineCatalogFixture({ ...{
  "id": "flash.statuses",
  "name": "Feedback statuses",
  "description": "Compare the supported presentations with consistent reading-log content.",
  "tags": [
    "flash",
    "statuses"
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
}, code: Example32Source, render: () => <Example32 /> }),
  defineCatalogFixture({ ...{
  "id": "loading.entries",
  "name": "Loading sizes and visible context",
  "description": "Compare the supported presentations with consistent reading-log content.",
  "tags": [
    "loading",
    "entries"
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
}, code: Example33Source, render: () => <Example33 /> }),
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

const specs: readonly DocSpec[] = [
  { id: "component.textarea", route: "/components/forms/textarea", name: "TextArea", category: "forms", summary: "Collects multiline text with the same deterministic field anatomy as Input.", fixtureId: "textarea.reading-notes", sourcePath: "src/components/forms/controls.tsx", react: ["TextArea"], types: ["TextAreaProps"], css: ["paper-textarea"], when: "Use for prose, notes, and other values that genuinely need multiple lines.", avoid: "Use Input for short single-line values and do not resize away the user's editing space.", choose: "Choose TextArea over Input by content length, not by visual preference.", behavior: "Native text editing, Tab focus, and React Hook Form validation remain intact.", accessibility: "Associate the persistent label, hint, and recoverable error with the native textarea.", migration: "Replace legacy TextArea while retaining the existing field name and register rules." },
  { id: "component.select", route: "/components/forms/select", name: "Select", category: "forms", summary: "Uses the native platform picker for a known set of concise choices.", fixtureId: "select.language", sourcePath: "src/components/forms/controls.tsx", react: ["Select"], types: ["SelectProps", "Option", "OptionGroup"], when: "Use when one value must be chosen from a stable, reasonably short option set.", avoid: "Use Autocomplete for long searchable collections and RadioSelect when every choice should remain visible.", choose: "Prefer native Select unless a documented requirement needs composite listbox behavior.", behavior: "Platform keyboard, touch, and form submission behavior is preserved.", accessibility: "Provide a visible label and meaningful option labels; disabled options remain programmatically unavailable.", migration: "Map legacy values and groups directly without copying global select selectors." },
  { id: "component.checkbox", route: "/components/forms/checkbox", name: "Checkbox", category: "forms", summary: "Captures an independent boolean choice with native semantics.", fixtureId: "checkbox.public-entry", sourcePath: "src/components/forms/controls.tsx", react: ["Checkbox"], types: ["CheckboxProps"], when: "Use for an independent yes/no setting or acknowledgement.", avoid: "Use radio choices when exactly one option in a set must be selected.", choose: "A checkbox toggles one proposition; a radio group chooses among mutually exclusive propositions.", behavior: "Space toggles the focused native checkbox and React Hook Form owns its boolean value.", accessibility: "Put the consequence in the visible label or hint and retain the native checkbox input.", migration: "Replace legacy Checkbox without converting it into a custom switch role." },
  { id: "component.radio-select", route: "/components/forms/radio-select", name: "RadioSelect", category: "forms", summary: "Shows concise mutually exclusive options as native radios, with an optional connected segmented presentation.", fixtureId: "radio-select.viewport", sourcePath: "src/components/forms/controls.tsx", react: ["RadioSelect"], types: ["RadioSelectProps", "RadioSelectVariant", "Option"], when: "Use for two to four short choices that benefit from remaining visible; use segmented for compact peer choices such as preview sizes.", avoid: "Use Select for longer sets, RadioGroup when each choice needs description copy, and Button when activation does not select a form value.", choose: "Default exposes familiar radio controls; segmented visually connects short peers without changing their native semantics.", variants: "Default presents inline native radios. Segmented presents the same inputs as a connected quiet-surface group with an accent edge and check for the selected value.", states: "Both variants cover selected, unselected, focus-visible, disabled, and invalid states in light, dark, comfortable, compact, narrow, and forced-color settings.", behavior: "Arrow keys move within the native same-name radio set and Space selects. Segmented remains a form control and never exposes button or aria-pressed semantics.", content: "Use two to four parallel, concise nouns such as Phone, Tablet, and Desktop; require and initialize one value when the decision cannot be empty.", accessibility: "A fieldset and legend name the group while each label names its native radio. The segmented check and accent edge reinforce selection without replacing checked state.", migration: 'Preserve submitted option values and replace button-like viewport selectors with variant="segmented" rather than recreating toggle-button semantics.', pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] },
  { id: "component.radio-group", route: "/components/forms/radio-group", name: "RadioGroup", category: "forms", summary: "Presents mutually exclusive choices that each require supporting description.", fixtureId: "radio-group.format", sourcePath: "src/components/forms/controls.tsx", react: ["RadioGroup"], types: ["RadioGroupProps", "RadioGroupOption"], when: "Use when each mutually exclusive choice needs a label and explanatory description.", avoid: "Use RadioSelect for terse choices and Checkbox for independent choices.", choose: "Choose the smallest radio presentation that gives users enough context to decide.", behavior: "Native radio focus and arrow-key selection work without a custom composite widget.", accessibility: "Use a legend for the decision, preserve disabled semantics, and never rely on the selected border alone.", migration: "Remove Headless UI RadioGroup imports; Paper uses native radios and React Hook Form." },
  { id: "component.amount-with-unit", route: "/components/forms/amount-with-unit", name: "AmountWithUnit", category: "forms", summary: "Groups a numeric amount with the unit that gives it meaning.", fixtureId: "amount.progress", sourcePath: "src/components/forms/controls.tsx", react: ["AmountWithUnit"], types: ["AmountWithUnitProps", "Option"], when: "Use when a numeric value is invalid or ambiguous without one unit from a short set.", avoid: "Use Input alone when the unit is fixed and can be stated in the label.", choose: "Keep the unit fixed in content when users cannot change it; otherwise group amount and native Select.", behavior: "The amount and unit submit as nameValue and nameUnit for legacy-compatible migration.", accessibility: "Name the amount visibly and provide a generated accessible name for the unit picker.", migration: "Retain the legacy paired field names while replacing its implicit layout and error selectors." },
  { id: "component.autocomplete", route: "/components/forms/autocomplete", name: "Autocomplete", category: "forms", summary: "Filters a long option collection while Base UI manages active-descendant selection.", fixtureId: "autocomplete.language", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteInput"], types: ["AutocompleteInputProps"], when: "Use for a long known collection where searching is materially faster than scanning.", avoid: "Use native Select for short sets and Input when free-form values are valid.", choose: "Autocomplete selects one known object; Input accepts text and Select exposes a short known list.", behavior: "Typing filters, Arrow keys highlight, Enter selects, and Escape dismisses the popup and preserves the selected value, including when the popup is already closed.", accessibility: "Keep the combobox label, expanded state, active option, empty result, and error relationships programmatic.", migration: "Replace Headless UI Combobox and pass stable format/getId functions to Paper." },
  { id: "component.multi-autocomplete", route: "/components/forms/multi-autocomplete", name: "MultiAutocomplete", category: "forms", summary: "Selects several known objects from a searchable collection and represents them as removable chips.", fixtureId: "multi-autocomplete.languages", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteMultiInput"], types: ["AutocompleteMultiInputProps"], when: "Use when several values may be selected from one long known collection.", avoid: "Use Checkbox lists for a short stable set and TagsInput for simple string labels.", choose: "MultiAutocomplete preserves object identity; TagsInput is optimized for string tags.", behavior: "Base UI manages filtering and selection; named remove buttons update the React Hook Form array.", accessibility: "Each chip has a named removal action and the input exposes combobox state and results.", migration: "Replace Headless UI multi-combobox and supply stable object identifiers." },
  { id: "component.tags-input", route: "/components/forms/tags-input", name: "TagsInput", category: "forms", summary: "Selects a bounded set of known string tags with removable chip semantics.", fixtureId: "tags.entry", sourcePath: "src/components/forms/controls.tsx", react: ["TagsInput"], types: ["TagsInputProps"], when: "Use for several short known labels such as reading-entry categories.", avoid: "Do not use for arbitrary server-backed creation or object values; use MultiAutocomplete instead.", choose: "TagsInput is the string specialization of MultiAutocomplete and does not hide asynchronous data fetching.", behavior: "Typing filters known tags, Enter selects, Backspace and named controls support removal, and limits disable further input.", accessibility: "Announce the combobox results and give every selected tag an explicit remove label.", migration: "Move suggestion data loading to the application and pass a deterministic option set to Paper." },
  { id: "component.button-group", route: "/components/actions/button-group", name: "ButtonGroup", category: "actions", summary: "Keeps related visible actions together without changing link or button semantics.", fixtureId: "button-group.entry-actions", sourcePath: "src/components/feedback/feedback.tsx", react: ["ButtonGroup"], types: ["ButtonGroupProps", "ButtonGroupAction"], when: "Use for a small, stable set of related actions that should remain visible.", avoid: "Use ActionMenu when the set is contextual or too large for the available width.", choose: "ButtonGroup prioritizes visibility; ActionMenu prioritizes compact contextual access.", behavior: "Each entry remains a real anchor or button and the wrapper contributes only a labelled group.", accessibility: "Preserve native roles, express disabled links with aria-disabled plus blocked activation, and label the group.", migration: "Replace responsive Next Link coupling with router-neutral href or onSelect actions." },
  { id: "component.flash", route: "/components/feedback/flash", name: "Flash", category: "feedback", summary: "Places persistent contextual feedback beside the content it describes.", fixtureId: "flash.statuses", sourcePath: "src/components/feedback/feedback.tsx", react: ["Flash"], types: ["FlashProps", "FlashVariant"], when: "Use for page or section feedback that should remain until the context changes, such as a saved result, a contest warning, or a failed page-level operation.", avoid: "Use Toast for brief asynchronous confirmation and an inline field error when the user must repair one value. Do not put required recovery only in a Flash that may disappear with its surrounding content.", choose: "Flash stays beside the content it explains. Toast is transient and global; field errors identify a specific control and its recovery.", variants: "Information covers neutral context, success confirms a completed outcome, and warning calls for review before proceeding. Danger is reserved for an urgent failure and changes the live-region role from status to alert.", states: "Setting visible=false removes the message from the document. Actions are caller-owned links or buttons, so their label, disabled state, and navigation behavior must remain explicit.", behavior: "Danger feedback is announced as an alert; information, success, and warning use a polite status announcement. Changing visible to false unmounts the message instead of merely hiding it.", content: "Write the title as the outcome or issue, then use the body for consequence and recovery. Keep any action label specific\u2014Review fields or Try again instead of OK.", accessibility: "Pair the colored edge with a title and message text, because color does not name the status. Provide actions as real links or buttons and reserve danger for messages that warrant an interrupting alert.", migration: "Map error to danger, remove Next Link ownership, and pass any action as router-owned React content.", pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] },
  { id: "component.loading", route: "/components/feedback/loading", name: "Loading", category: "feedback", summary: "Communicates indeterminate progress with a named, reduced-motion-safe status.", fixtureId: "loading.entries", sourcePath: "src/components/feedback/feedback.tsx", react: ["Loading"], types: ["LoadingProps"], when: "Use while a bounded region is waiting and no meaningful progress value exists.", avoid: "Use native progress for measurable work and retain existing content when optimistic updates are safer.", choose: "Loading indicates indeterminate work; Button loading prevents repeat action while keeping its name.", behavior: "The status label is available to assistive technology and animation stops under reduced motion.", accessibility: "Supply a label that names what is loading and do not rely on spinner motion alone.", migration: "Replace legacy Loading and remove application-owned spinner SVGs." },
  { id: "component.surface", route: "/components/data-display/surface", name: "Surface", category: "data-display", summary: "Composes flat, floating, or showcase content with optional Paper accent rail.", fixtureId: "surface.summary", sourcePath: "src/components/feedback/feedback.tsx", react: ["Surface", "surfaceClassName"], types: ["SurfaceProps"], css: ["paper-surface-card", "paper-elevation-*"], when: "Use to group related content when a semantic article, section, or neutral container also needs Paper surface treatment.", avoid: "Do not wrap every block in a card or use elevation as decoration without hierarchy.", choose: "Flat separates with rules; floating marks transient or interactive content; showcase is reserved for strong demonstrations.", behavior: "Surface adds no interaction and preserves the semantics selected with the as prop.", accessibility: "Choose article or section only when the content merits that landmark; elevation cannot replace headings.", migration: "Replace copied card and shadow classes with the shared recipe or Surface." },
  { id: "component.toast", route: "/components/feedback/toast", name: "Toast", category: "feedback", summary: "Queues brief asynchronous notifications with polite/urgent priority and explicit dismissal.", fixtureId: "toast.entry-saved", sourcePath: "src/components/feedback/feedback.tsx", react: ["ToastProvider", "useToast", "ToastContainer"], types: ["ToastProviderProps"], when: "Use for brief confirmation or failure that is not tied to one visible field or section.", avoid: "Use Flash for persistent contextual feedback and never put required recovery only in an expiring toast.", choose: "Toast is transient global feedback; Flash remains in context until it is no longer relevant.", behavior: "Base UI owns queueing, live announcements, timeout pause, swipe, and stacking behind the Paper provider.", accessibility: "Choose low or high priority intentionally, keep a visible dismiss action, and repeat critical recovery in context.", migration: "Replace react-toastify setup with ToastProvider and useToast; ToastContainer remains a migration alias." }
];

export const phaseThreeFormsFeedbackDocuments = specs.map(document);
