import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
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
  surfaceClassName,
  useToast,
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
  it("submits associated native values", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    render(<NativeControls onSubmit={submit} />);
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
  });

  it("keeps amount and unit as two explicit form fields", () => {
    render(<AmountForm />);
    expect(screen.getByRole("spinbutton", { name: "Progress" })).toHaveValue(12);
    expect(screen.getByRole("combobox", { name: "Unit for progress" })).toHaveValue("pages");
  });
});

describe("Base UI autocomplete controls", () => {
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
