import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { phaseThreeFormsFeedbackFixtures } from "../src/catalog/phase-three-forms-feedback";
import {
  AmountWithUnit,
  AutocompleteInput,
  AutocompleteMultiInput,
  Button,
  ButtonGroup,
  Checkbox,
  Flash,
  Input,
  Loading,
  RadioGroup,
  RadioSelect,
  Select,
  Surface,
  TagsInput,
  TextArea,
  ToastProvider,
  surfaceClassName,
  useToast,
} from "../src";

const LANGUAGES = [
  { id: "ja", label: "Japanese" },
  { id: "zh", label: "Chinese" },
  { id: "ko", label: "Korean" },
] as const;

describe("form usage examples", () => {
  it.each(["textarea.reading-notes", "select.language", "radio-select.viewport", "radio-group.format", "amount.progress", "autocomplete.language", "multi-autocomplete.languages", "tags.entry"])("demonstrates required validation in %s.empty", async (id) => {
    const fixture = phaseThreeFormsFeedbackFixtures.find((entry) => entry.id === `${id}.empty`);
    const user = userEvent.setup();
    render(<>{fixture!.render()}</>);
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("alert")).not.toBeEmptyDOMElement();
    expect(screen.queryByText("Entry saved")).not.toBeInTheDocument();
  });

  it("saves the submitted values and resets the example", async () => {
    const fixture = phaseThreeFormsFeedbackFixtures.find((entry) => entry.id === "textarea.reading-notes.empty");
    const user = userEvent.setup();
    render(<>{fixture!.render()}</>);
    const input = screen.getByRole("textbox", { name: /Reading notes/u });
    await user.type(input, "Finished chapter three.");
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("status")).toHaveTextContent("Finished chapter three.");
    await user.clear(input);
    expect(screen.getByRole("status")).toHaveTextContent("Finished chapter three.");
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(input).toHaveValue("");
  });
});

function NativeControls({ onSubmit }: { onSubmit: (values: unknown) => void }) {
  const methods = useForm({ defaultValues: { notes: "", language: "", public: false, pace: "", format: "" } });
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={methods.handleSubmit(onSubmit)}>
        <TextArea name="notes" label="Notes" hint="Keep spoilers out." required />
        <Select name="language" label="Language" options={LANGUAGES.map(({ id, label }) => ({ value: id, label }))} placeholder="Choose language" required />
        <Checkbox name="public" label="Show on my profile" />
        <RadioSelect name="pace" label="Reading pace" options={[{ value: "pages", label: "Pages" }, { value: "minutes", label: "Minutes" }]} required />
        <RadioGroup name="format" label="Format" options={[{ value: "book", label: "Book", description: "Printed or digital book" }, { value: "audio", label: "Audio", description: "Narrated content" }]} />
        <Button type="submit">Save entry</Button>
      </form>
    </FormProvider>
  );
}

function SegmentedRadioSelect({
  defaultViewport = "tablet",
  onSubmit = () => undefined,
}: {
  defaultViewport?: string;
  onSubmit?: (values: unknown) => void;
}) {
  const methods = useForm({ defaultValues: { viewport: defaultViewport } });
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={methods.handleSubmit(onSubmit)}>
        <RadioSelect
          name="viewport"
          label="Preview size"
          hint="Choose one viewport for the isolated preview."
          variant="segmented"
          required
          options={[
            { value: "phone", label: "Phone" },
            { value: "tablet", label: "Tablet" },
            { value: "desktop", label: "Desktop" },
            { value: "wide", label: "Wide", disabled: true },
          ]}
        />
        <Button type="submit">Apply viewport</Button>
      </form>
    </FormProvider>
  );
}

function AutocompleteForm({ multiple = false, tags = false }: { multiple?: boolean; tags?: boolean }) {
  const methods = useForm({ defaultValues: { language: multiple ? [] : null } });
  return (
    <FormProvider {...methods}>
      {tags ? (
        <TagsInput name="language" label="Tags" options={["fiction", "history", "manga"]} placeholder="Add tag" />
      ) : multiple ? (
        <AutocompleteMultiInput
          name="language"
          label="Languages"
          options={LANGUAGES}
          format={(option) => option.label}
          getId={(option) => option.id}
          placeholder="Add language"
        />
      ) : (
        <AutocompleteInput
          name="language"
          label="Languages"
          options={LANGUAGES}
          format={(option) => option.label}
          getId={(option) => option.id}
          placeholder="Choose language"
        />
      )}
      <output data-testid="value">{JSON.stringify(methods.watch("language"))}</output>
    </FormProvider>
  );
}

function AmountForm() {
  const methods = useForm({ defaultValues: { progressValue: 12, progressUnit: "pages" } });
  return (
    <FormProvider {...methods}>
      <AmountWithUnit name="progress" label="Progress" units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes" }]} />
    </FormProvider>
  );
}

function ToastTrigger() {
  const toast = useToast();
  return <Button onClick={() => toast.add({ title: "Entry saved", description: "12 pages added." })}>Show notification</Button>;
}

describe("native React Hook Form controls", () => {
  it("places helper text after controls so optional hints do not offset a row of fields", () => {
    function HintedFields() {
      const methods = useForm({ defaultValues: { title: "", date: "2026-08-31", notes: "", language: "", public: false, progressValue: 1, progressUnit: "pages", pace: "", format: "" } });
      return (
        <FormProvider {...methods}>
          <Input name="title" label="Title" />
          <Input name="date" label="Date" type="date" hint="The date you read." />
          <TextArea name="notes" label="Notes" hint="No spoilers." />
          <Select name="language" label="Language" hint="Reading language." options={[]} />
          <Checkbox name="public" label="Public entry" hint="Shown on your profile." />
          <AmountWithUnit name="progress" label="Progress" hint="Completed pages." units={[{ value: "pages", label: "pages" }]} />
          <RadioSelect name="pace" label="Pace" hint="Choose a pace." options={[{ value: "pages", label: "Pages" }]} />
          <RadioGroup name="format" label="Format" hint="Choose a format." options={[{ value: "book", label: "Book", description: "Printed or digital" }]} />
        </FormProvider>
      );
    }
    render(<HintedFields />);
    for (const label of ["Date", "Notes", "Language", "Public entry", "Progress"]) {
      const control = screen.getByLabelText(label, { selector: "input, textarea, select" });
      const hint = document.getElementById(control.getAttribute("aria-describedby")!);
      expect(hint).not.toBeNull();
      expect(control.compareDocumentPosition(hint!) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
      expect(control).toHaveAccessibleDescription(hint!.textContent!);
    }
    for (const name of ["Pace", "Format"]) {
      const group = screen.getByRole("group", { name });
      const hint = document.getElementById(group.getAttribute("aria-describedby")!);
      const radio = within(group).getByRole("radio");
      expect(radio.compareDocumentPosition(hint!) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
      expect(group).toHaveAccessibleDescription(hint!.textContent!);
    }
  });

  it("preserves caller descriptions alongside built-in hints", () => {
    function DescribedFields() {
      const methods = useForm({ defaultValues: { title: "", notes: "", language: "", public: false, progressValue: 1, progressUnit: "pages" } });
      return <FormProvider {...methods}><p id="visibility-policy">Visible to contest moderators.</p><Input name="title" label="Title" hint="A short title." aria-describedby="visibility-policy" /><TextArea name="notes" label="Notes" hint="No spoilers." aria-describedby="visibility-policy" /><Select name="language" label="Language" hint="Reading language." options={[]} aria-describedby="visibility-policy" /><Checkbox name="public" label="Public entry" hint="Shown on your profile." aria-describedby="visibility-policy" /><AmountWithUnit name="progress" label="Progress" hint="Completed pages." units={[{ value: "pages", label: "pages" }]} aria-describedby="visibility-policy" /></FormProvider>;
    }
    render(<DescribedFields />);
    for (const label of ["Title", "Notes", "Language", "Public entry"]) {
      expect(screen.getByLabelText(label)).toHaveAccessibleDescription(/Visible to contest moderators\./u);
    }
    expect(screen.getByLabelText("Notes")).toHaveAccessibleDescription("No spoilers. Visible to contest moderators.");
    expect(screen.getByRole("spinbutton", { name: "Progress" })).toHaveAccessibleDescription("Completed pages. Visible to contest moderators.");
  });

  it("omits disabled input values while retaining read-only values", async () => {
    const submit = vi.fn();
    function DisabledFields() {
      const methods = useForm({ defaultValues: { disabledTitle: "Unavailable", readonlyTitle: "Reading log", notes: "Hidden note", language: "ja", public: true, viewport: "phone" } });
      return <FormProvider {...methods}><form noValidate onSubmit={methods.handleSubmit(submit)}><Input name="disabledTitle" label="Unavailable title" disabled required /><Input name="readonlyTitle" label="Read-only title" readOnly /><TextArea name="notes" label="Notes" disabled /><Select name="language" label="Language" options={[{ value: "ja", label: "Japanese" }]} disabled /><Checkbox name="public" label="Public" disabled /><RadioSelect name="viewport" label="Viewport" options={[{ value: "phone", label: "Phone" }]} disabled /><Button type="submit">Save</Button></form></FormProvider>;
    }
    const user = userEvent.setup();
    render(<DisabledFields />);
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(submit).toHaveBeenCalledWith({ disabledTitle: undefined, readonlyTitle: "Reading log", notes: undefined, language: undefined, public: undefined, viewport: undefined }, expect.anything());
  });

  it("submits associated native values", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    render(<NativeControls onSubmit={submit} />);
    expect(screen.getByRole("radio", { name: /Book/u })).toBeRequired();
    await user.type(screen.getByRole("textbox", { name: "Notes" }), "Finished chapter two");
    await user.selectOptions(screen.getByRole("combobox", { name: "Language" }), "ja");
    await user.click(screen.getByRole("checkbox", { name: "Show on my profile" }));
    await user.click(screen.getByRole("radio", { name: "Pages" }));
    await user.click(screen.getByRole("radio", { name: /Book/u }));
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(submit).toHaveBeenCalledWith(expect.objectContaining({
      notes: "Finished chapter two",
      language: "ja",
      public: true,
      pace: "pages",
      format: "book",
    }), expect.anything());
  });

  it("reports validation through field relationships", async () => {
    const user = userEvent.setup();
    render(<NativeControls onSubmit={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    const notes = screen.getByRole("textbox", { name: "Notes" });
    expect(notes).toHaveAttribute("aria-invalid", "true");
    expect(notes).toHaveAccessibleDescription("Keep spoilers out. This field is required.");

    const format = screen.getByRole("group", { name: /Format/u });
    expect(format).toHaveAttribute("aria-invalid", "true");
    expect(format).toHaveAccessibleDescription("Choose an option.");
    for (const radio of within(format).getAllByRole("radio")) {
      expect(radio).toHaveAttribute("aria-invalid", "true");
    }
  });

  it("keeps amount and unit as two explicit form fields", () => {
    render(<AmountForm />);
    expect(screen.getByRole("spinbutton", { name: "Progress" })).toHaveValue(12);
    expect(screen.getByRole("combobox", { name: "Unit for progress" })).toHaveValue("pages");
  });

  it("validates compound amount bounds and disables the complete compound control", async () => {
    function ConstrainedAmount({ disabled = false }: { disabled?: boolean }) {
      const methods = useForm({ defaultValues: { progressValue: 0, progressUnit: "pages" } });
      return <FormProvider {...methods}><form noValidate onSubmit={methods.handleSubmit(() => undefined)}><AmountWithUnit name="progress" label="Progress" min={1} required disabled={disabled} units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes", disabled: true }]} /><Button type="submit">Save</Button></form></FormProvider>;
    }
    const user = userEvent.setup();
    const { rerender } = render(<ConstrainedAmount />);
    expect(screen.getByRole("option", { name: "minutes" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Enter at least 1.");
    rerender(<ConstrainedAmount disabled />);
    expect(screen.getByRole("spinbutton", { name: /Progress/u })).toBeDisabled();
    expect(screen.getByRole("combobox", { name: "Unit for progress" })).toBeDisabled();
  });

  it("renders segmented RadioSelect as one native exclusive choice", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    const { container } = render(<SegmentedRadioSelect onSubmit={submit} />);
    const group = screen.getByRole("group", { name: /Preview size/u });
    const radios = within(group).getAllByRole("radio");

    expect(group).toHaveClass("paper-radio-select--segmented");
    expect(radios).toHaveLength(4);
    expect(screen.getByRole("radio", { name: "Tablet" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "Wide" })).toBeDisabled();
    expect(container.querySelector("button[aria-pressed]")).toBeNull();
    expect(group.querySelector(".paper-button")).toBeNull();

    await user.click(screen.getByRole("radio", { name: "Phone" }));
    expect(screen.getByRole("radio", { name: "Phone" })).toBeChecked();
    expect(screen.getByRole("radio", { name: "Tablet" })).not.toBeChecked();
    await user.click(screen.getByRole("button", { name: "Apply viewport" }));
    expect(submit).toHaveBeenCalledWith(
      expect.objectContaining({ viewport: "phone" }),
      expect.anything(),
    );
  });

  it("reports a missing required segmented choice on its native group and radios", async () => {
    const user = userEvent.setup();
    render(<SegmentedRadioSelect defaultViewport="" />);

    await user.click(screen.getByRole("button", { name: "Apply viewport" }));

    const group = screen.getByRole("group", { name: /Preview size/u });
    expect(group).toHaveAttribute("aria-invalid", "true");
    expect(group).toHaveAccessibleDescription(
      "Choose one viewport for the isolated preview. Choose an option.",
    );
    for (const radio of within(group).getAllByRole("radio")) {
      expect(radio).toHaveAttribute("aria-invalid", "true");
    }
  });
});

describe("Base UI autocomplete controls", () => {
  it.each([false, true])("preserves a %s multi-selection when Escape dismisses and is pressed again after closing", async (multiple) => {
    const user = userEvent.setup();
    render(<AutocompleteForm multiple={multiple} />);
    const input = screen.getByRole("combobox", { name: "Languages" });
    await user.type(input, "Japanese");
    await user.click(await screen.findByRole("option", { name: "Japanese" }));
    expect(screen.getByTestId("value")).toHaveTextContent('"id":"ja"');
    input.focus();
    await user.keyboard("{Escape}{Escape}");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(screen.getByTestId("value")).toHaveTextContent('"id":"ja"');
    if (multiple) expect(screen.getByRole("button", { name: "Remove Japanese" })).toBeVisible();
    else expect(input).toHaveValue("Japanese");
  });

  it.each([false, true])("runs consumer validation for a %s multi-selection field and exposes its recovery message", async (multiple) => {
    const saved = vi.fn();
    function ValidatedAutocomplete() {
      const methods = useForm<{ language: typeof LANGUAGES[number] | typeof LANGUAGES[number][] }>({ defaultValues: { language: multiple ? [LANGUAGES[0]] : LANGUAGES[0] } });
      const common = { name: "language", label: "Contest language", options: LANGUAGES,
        format: (value: typeof LANGUAGES[number]) => value.label,
        getId: (value: typeof LANGUAGES[number]) => value.id,
        rules: { validate: (value: unknown) => (multiple ? Array.isArray(value) && value.length === 2 : (value as typeof LANGUAGES[number])?.id === "ko") || "Choose an eligible contest language." },
      };
      return <FormProvider {...methods}><form onSubmit={methods.handleSubmit(saved)}>
        {multiple ? <AutocompleteMultiInput {...common} /> : <AutocompleteInput {...common} />}
        <Button type="submit">Save registration</Button>
        <Button onClick={() => methods.setValue("language", multiple ? [LANGUAGES[0], LANGUAGES[2]] : LANGUAGES[2], { shouldValidate: true })}>Use eligible selection</Button>
      </form></FormProvider>;
    }
    const user = userEvent.setup();
    render(<ValidatedAutocomplete />);
    await user.click(screen.getByRole("button", { name: "Save registration" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Choose an eligible contest language.");
    expect(screen.getByRole("combobox", { name: "Contest language" })).toHaveAccessibleDescription("Choose an eligible contest language.");
    expect(saved).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Use eligible selection" }));
    await user.click(screen.getByRole("button", { name: "Save registration" }));
    expect(saved).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("validates required selections on blur and exposes the error relationship", async () => {
    function RequiredAutocomplete() {
      const methods = useForm({ mode: "onBlur", defaultValues: { language: null } });
      return <FormProvider {...methods}><AutocompleteInput name="language" label="Language" required options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} /><button>Next</button></FormProvider>;
    }
    const user = userEvent.setup();
    render(<RequiredAutocomplete />);
    const input = screen.getByRole("combobox", { name: /Language/u });
    await user.click(input);
    await user.keyboard("{Escape}{Tab}");
    expect(await screen.findByRole("alert")).toHaveTextContent("Choose an option.");
    expect(input).toHaveAttribute("aria-required", "true");
    expect(input).toHaveAccessibleDescription("Choose an option.");
  });

  it("prevents opening and adding beyond the tag limit while keeping removal available", async () => {
    function LimitedTags() {
      const methods = useForm({ defaultValues: { tags: ["fiction"] } });
      return <FormProvider {...methods}><TagsInput name="tags" label="Tags" options={["fiction", "history"]} maxSelections={1} /></FormProvider>;
    }
    const user = userEvent.setup();
    render(<LimitedTags />);
    expect(screen.getByRole("button", { name: "Show tags options" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Remove fiction" }));
    expect(screen.getByRole("combobox", { name: "Tags" })).not.toBeDisabled();
    expect(screen.getByRole("button", { name: "Show tags options" })).not.toBeDisabled();
  });

  it("uses Paper icons instead of text glyphs for combobox actions", () => {
    render(<AutocompleteForm />);
    const trigger = screen.getByRole("button", { name: "Show languages options" });
    expect(trigger.querySelector("svg")).not.toBeNull();
    expect(trigger).not.toHaveTextContent("⌄");
  });

  it("filters and selects one option from the keyboard", async () => {
    const user = userEvent.setup();
    render(<AutocompleteForm />);
    const input = screen.getByRole("combobox", { name: "Languages" });
    await user.click(input);
    await user.type(input, "jap");
    await user.keyboard("{ArrowDown}{Enter}");
    expect(screen.getByTestId("value")).toHaveTextContent('"id":"ja"');
  });

  it("selects and removes multiple values with named chip actions", async () => {
    const user = userEvent.setup();
    render(<AutocompleteForm multiple />);
    const input = screen.getByRole("combobox", { name: "Languages" });
    await user.click(input);
    await user.type(input, "jap");
    await user.keyboard("{ArrowDown}{Enter}");
    const remove = screen.getByRole("button", { name: "Remove Japanese" });
    expect(remove.querySelector("svg")).not.toBeNull();
    expect(remove).not.toHaveTextContent("×");
    await user.click(remove);
    expect(screen.getByTestId("value")).toHaveTextContent("[]");
  });

  it("reuses multi-autocomplete semantics for tags", async () => {
    const user = userEvent.setup();
    render(<AutocompleteForm tags />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    await user.click(input);
    await user.type(input, "man");
    await user.keyboard("{ArrowDown}{Enter}");
    expect(screen.getByRole("button", { name: "Remove manga" })).toBeInTheDocument();
  });

  it("keeps its popup in the input owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(<AutocompleteForm />, {
      container,
      baseElement: frameDocument.body,
    });

    await user.click(within(frameDocument.body).getByRole("combobox", { name: "Languages" }));
    expect(await within(frameDocument.body).findByRole("listbox")).toBeInTheDocument();
    expect(document.body.querySelector('[role="listbox"]')).toBeNull();
    unmount();
    frame.remove();
  });
});

describe("feedback and action compositions", () => {
  it("uses urgent and polite live-region semantics", () => {
    render(<><Flash title="Saved">Entry updated.</Flash><Flash variant="danger">Could not delete entry.</Flash><Loading label="Loading entries" /></>);
    expect(screen.getAllByRole("status")).toHaveLength(2);
    expect(screen.getByRole("alert")).toHaveTextContent("Could not delete entry.");
    expect(screen.getByText("Loading entries")).toHaveClass("paper-visually-hidden");
  });

  it("preserves links and buttons inside one labelled ButtonGroup", async () => {
    const user = userEvent.setup();
    const selected = vi.fn();
    render(<ButtonGroup label="Log actions" actions={[{ id: "view", label: "View", href: "/logs/1" }, { id: "delete", label: "Delete", variant: "destructive", onSelect: selected }]} />);
    expect(screen.getByRole("group", { name: "Log actions" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "View" })).toHaveAttribute("href", "/logs/1");
    await user.click(screen.getByRole("button", { name: "Delete" }));
    expect(selected).toHaveBeenCalledOnce();
  });

  it("shares surface recipes and renders queued Base UI toasts", async () => {
    const user = userEvent.setup();
    render(<ToastProvider timeout={0}><Surface elevation="floating" accent>Reading summary</Surface><ToastTrigger /></ToastProvider>);
    expect(screen.getByText("Reading summary")).toHaveClass(surfaceClassName({ elevation: "floating", accent: true }));
    await user.click(screen.getByRole("button", { name: "Show notification" }));
    expect(await screen.findByText("Entry saved")).toBeInTheDocument();
    const dismiss = screen.getByRole("button", { name: "Dismiss notification" });
    expect(dismiss.querySelector("svg")).not.toBeNull();
    expect(dismiss).not.toHaveTextContent("×");
  });

  it("keeps its toast viewport in the provider owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <ToastProvider timeout={0}><ToastTrigger /></ToastProvider>,
      { container, baseElement: frameDocument.body },
    );

    await user.click(within(frameDocument.body).getByRole("button", { name: "Show notification" }));
    expect(await within(frameDocument.body).findByText("Entry saved")).toBeInTheDocument();
    expect(document.body).not.toHaveTextContent("Entry saved");
    unmount();
    frame.remove();
  });
});
