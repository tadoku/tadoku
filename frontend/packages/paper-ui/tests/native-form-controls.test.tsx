import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import {
  AmountWithUnit,
  Button,
  Checkbox,
  Input,
  RadioGroup,
  RadioSelect,
  Select,
  TextArea,
} from "../src";

const LANGUAGES = [
  { id: "ja", label: "Japanese" },
  { id: "zh", label: "Chinese" },
  { id: "ko", label: "Korean" },
] as const;

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

function AmountForm() {
  const methods = useForm({ defaultValues: { progressValue: 12, progressUnit: "pages" } });
  return (
    <FormProvider {...methods}>
      <AmountWithUnit name="progress" label="Progress" units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes" }]} />
    </FormProvider>
  );
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
