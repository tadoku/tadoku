import type { ReactNode } from "react";
import { FormProvider, useForm } from "react-hook-form";
import {
  AmountWithUnit,
  AutocompleteInput,
  AutocompleteMultiInput,
  Button,
  ButtonGroup,
  Checkbox,
  Flash,
  Loading,
  RadioGroup,
  RadioSelect,
  Select,
  Surface,
  TagsInput,
  TextArea,
  ToastProvider,
  useToast,
} from "../index";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type ComponentCategory,
  type ComponentDocumentationSections,
  type ComponentPageSectionKey,
  type RequiredComponentSections,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

const LANGUAGES = [
  { id: "ja", label: "Japanese" },
  { id: "zh", label: "Chinese" },
  { id: "ko", label: "Korean" },
] as const;

type FixtureKind =
  | "textarea"
  | "select"
  | "checkbox"
  | "radio-select"
  | "radio-group"
  | "amount"
  | "autocomplete"
  | "multi-autocomplete"
  | "tags";

function FormFixture({ kind }: { kind: FixtureKind }) {
  const methods = useForm({
    defaultValues: {
      notes: "Finished chapter two. The pacing is picking up.",
      language: "ja",
      public: true,
      viewport: "desktop",
      format: "book",
      progressValue: 48,
      progressUnit: "pages",
      autocomplete: LANGUAGES[0],
      multiple: [LANGUAGES[0]],
      tags: ["fiction"],
    },
  });
  return (
    <FormProvider {...methods}>
      <form className="paper-fixture-form" noValidate onSubmit={methods.handleSubmit(() => undefined)}>
        {kind === "textarea" ? <TextArea name="notes" label="Reading notes" hint="Keep spoilers out of public notes." required /> : null}
        {kind === "select" ? <Select name="language" label="Language" hint="Choose the language you read." options={LANGUAGES.map(({ id, label }) => ({ value: id, label }))} /> : null}
        {kind === "checkbox" ? <Checkbox name="public" label="Show this entry on my profile" hint="Contest moderators can always review submitted entries." /> : null}
        {kind === "radio-select" ? <RadioSelect name="viewport" label="Preview size" hint="Choose one viewport for the isolated preview." variant="segmented" required options={[{ value: "phone", label: "Phone" }, { value: "tablet", label: "Tablet" }, { value: "desktop", label: "Desktop" }]} /> : null}
        {kind === "radio-group" ? <RadioGroup name="format" label="Reading format" options={[{ value: "book", label: "Book", description: "Printed or digital text" }, { value: "audio", label: "Audio", description: "Narrated content" }, { value: "other", label: "Other", description: "Another eligible format", disabled: true }]} /> : null}
        {kind === "amount" ? <AmountWithUnit name="progress" label="Progress" units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes" }]} /> : null}
        {kind === "autocomplete" ? <AutocompleteInput name="autocomplete" label="Language" options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} placeholder="Search languages" /> : null}
        {kind === "multi-autocomplete" ? <AutocompleteMultiInput name="multiple" label="Languages" hint="Select every language used in the entry." options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} placeholder="Add language" /> : null}
        {kind === "tags" ? <TagsInput name="tags" label="Tags" options={["fiction", "history", "manga", "nonfiction"]} maxSelections={4} placeholder="Add tag" /> : null}
        <Button type="submit">Save entry</Button>
      </form>
    </FormProvider>
  );
}

function ToastFixture() {
  const toast = useToast();
  return <Button onClick={() => toast.add({ title: "Entry saved", description: "48 pages added to August Japanese." })}>Show notification</Button>;
}

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

function sections(spec: DocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [`${spec.name} combines semantic HTML, Paper field or surface structure, and explicit state styling.`] },
    recommendedExample: { heading: "Recommended example", content: [`The Tadoku fixture demonstrates ${spec.name.toLocaleLowerCase()} with realistic reading-log content and deterministic values.`] },
    variants: { heading: "Variants", content: [spec.variants ?? `Use only the variants exposed by ${spec.react.join(" and ")}; visual differences must represent a documented semantic job.`] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states ?? "Review empty, populated, disabled, invalid, long-content, light/dark, comfortable/compact, and narrow states where applicable."] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content ?? `Use concrete Tadoku nouns and short recovery-oriented messages in ${spec.name}.`] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [`Import ${spec.react.join(", ")} from paper-ui and load paper-ui/styles.css once. Form values remain owned by react-hook-form.`] },
    apiReference: { heading: "API reference", content: [`Public React APIs: ${spec.react.join(", ")}. Public types: ${spec.types.join(", ") || "none"}.`] },
    relatedPatterns: { heading: "Related patterns", content: [`Combine ${spec.name} with Button and the nearest documented form, feedback, or action pattern rather than inventing local structure.`] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: [`Stable in Paper 0.1.0 with deterministic fixtures and behavior coverage.`] },
  };
  return {
    required,
    // These generated pages have specific usage, behavior, and accessibility
    // evidence. Generic variant/content boilerplate stays in registry evidence
    // until an author replaces it with component-specific guidance.
    pageSections: spec.pageSections ?? ["usage", "examples", "behavior", "accessibility"],
  };
}

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
      content: [`Use concise, specific labels and recovery copy in ${spec.name}.`],
      commonMistakes: [spec.choose],
    },
    accessibility: {
      requirements: [spec.accessibility],
      keyboard: [spec.behavior],
      knownConstraints: [spec.avoid],
    },
    api: {
      react: spec.react,
      cssClasses: spec.css ?? [],
      publicTypes: spec.types,
      defaults: ["Theme and density inherit from the nearest Paper root."],
      invalidCombinations: [spec.avoid],
    },
    fixtureIds: [spec.fixtureId],
    dependencies: { documents: ["component.button"], packages: ["paper-ui", "react-hook-form"] },
    migration: { legacy: [`ui ${spec.name}`], notes: [spec.migration] },
    changelog: [{ date: REVIEW_DATE, note: `Published ${spec.name} as Stable.` }],
    behaviorTestIds: [`${spec.id}.semantics`, `${spec.id}.behavior`],
    sections: sections(spec),
  });
}

const fixture = (
  id: string,
  name: string,
  description: string,
  code: string,
  render: () => ReactNode,
) => defineCatalogFixture({
  id,
  name,
  description,
  tags: id.split(/[.-]/u),
  themes: ["light", "dark"],
  densities: ["comfortable", "compact"],
  viewports: VIEWPORTS,
  deterministic: true,
  code,
  render,
});

export const phaseThreeFormsFeedbackFixtures = [
  fixture("textarea.reading-notes", "Reading notes", "A required multiline reading note with spoiler guidance.", `import { TextArea } from "paper-ui";\n\n<TextArea name="notes" label="Reading notes" hint="Keep spoilers out of public notes." required />`, () => <FormFixture kind="textarea" />),
  fixture("select.language", "Language select", "A native language selection retaining platform behavior.", `import { Select } from "paper-ui";\n\n<Select name="language" label="Language" options={languages} />`, () => <FormFixture kind="select" />),
  fixture("checkbox.public-entry", "Public entry", "An optional boolean setting with consequence guidance.", `import { Checkbox } from "paper-ui";\n\n<Checkbox name="public" label="Show this entry on my profile" />`, () => <FormFixture kind="checkbox" />),
  fixture("radio-select.viewport", "Preview viewport", "A connected segmented presentation of three required native radio choices.", `import { RadioSelect } from "paper-ui";\n\n<RadioSelect name="viewport" label="Preview size" variant="segmented" required options={viewports} />`, () => <FormFixture kind="radio-select" />),
  fixture("radio-group.format", "Reading format cards", "Described card choices with a disabled state.", `import { RadioGroup } from "paper-ui";\n\n<RadioGroup name="format" label="Reading format" options={formats} />`, () => <FormFixture kind="radio-group" />),
  fixture("amount.progress", "Progress with unit", "Numeric progress paired with an explicit unit.", `import { AmountWithUnit } from "paper-ui";\n\n<AmountWithUnit name="progress" label="Progress" units={units} />`, () => <FormFixture kind="amount" />),
  fixture("autocomplete.language", "Language autocomplete", "A keyboard-filterable single language selection.", `import { AutocompleteInput } from "paper-ui";\n\n<AutocompleteInput name="language" label="Language" options={languages} format={formatLanguage} getId={getLanguageId} />`, () => <FormFixture kind="autocomplete" />),
  fixture("multi-autocomplete.languages", "Multiple languages", "Several selected languages represented as removable chips.", `import { AutocompleteMultiInput } from "paper-ui";\n\n<AutocompleteMultiInput name="languages" label="Languages" options={languages} format={formatLanguage} getId={getLanguageId} />`, () => <FormFixture kind="multi-autocomplete" />),
  fixture("tags.entry", "Entry tags", "A bounded set of known reading-entry tags.", `import { TagsInput } from "paper-ui";\n\n<TagsInput name="tags" label="Tags" options={["fiction", "history", "manga"]} />`, () => <FormFixture kind="tags" />),
  fixture("button-group.entry-actions", "Entry actions", "Router-neutral links and actions with explicit hierarchy.", `import { ButtonGroup } from "paper-ui";\n\n<ButtonGroup label="Entry actions" actions={actions} />`, () => <ButtonGroup label="Entry actions" actions={[{ id: "view", label: "View log", href: "#view", variant: "outline" }, { id: "edit", label: "Edit log", onSelect: () => undefined }, { id: "delete", label: "Delete log", onSelect: () => undefined, variant: "destructive" }]} />),
  fixture("flash.statuses", "Feedback statuses", "Information, success, warning, and danger use text plus edge color.", `import { Flash } from "paper-ui";\n\n<Flash variant="success" title="Entry saved">48 pages added.</Flash>`, () => <div className="paper-fixture-stack"><Flash title="Contest note">Only finished reading counts.</Flash><Flash variant="success" title="Entry saved">48 pages added.</Flash><Flash variant="warning" title="Check the date">This entry is outside the contest window.</Flash><Flash variant="danger" title="Could not save">Review the highlighted fields.</Flash></div>),
  fixture("loading.entries", "Loading entries", "A named progress status that respects reduced motion.", `import { Loading } from "paper-ui";\n\n<Loading label="Loading reading entries" />`, () => <Loading label="Loading reading entries" size="large" />),
  fixture("surface.summary", "Reading summary surface", "Flat, floating, and accented composition surfaces.", `import { Surface } from "paper-ui";\n\n<Surface as="article" elevation="floating" accent>Reading summary</Surface>`, () => <Surface as="article" elevation="floating" accent><h3>August Japanese</h3><p>1,240 pages across 18 entries.</p></Surface>),
  fixture("toast.entry-saved", "Entry-saved toast", "A polite queued notification with explicit dismissal.", `import { Button, ToastProvider, useToast } from "paper-ui";\n\n<ToastProvider><SaveNotification /></ToastProvider>`, () => <ToastProvider timeout={0}><ToastFixture /></ToastProvider>),
] as const;

const specs: readonly DocSpec[] = [
  { id: "component.textarea", route: "/components/forms/textarea", name: "TextArea", category: "forms", summary: "Collects multiline text with the same deterministic field anatomy as Input.", fixtureId: "textarea.reading-notes", sourcePath: "src/components/forms/controls.tsx", react: ["TextArea"], types: ["TextAreaProps"], css: ["paper-textarea"], when: "Use for prose, notes, and other values that genuinely need multiple lines.", avoid: "Use Input for short single-line values and do not resize away the user's editing space.", choose: "Choose TextArea over Input by content length, not by visual preference.", behavior: "Native text editing, Tab focus, and React Hook Form validation remain intact.", accessibility: "Associate the persistent label, hint, and recoverable error with the native textarea.", migration: "Replace legacy TextArea while retaining the existing field name and register rules." },
  { id: "component.select", route: "/components/forms/select", name: "Select", category: "forms", summary: "Uses the native platform picker for a known set of concise choices.", fixtureId: "select.language", sourcePath: "src/components/forms/controls.tsx", react: ["Select"], types: ["SelectProps", "Option", "OptionGroup"], when: "Use when one value must be chosen from a stable, reasonably short option set.", avoid: "Use Autocomplete for long searchable collections and RadioSelect when every choice should remain visible.", choose: "Prefer native Select unless a documented requirement needs composite listbox behavior.", behavior: "Platform keyboard, touch, and form submission behavior is preserved.", accessibility: "Provide a visible label and meaningful option labels; disabled options remain programmatically unavailable.", migration: "Map legacy values and groups directly without copying global select selectors." },
  { id: "component.checkbox", route: "/components/forms/checkbox", name: "Checkbox", category: "forms", summary: "Captures an independent boolean choice with native semantics.", fixtureId: "checkbox.public-entry", sourcePath: "src/components/forms/controls.tsx", react: ["Checkbox"], types: ["CheckboxProps"], when: "Use for an independent yes/no setting or acknowledgement.", avoid: "Use radio choices when exactly one option in a set must be selected.", choose: "A checkbox toggles one proposition; a radio group chooses among mutually exclusive propositions.", behavior: "Space toggles the focused native checkbox and React Hook Form owns its boolean value.", accessibility: "Put the consequence in the visible label or hint and retain the native checkbox input.", migration: "Replace legacy Checkbox without converting it into a custom switch role." },
  { id: "component.radio-select", route: "/components/forms/radio-select", name: "RadioSelect", category: "forms", summary: "Shows concise mutually exclusive options as native radios, with an optional connected segmented presentation.", fixtureId: "radio-select.viewport", sourcePath: "src/components/forms/controls.tsx", react: ["RadioSelect"], types: ["RadioSelectProps", "RadioSelectVariant", "Option"], when: "Use for two to four short choices that benefit from remaining visible; use segmented for compact peer choices such as preview sizes.", avoid: "Use Select for longer sets, RadioGroup when each choice needs description copy, and Button when activation does not select a form value.", choose: "Default exposes familiar radio controls; segmented visually connects short peers without changing their native semantics.", variants: "Default presents inline native radios. Segmented presents the same inputs as a connected quiet-surface group with an accent edge and check for the selected value.", states: "Both variants cover selected, unselected, focus-visible, disabled, and invalid states in light, dark, comfortable, compact, narrow, and forced-color settings.", behavior: "Arrow keys move within the native same-name radio set and Space selects. Segmented remains a form control and never exposes button or aria-pressed semantics.", content: "Use two to four parallel, concise nouns such as Phone, Tablet, and Desktop; require and initialize one value when the decision cannot be empty.", accessibility: "A fieldset and legend name the group while each label names its native radio. The segmented check and accent edge reinforce selection without replacing checked state.", migration: "Preserve submitted option values and replace button-like viewport selectors with variant=\"segmented\" rather than recreating toggle-button semantics.", pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] },
  { id: "component.radio-group", route: "/components/forms/radio-group", name: "RadioGroup", category: "forms", summary: "Presents mutually exclusive choices that each require supporting description.", fixtureId: "radio-group.format", sourcePath: "src/components/forms/controls.tsx", react: ["RadioGroup"], types: ["RadioGroupProps", "RadioGroupOption"], when: "Use when each mutually exclusive choice needs a label and explanatory description.", avoid: "Use RadioSelect for terse choices and Checkbox for independent choices.", choose: "Choose the smallest radio presentation that gives users enough context to decide.", behavior: "Native radio focus and arrow-key selection work without a custom composite widget.", accessibility: "Use a legend for the decision, preserve disabled semantics, and never rely on the selected border alone.", migration: "Remove Headless UI RadioGroup imports; Paper uses native radios and React Hook Form." },
  { id: "component.amount-with-unit", route: "/components/forms/amount-with-unit", name: "AmountWithUnit", category: "forms", summary: "Groups a numeric amount with the unit that gives it meaning.", fixtureId: "amount.progress", sourcePath: "src/components/forms/controls.tsx", react: ["AmountWithUnit"], types: ["AmountWithUnitProps", "Option"], when: "Use when a numeric value is invalid or ambiguous without one unit from a short set.", avoid: "Use Input alone when the unit is fixed and can be stated in the label.", choose: "Keep the unit fixed in content when users cannot change it; otherwise group amount and native Select.", behavior: "The amount and unit submit as nameValue and nameUnit for legacy-compatible migration.", accessibility: "Name the amount visibly and provide a generated accessible name for the unit picker.", migration: "Retain the legacy paired field names while replacing its implicit layout and error selectors." },
  { id: "component.autocomplete", route: "/components/forms/autocomplete", name: "Autocomplete", category: "forms", summary: "Filters a long option collection while Base UI manages active-descendant selection.", fixtureId: "autocomplete.language", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteInput"], types: ["AutocompleteInputProps"], when: "Use for a long known collection where searching is materially faster than scanning.", avoid: "Use native Select for short sets and Input when free-form values are valid.", choose: "Autocomplete selects one known object; Input accepts text and Select exposes a short known list.", behavior: "Typing filters, Arrow keys highlight, Enter selects, and Escape dismisses without exposing Base UI.", accessibility: "Keep the combobox label, expanded state, active option, empty result, and error relationships programmatic.", migration: "Replace Headless UI Combobox and pass stable format/getId functions to Paper." },
  { id: "component.multi-autocomplete", route: "/components/forms/multi-autocomplete", name: "MultiAutocomplete", category: "forms", summary: "Selects several known objects from a searchable collection and represents them as removable chips.", fixtureId: "multi-autocomplete.languages", sourcePath: "src/components/forms/controls.tsx", react: ["AutocompleteMultiInput"], types: ["AutocompleteMultiInputProps"], when: "Use when several values may be selected from one long known collection.", avoid: "Use Checkbox lists for a short stable set and TagsInput for simple string labels.", choose: "MultiAutocomplete preserves object identity; TagsInput is optimized for string tags.", behavior: "Base UI manages filtering and selection; named remove buttons update the React Hook Form array.", accessibility: "Each chip has a named removal action and the input exposes combobox state and results.", migration: "Replace Headless UI multi-combobox and supply stable object identifiers." },
  { id: "component.tags-input", route: "/components/forms/tags-input", name: "TagsInput", category: "forms", summary: "Selects a bounded set of known string tags with removable chip semantics.", fixtureId: "tags.entry", sourcePath: "src/components/forms/controls.tsx", react: ["TagsInput"], types: ["TagsInputProps"], when: "Use for several short known labels such as reading-entry categories.", avoid: "Do not use for arbitrary server-backed creation or object values; use MultiAutocomplete instead.", choose: "TagsInput is the string specialization of MultiAutocomplete and does not hide asynchronous data fetching.", behavior: "Typing filters known tags, Enter selects, Backspace and named controls support removal, and limits disable further input.", accessibility: "Announce the combobox results and give every selected tag an explicit remove label.", migration: "Move suggestion data loading to the application and pass a deterministic option set to Paper." },
  { id: "component.button-group", route: "/components/actions/button-group", name: "ButtonGroup", category: "actions", summary: "Keeps related visible actions together without changing link or button semantics.", fixtureId: "button-group.entry-actions", sourcePath: "src/components/feedback/feedback.tsx", react: ["ButtonGroup"], types: ["ButtonGroupProps", "ButtonGroupAction"], when: "Use for a small, stable set of related actions that should remain visible.", avoid: "Use ActionMenu when the set is contextual or too large for the available width.", choose: "ButtonGroup prioritizes visibility; ActionMenu prioritizes compact contextual access.", behavior: "Each entry remains a real anchor or button and the wrapper contributes only a labelled group.", accessibility: "Preserve native roles, express disabled links with aria-disabled plus blocked activation, and label the group.", migration: "Replace responsive Next Link coupling with router-neutral href or onSelect actions." },
  { id: "component.flash", route: "/components/feedback/flash", name: "Flash", category: "feedback", summary: "Places persistent contextual feedback beside the content it describes.", fixtureId: "flash.statuses", sourcePath: "src/components/feedback/feedback.tsx", react: ["Flash"], types: ["FlashProps", "FlashVariant"], when: "Use for page or section feedback that should remain until the context changes, such as a saved result, a contest warning, or a failed page-level operation.", avoid: "Use Toast for brief asynchronous confirmation and an inline field error when the user must repair one value. Do not put required recovery only in a Flash that may disappear with its surrounding content.", choose: "Flash stays beside the content it explains. Toast is transient and global; field errors identify a specific control and its recovery.", variants: "Information covers neutral context, success confirms a completed outcome, and warning calls for review before proceeding. Danger is reserved for an urgent failure and changes the live-region role from status to alert.", states: "Setting visible=false removes the message from the document. Actions are caller-owned links or buttons, so their label, disabled state, and navigation behavior must remain explicit.", behavior: "Danger feedback is announced as an alert; information, success, and warning use a polite status announcement. Changing visible to false unmounts the message instead of merely hiding it.", content: "Write the title as the outcome or issue, then use the body for consequence and recovery. Keep any action label specific—Review fields or Try again instead of OK.", accessibility: "Pair the colored edge with a title and message text, because color does not name the status. Provide actions as real links or buttons and reserve danger for messages that warrant an interrupting alert.", migration: "Map error to danger, remove Next Link ownership, and pass any action as router-owned React content.", pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] },
  { id: "component.loading", route: "/components/feedback/loading", name: "Loading", category: "feedback", summary: "Communicates indeterminate progress with a named, reduced-motion-safe status.", fixtureId: "loading.entries", sourcePath: "src/components/feedback/feedback.tsx", react: ["Loading"], types: ["LoadingProps"], when: "Use while a bounded region is waiting and no meaningful progress value exists.", avoid: "Use native progress for measurable work and retain existing content when optimistic updates are safer.", choose: "Loading indicates indeterminate work; Button loading prevents repeat action while keeping its name.", behavior: "The status label is available to assistive technology and animation stops under reduced motion.", accessibility: "Supply a label that names what is loading and do not rely on spinner motion alone.", migration: "Replace legacy Loading and remove application-owned spinner SVGs." },
  { id: "component.surface", route: "/components/data-display/surface", name: "Surface", category: "data-display", summary: "Composes flat, floating, or showcase content with optional Paper accent rail.", fixtureId: "surface.summary", sourcePath: "src/components/feedback/feedback.tsx", react: ["Surface", "surfaceClassName"], types: ["SurfaceProps"], css: ["paper-surface-card", "paper-elevation-*"], when: "Use to group related content when a semantic article, section, or neutral container also needs Paper surface treatment.", avoid: "Do not wrap every block in a card or use elevation as decoration without hierarchy.", choose: "Flat separates with rules; floating marks transient or interactive content; showcase is reserved for strong demonstrations.", behavior: "Surface adds no interaction and preserves the semantics selected with the as prop.", accessibility: "Choose article or section only when the content merits that landmark; elevation cannot replace headings.", migration: "Replace copied card and shadow classes with the shared recipe or Surface." },
  { id: "component.toast", route: "/components/feedback/toast", name: "Toast", category: "feedback", summary: "Queues brief asynchronous notifications with polite/urgent priority and explicit dismissal.", fixtureId: "toast.entry-saved", sourcePath: "src/components/feedback/feedback.tsx", react: ["ToastProvider", "useToast", "ToastContainer"], types: ["ToastProviderProps"], when: "Use for brief confirmation or failure that is not tied to one visible field or section.", avoid: "Use Flash for persistent contextual feedback and never put required recovery only in an expiring toast.", choose: "Toast is transient global feedback; Flash remains in context until it is no longer relevant.", behavior: "Base UI owns queueing, live announcements, timeout pause, swipe, and stacking behind the Paper provider.", accessibility: "Choose low or high priority intentionally, keep a visible dismiss action, and repeat critical recovery in context.", migration: "Replace react-toastify setup with ToastProvider and useToast; ToastContainer remains a migration alias." },
];

export const phaseThreeFormsFeedbackDocuments = specs.map(document);
