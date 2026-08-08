import { useEffect } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { ActionMenu } from "../components/overlays/ActionMenu";
import { Button, buttonClassName } from "../components/actions/Button";
import { Input } from "../components/forms/Input";
import { Modal } from "../components/overlays/Modal";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  COMPONENT_PAGE_SECTION_KEYS,
  type CatalogDocument,
  type ComponentCategory,
  type ComponentDocumentationSections,
  type RequiredComponentSectionKey,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const PACKAGE_VERSION = "0.1.0";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

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

function InputRecommendedFixture() {
  const methods = useForm<{ title: string }>({
    defaultValues: { title: "August reading log" },
  });
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={methods.handleSubmit(() => undefined)}>
        <Input
          name="title"
          label="Log title"
          hint="Give this reading session a short, recognizable name."
          rules={{ required: "Enter a log title." }}
          required
        />
        <Button type="submit">Save log</Button>
      </form>
    </FormProvider>
  );
}

function InputStatesFixture() {
  const methods = useForm({
    defaultValues: { readonly: "Japanese", disabled: "Archived" },
  });
  return (
    <FormProvider {...methods}>
      <div className="paper-fixture-stack">
        <Input name="readonly" label="Language" readOnly />
        <Input name="disabled" label="Status" disabled />
      </div>
    </FormProvider>
  );
}

function InputErrorFixture() {
  const methods = useForm<{ pages: string }>({ defaultValues: { pages: "" } });
  useEffect(() => {
    methods.setError("pages", { message: "Enter the number of pages read." });
  }, [methods]);
  return (
    <FormProvider {...methods}>
      <Input
        name="pages"
        label="Pages read"
        hint="Use whole pages."
        inputMode="numeric"
        required
      />
    </FormProvider>
  );
}

export const phaseTwoFixtures = [
  defineCatalogFixture({
    id: "button.variants",
    name: "Button variants",
    description: "The complete action hierarchy beside a semantic anchor.",
    tags: ["button", "variants", "anchor"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Button, buttonClassName } from "paper-ui";

<Button>Save log</Button>
<Button variant="outline">Cancel</Button>
<Button variant="ghost">More options</Button>
<Button variant="link">Clear filters</Button>
<Button variant="destructive">Delete log</Button>
<a className={buttonClassName({ variant: "outline" })} href="/logs">
  View logs
</a>`,
    render: () => (
      <div className="paper-fixture-row">
        <Button>Save log</Button>
        <Button variant="outline">Cancel</Button>
        <Button variant="ghost">More options</Button>
        <Button variant="link">Clear filters</Button>
        <Button variant="destructive">Delete log</Button>
        <a className={buttonClassName({ variant: "outline" })} href="#logs">
          View logs
        </a>
      </div>
    ),
  }),
  defineCatalogFixture({
    id: "button.loading",
    name: "Loading action",
    description: "A busy action keeps its visible and accessible name stable.",
    tags: ["button", "loading", "disabled"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Button } from "paper-ui";

<Button loading loadingLabel="Saving reading log">
  Save log
</Button>`,
    render: () => (
      <Button loading loadingLabel="Saving reading log">
        Save log
      </Button>
    ),
  }),
  defineCatalogFixture({
    id: "button.narrow",
    name: "Long, full-width action",
    description: "Long localized content at a narrow viewport.",
    tags: ["button", "full width", "long content", "narrow"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Button } from "paper-ui";

<Button fullWidth>Add this finished book to the August reading log</Button>`,
    render: () => (
      <Button fullWidth>Add this finished book to the August reading log</Button>
    ),
  }),
  defineCatalogFixture({
    id: "input.recommended",
    name: "Reading-log title",
    description: "Label, hint, required rule, value, and submission in one field.",
    tags: ["input", "form", "required", "hint"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Button, Input } from "paper-ui";
import { FormProvider, useForm } from "react-hook-form";

const methods = useForm({ defaultValues: { title: "August reading log" } });
<FormProvider {...methods}>
  <form onSubmit={methods.handleSubmit(saveLog)}>
    <Input
      name="title"
      label="Log title"
      hint="Give this reading session a short, recognizable name."
      rules={{ required: "Enter a log title." }}
      required
    />
    <Button type="submit">Save log</Button>
  </form>
</FormProvider>`,
    render: () => <InputRecommendedFixture />,
  }),
  defineCatalogFixture({
    id: "input.states",
    name: "Read-only and disabled",
    description: "Two non-editable states with distinct native semantics.",
    tags: ["input", "readonly", "disabled"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Input } from "paper-ui";

<Input name="language" label="Language" readOnly />
<Input name="status" label="Status" disabled />`,
    render: () => <InputStatesFixture />,
  }),
  defineCatalogFixture({
    id: "input.error",
    name: "Validation error",
    description: "Hint and error remain associated with the invalid input.",
    tags: ["input", "error", "accessibility"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Input } from "paper-ui";

<Input
  name="pages"
  label="Pages read"
  hint="Use whole pages."
  inputMode="numeric"
  required
/>`,
    render: () => <InputErrorFixture />,
  }),
  defineCatalogFixture({
    id: "modal.recommended",
    name: "Confirm log deletion",
    description: "A focused modal with explanatory copy and two close paths.",
    tags: ["modal", "focus", "destructive"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Modal } from "paper-ui";

<Modal
  triggerLabel="Review deletion"
  triggerVariant="destructive"
  title="Delete this reading log?"
  description="This removes the log from your history."
  closeLabel="Keep log"
  action={{ label: "Delete log", variant: "destructive", onAction: deleteLog }}
>
  <p>August Japanese reading · 1,240 pages</p>
</Modal>`,
    render: () => (
      <Modal
        triggerLabel="Review deletion"
        triggerVariant="destructive"
        title="Delete this reading log?"
        description="This removes the log from your history."
        closeLabel="Keep log"
        action={{
          label: "Delete log",
          variant: "destructive",
          onAction: () => undefined,
        }}
      >
        <p>August Japanese reading · 1,240 pages</p>
      </Modal>
    ),
  }),
  defineCatalogFixture({
    id: "modal.long-content",
    name: "Long modal content",
    description: "Overflow remains inside a viewport-bounded dialog.",
    tags: ["modal", "overflow", "long content"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Modal } from "paper-ui";

<Modal triggerLabel="Review rules" title="August challenge rules">
  <p>Read the complete challenge rules before joining.</p>
</Modal>`,
    render: () => (
      <Modal triggerLabel="Review rules" title="August challenge rules">
        <p>
          Log pages or minutes only after reading. Choose the language you read,
          keep notes concise, and correct accidental duplicates before the contest
          closes. Moderators may review unusual entries to keep the challenge fair.
        </p>
      </Modal>
    ),
  }),
  defineCatalogFixture({
    id: "action-menu.recommended",
    name: "Reading-log actions",
    description: "Common, unavailable, and destructive actions in one menu.",
    tags: ["action menu", "keyboard", "disabled", "destructive"],
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
    code: `import { ActionMenu } from "paper-ui";

<ActionMenu
  label="Log actions"
  items={[
    { id: "edit", label: "Edit log", onSelect: editLog },
    { id: "duplicate", label: "Duplicate log", disabled: true, onSelect: duplicateLog },
    { id: "delete", label: "Delete log", destructive: true, onSelect: deleteLog },
  ]}
/>`,
    render: () => (
      <ActionMenu
        label="Log actions"
        items={[
          { id: "edit", label: "Edit log", onSelect: () => undefined },
          {
            id: "duplicate",
            label: "Duplicate log",
            disabled: true,
            onSelect: () => undefined,
          },
          {
            id: "delete",
            label: "Delete log",
            destructive: true,
            onSelect: () => undefined,
          },
        ]}
      />
    ),
  }),
] as const;

export const phaseTwoDocuments = [
  componentDocument({
    id: "component.button",
    route: "/components/actions/button",
    name: "Button",
    category: "actions",
    summary: "Triggers an immediate action with explicit hierarchy and safe defaults.",
    keywords: ["button", "action", "loading", "anchor", "submit"],
    sourcePath: "src/components/actions/Button/Button.tsx",
    fixtureIds: ["button.variants", "button.loading", "button.narrow"],
    behaviorTestIds: ["button.semantics", "button.recipe-parity", "button.loading"],
    guidance: {
      whenToUse: ["Trigger an immediate operation such as saving or deleting."],
      whenNotToUse: ["Use an anchor for navigation to another resource."],
      content: ["Start with a specific verb and keep labels stable while loading."],
      commonMistakes: ["Do not use the link variant to turn navigation into a button."],
    },
    accessibility: {
      requirements: ["Provide a stable accessible name and visible focus."],
      keyboard: ["Enter and Space activate a button; anchors retain native Enter behavior."],
      knownConstraints: ["Icon-only actions need an explicit aria-label."],
    },
    api: {
      react: ["Button"],
      cssClasses: ["paper-button", "buttonClassName()"],
      publicTypes: ["ButtonProps", "ButtonVariant", "ButtonRecipeOptions"],
      defaults: ["variant=default", "type=button", "loading=false"],
      invalidCombinations: ["Do not put href on Button; style a real anchor with buttonClassName()."],
    },
    migration: {
      legacy: ["ui Button", "btn primary", "btn secondary", "btn danger"],
      notes: ["Map intent: primary to default, secondary to outline, danger to destructive."],
    },
    sections: componentSections("Button", {
      overview: section("Overview", "Button expresses action hierarchy without changing native semantics."),
      whenToUse: section("When to use", "Use Button for an immediate user-initiated operation such as saving a log, applying a filter, or deleting an entry. Use one default button for the primary action in a local task; give supporting actions less emphasis."),
      whenNotToUse: section("When not to use", "Use an anchor when activation navigates to a URL, including links styled with buttonClassName(). Do not use a disabled button as an explanation; keep the reason visible near the unavailable action."),
      choosingBetween: section("Choose between", "Default is the emphasized action, outline is a neutral alternative, ghost is a low-emphasis toolbar action, and link is an action embedded in prose. Destructive communicates irreversible intent; it does not replace a confirmation when the consequence is difficult to undo."),
      anatomy: section("Anatomy", "A control contains an optional icon, a stable text label, and a lower interactive edge."),
      recommendedExample: section("Recommended example", "Save log is specific, short, and defaults to type=button."),
      variants: section("Variants", "Use default for the main action in a task, outline for a visible alternative, ghost for compact secondary actions, link for button behavior that belongs inline with text, and destructive for a destructive operation. buttonClassName() gives a real anchor the same visual hierarchy without changing its navigation semantics."),
      statesAndAdaptation: section("States and adaptation", "loading prevents repeat activation, sets aria-busy, keeps the original accessible name, and announces loadingLabel as a description. fullWidth is intended for constrained layouts; density changes sizing without changing hierarchy."),
      behavior: section("Behavior", "Buttons activate on Enter or Space and default to type=button, preventing accidental form submission. Explicitly set type=submit inside a React Hook Form and submit through methods.handleSubmit(); use methods.reset() only for a deliberate reset workflow."),
      contentGuidance: section("Content guidance", "Start with a specific verb and name the object when context is not obvious: Save log, Delete entry, or Clear filters. Keep the visible label stable while loading and avoid vague labels such as OK or Submit."),
      accessibility: section("Accessibility", "Keep an accessible name, do not convey danger through color alone, and preserve native anchor roles for navigation."),
      implementation: section("Implementation", "Button and buttonClassName() call the same recipe; load paper-ui/styles.css once at the application root."),
      apiReference: section("API reference", "ButtonProps adds variant, loading, loadingLabel, icons, and fullWidth to native button attributes."),
      relatedPatterns: section("Related patterns", "Use Modal for focused confirmation and ActionMenu when several contextual actions compete."),
      migration: section("Migration", "Review old primary, secondary, and danger classes by semantic job instead of renaming mechanically."),
      lifecycle: section("Lifecycle", "Stable in Paper 0.1.0; behavior and variant vocabulary are migration-ready."),
    }),
  }),
  componentDocument({
    id: "component.input",
    route: "/components/forms/input",
    name: "Input",
    category: "forms",
    summary: "Collects one line of text with complete field anatomy and form state.",
    keywords: ["input", "field", "form", "error", "hint"],
    sourcePath: "src/components/forms/Input/Input.tsx",
    fixtureIds: ["input.recommended", "input.states", "input.error"],
    behaviorTestIds: ["input.associations", "input.validation", "input.native-states"],
    guidance: {
      whenToUse: ["Collect a short, free-form value in a React Hook Form."],
      whenNotToUse: ["Use a choice control when the set of valid values is known."],
      content: ["Use a visible noun label and a hint only when it adds constraints or context."],
      commonMistakes: ["Do not use placeholder text as the only label."],
    },
    accessibility: {
      requirements: ["Associate label, hint, and error IDs with the native input."],
      keyboard: ["The native input follows platform text-editing behavior."],
      knownConstraints: ["Input requires a react-hook-form FormProvider."],
    },
    api: {
      react: ["Input"],
      cssClasses: ["paper-field", "paper-input"],
      publicTypes: ["InputProps"],
      defaults: ["Native text input behavior", "Errors come from react-hook-form state"],
      invalidCombinations: ["Do not render Input outside FormProvider."],
    },
    migration: {
      legacy: ["ui/components/Form Input"],
      notes: ["Keep names and validation rules in react-hook-form; replace implicit global field styling."],
    },
    sections: componentSections("Input", {
      overview: section("Overview", "Input owns the complete visible and semantic field relationship."),
      whenToUse: section("When to use", "Use Input for a short free-form value that fits on one line, such as a reading-log title, page count, email address, or password. Set the appropriate native type or inputMode so platform keyboards and validation can help."),
      whenNotToUse: section("When not to use", "Use TextArea for prose and Select, radio controls, or Autocomplete when valid choices are known. Do not use placeholder text as the label: it disappears after entry and does not provide a persistent field name."),
      choosingBetween: section("Choose between", "A read-only value remains focusable and selectable and is still submitted; a disabled value is unavailable and omitted from submission. Use readOnly when users may need to inspect or copy the value, and disabled only when the control does not currently participate in the form."),
      anatomy: section("Anatomy", "The field contains a persistent label, optional hint, native input, and validation message."),
      recommendedExample: section("Recommended example", "Log title demonstrates a useful label, constraint hint, required rule, and explicit submit action."),
      variants: section("Variants", "Text, email, password, numeric-input-mode, and other native types share the field anatomy."),
      statesAndAdaptation: section("States and adaptation", "required adds the native required state and a visible marker. An invalid field sets aria-invalid and connects its alert message alongside any hint. Read-only remains operable for selection; disabled is visually muted and removed from interaction and submission."),
      behavior: section("Behavior", "Registration, value, blur, and validation state come from the nearest FormProvider."),
      contentGuidance: section("Content guidance", "Use a persistent noun phrase for the label. Add a hint only for a format, constraint, or consequence the label cannot carry. An error should explain how to fix the value—Enter the number of pages read is more useful than Invalid input."),
      accessibility: section("Accessibility", "htmlFor, aria-describedby, aria-invalid, and role=alert connect the field anatomy without relying on color."),
      implementation: section("Implementation", "Create methods with useForm(), wrap fields in FormProvider, and submit through methods.handleSubmit()."),
      apiReference: section("API reference", "InputProps requires name and label, accepts hint and register rules, and otherwise follows native input attributes."),
      relatedPatterns: section("Related patterns", "Button submits the surrounding form; future choice controls reuse the same field anatomy."),
      migration: section("Migration", "Move validation into register rules and remove application-owned label/error selectors."),
      lifecycle: section("Lifecycle", "Stable in Paper 0.1.0 with deterministic associations and React Hook Form ownership."),
    }),
  }),
  componentDocument({
    id: "component.modal",
    route: "/components/overlays/modal",
    name: "Modal",
    category: "overlays",
    summary: "Temporarily focuses attention while managing focus, dismissal, and return.",
    keywords: ["modal", "dialog", "overlay", "focus", "confirmation"],
    sourcePath: "src/components/overlays/Modal/Modal.tsx",
    fixtureIds: ["modal.recommended", "modal.long-content"],
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
      recommendedExample: section("Recommended example", "Deletion review states what will be removed before the user commits elsewhere."),
      variants: section("Variants", "Trigger hierarchy can vary; the modal surface and focus behavior stay consistent."),
      statesAndAdaptation: section("States and adaptation", "The viewport scrolls long content and remains bounded at phone, tablet, and desktop widths."),
      behavior: section("Behavior", "Base UI manages opening, focus containment, Escape/outside dismissal, and focus return."),
      contentGuidance: section("Content guidance", "Use a question for confirmation titles and put consequences in the description."),
      accessibility: section("Accessibility", "Title and description label the dialog; two close paths ensure touch-screen-reader escape."),
      implementation: section("Implementation", "Import Modal from paper-ui; Base UI remains a private interaction dependency."),
      apiReference: section("API reference", "ModalProps supports controlled or uncontrolled open state, title, description, content, trigger, close, and optional primary action contracts."),
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
      recommendedExample: section("Recommended example", "Reading-log actions place common editing first, unavailable duplication in place, and deletion last."),
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
