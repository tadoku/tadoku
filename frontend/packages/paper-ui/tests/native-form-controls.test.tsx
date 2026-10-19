import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { usePaperForm, Button } from '../src';

import { describe, expect, it, vi } from "vitest";

const LANGUAGES = [
  { id: "ja", label: "Japanese" },
  { id: "zh", label: "Chinese" },
  { id: "ko", label: "Korean" },
] as const;

function NativeControls({ onSubmit }: { onSubmit: (values: unknown) => void }) {
  const methods = usePaperForm({ defaultValues: { notes: "", language: "", public: false, pace: "", format: "" }, onSubmit: ({ value }) => onSubmit(value) });
  return (
    <methods.AppForm>
      <methods.Form>
        <methods.AppField name="notes" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; } }}>{field => <field.TextArea label="Notes" hint="Keep spoilers out." required />}</methods.AppField>
        <methods.AppField name="language" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.Select label="Language" options={LANGUAGES.map(({ id, label }) => ({ value: id, label }))} placeholder="Choose language" required />}</methods.AppField>
        <methods.AppField name="public">{field => <field.Checkbox label="Show on my profile" />}</methods.AppField>
        <methods.AppField name="pace" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.RadioSelect label="Reading pace" options={[{ value: "pages", label: "Pages" }, { value: "minutes", label: "Minutes" }]} required />}</methods.AppField>
        <methods.AppField name="format" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.RadioGroup label="Format" options={[{ value: "book", label: "Book", description: "Printed or digital book" }, { value: "audio", label: "Audio", description: "Narrated content" }]} />}</methods.AppField>
        <Button type="submit">Save entry</Button>
      </methods.Form>
    </methods.AppForm>
  );
}

function SegmentedRadioSelect({
  defaultViewport = "tablet",
  onSubmit = () => undefined,
}: {
  defaultViewport?: string;
  onSubmit?: (values: unknown) => void;
}) {
  const methods = usePaperForm({ defaultValues: { viewport: defaultViewport }, onSubmit: ({ value }) => onSubmit(value) });
  return (
    <methods.AppForm>
      <methods.Form>
        <methods.AppField name="viewport" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.RadioSelect label="Preview size" hint="Choose one viewport for the isolated preview." variant="segmented" required options={[
            { value: "phone", label: "Phone" },
            { value: "tablet", label: "Tablet" },
            { value: "desktop", label: "Desktop" },
            { value: "wide", label: "Wide", disabled: true },
          ]} />}</methods.AppField>
        <Button type="submit">Apply viewport</Button>
      </methods.Form>
    </methods.AppForm>
  );
}

function AmountForm() {
  const methods = usePaperForm({ defaultValues: { progressValue: 12, progressUnit: "pages" } });
  return (
    <methods.AppForm>
      <methods.AppField name="progressUnit">{unitField => <methods.AppField name="progressValue" >{field => <field.AmountWithUnit label="Progress" units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes" }]} unitField={unitField} />}</methods.AppField>}</methods.AppField>
    </methods.AppForm>
  );
}

describe("native TanStack Form controls", () => {
  it("keeps numeric input values numeric when edited without a library-specific coercion rule", async () => {
    const submit = vi.fn();
    function NumericEntry() {
      const methods = usePaperForm({ defaultValues: { pages: 12 } , onSubmit: ({ value }) => submit(value) });
      return <methods.AppForm><methods.Form><methods.AppField name="pages">{field => <field.Input label="Pages" type="number" />}</methods.AppField><Button type="submit">Save pages</Button></methods.Form></methods.AppForm>;
    }
    const user = userEvent.setup();
    render(<NumericEntry />);
    await user.clear(screen.getByRole("spinbutton", { name: "Pages" }));
    await user.type(screen.getByRole("spinbutton", { name: "Pages" }), "24");
    await user.click(screen.getByRole("button", { name: "Save pages" }));
    expect(submit.mock.calls[0]?.[0]).toEqual({ pages: 24 });
  });
  it("places helper text after controls so optional hints do not offset a row of fields", () => {
    function HintedFields() {
      const methods = usePaperForm({ defaultValues: { title: "", date: "2026-08-31", notes: "", language: "", public: false, progressValue: 1, progressUnit: "pages", pace: "", format: "" } });
      return (
        <methods.AppForm>
          <methods.AppField name="title">{field => <field.Input label="Title" />}</methods.AppField>
          <methods.AppField name="date">{field => <field.Input label="Date" type="date" hint="The date you read." />}</methods.AppField>
          <methods.AppField name="notes">{field => <field.TextArea label="Notes" hint="No spoilers." />}</methods.AppField>
          <methods.AppField name="language">{field => <field.Select label="Language" hint="Reading language." options={[]} />}</methods.AppField>
          <methods.AppField name="public">{field => <field.Checkbox label="Public entry" hint="Shown on your profile." />}</methods.AppField>
          <methods.AppField name="progressUnit">{unitField => <methods.AppField name="progressValue" >{field => <field.AmountWithUnit label="Progress" hint="Completed pages." units={[{ value: "pages", label: "pages" }]} unitField={unitField} />}</methods.AppField>}</methods.AppField>
          <methods.AppField name="pace">{field => <field.RadioSelect label="Pace" hint="Choose a pace." options={[{ value: "pages", label: "Pages" }]} />}</methods.AppField>
          <methods.AppField name="format" validators={{ onChange: ({ value }) => { if (!value) return "Choose an option."; } }}>{field => <field.RadioGroup label="Format" hint="Choose a format." options={[{ value: "book", label: "Book", description: "Printed or digital" }]} />}</methods.AppField>
        </methods.AppForm>
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
      const methods = usePaperForm({ defaultValues: { title: "", notes: "", language: "", public: false, progressValue: 1, progressUnit: "pages" } });
      return <methods.AppForm><p id="visibility-policy">Visible to contest moderators.</p><methods.AppField name="title">{field => <field.Input label="Title" hint="A short title." aria-describedby="visibility-policy" />}</methods.AppField><methods.AppField name="notes">{field => <field.TextArea label="Notes" hint="No spoilers." aria-describedby="visibility-policy" />}</methods.AppField><methods.AppField name="language">{field => <field.Select label="Language" hint="Reading language." options={[]} aria-describedby="visibility-policy" />}</methods.AppField><methods.AppField name="public">{field => <field.Checkbox label="Public entry" hint="Shown on your profile." aria-describedby="visibility-policy" />}</methods.AppField><methods.AppField name="progressUnit">{unitField => <methods.AppField name="progressValue" >{field => <field.AmountWithUnit label="Progress" hint="Completed pages." units={[{ value: "pages", label: "pages" }]} aria-describedby="visibility-policy" unitField={unitField} />}</methods.AppField>}</methods.AppField></methods.AppForm>;
    }
    render(<DescribedFields />);
    for (const label of ["Title", "Notes", "Language", "Public entry"]) {
      expect(screen.getByLabelText(label)).toHaveAccessibleDescription(/Visible to contest moderators\./u);
    }
    expect(screen.getByLabelText("Notes")).toHaveAccessibleDescription("No spoilers. Visible to contest moderators.");
    expect(screen.getByRole("spinbutton", { name: "Progress" })).toHaveAccessibleDescription("Completed pages. Visible to contest moderators.");
  });

  it("retains form state for disabled and read-only controls", async () => {
    const submit = vi.fn();
    function DisabledFields() {
      const methods = usePaperForm({ defaultValues: { disabledTitle: "Unavailable", readonlyTitle: "Reading log", notes: "Hidden note", language: "ja", public: true, viewport: "phone" }, onSubmit: ({ value }) => submit(value) });
      return <methods.AppForm><methods.Form><methods.AppField name="disabledTitle">{field => <field.Input label="Unavailable title" disabled required />}</methods.AppField><methods.AppField name="readonlyTitle">{field => <field.Input label="Read-only title" readOnly />}</methods.AppField><methods.AppField name="notes">{field => <field.TextArea label="Notes" disabled />}</methods.AppField><methods.AppField name="language">{field => <field.Select label="Language" options={[{ value: "ja", label: "Japanese" }]} disabled />}</methods.AppField><methods.AppField name="public">{field => <field.Checkbox label="Public" disabled />}</methods.AppField><methods.AppField name="viewport">{field => <field.RadioSelect label="Viewport" options={[{ value: "phone", label: "Phone" }]} disabled />}</methods.AppField><Button type="submit">Save</Button></methods.Form></methods.AppForm>;
    }
    const user = userEvent.setup();
    render(<DisabledFields />);
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(submit).toHaveBeenCalledWith({ disabledTitle: "Unavailable", readonlyTitle: "Reading log", notes: "Hidden note", language: "ja", public: true, viewport: "phone" });
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
    }));
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
      const methods = usePaperForm({ defaultValues: { progressValue: 0, progressUnit: "pages" }, onSubmit: () => undefined });
      return <methods.AppForm><methods.Form><methods.AppField name="progressUnit">{unitField => <methods.AppField name="progressValue" validators={(disabled) ? {} : { onChange: ({ value }) => { if (String(value) === '' || !Number.isFinite(Number(value))) return "Enter an amount."; if (Number(value) < 1) return "Enter at least 1."; } }}>{field => <field.AmountWithUnit label="Progress" min={1} required disabled={disabled} units={[{ value: "pages", label: "pages" }, { value: "minutes", label: "minutes", disabled: true }]} unitField={unitField} />}</methods.AppField>}</methods.AppField><Button type="submit">Save</Button></methods.Form></methods.AppForm>;
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
      expect.objectContaining({ viewport: "phone" })
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
