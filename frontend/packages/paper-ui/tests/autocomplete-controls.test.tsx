import { useStore } from '@tanstack/react-form';
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { usePaperForm, Button } from '../src';

import { describe, expect, it, vi } from "vitest";

const LANGUAGES = [
  { id: "ja", label: "Japanese" },
  { id: "zh", label: "Chinese" },
  { id: "ko", label: "Korean" },
] as const;

function AutocompleteForm({ multiple = false, tags = false }: { multiple?: boolean; tags?: boolean }) {
  const methods = usePaperForm({ defaultValues: { language: (multiple ? [] : null) as (typeof LANGUAGES)[number] | (typeof LANGUAGES)[number][] | string[] | null } });
  return (
    <methods.AppForm>
      {tags ? (
        <methods.AppField name="language">{field => <field.TagsInput label="Tags" options={["fiction", "history", "manga"]} placeholder="Add tag" />}</methods.AppField>
      ) : multiple ? (
        <methods.AppField name="language">{field => <field.AutocompleteMultiInput label="Languages" options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} placeholder="Add language" />}</methods.AppField>
      ) : (
        <methods.AppField name="language">{field => <field.AutocompleteInput label="Languages" options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} placeholder="Choose language" />}</methods.AppField>
      )}
      <output data-testid="value">{JSON.stringify(useStore(methods.store, state => state.values["language"]))}</output>
    </methods.AppForm>
  );
}

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
      const methods = usePaperForm({ defaultValues: ({ language: multiple ? [LANGUAGES[0]] : LANGUAGES[0] } as { language: typeof LANGUAGES[number] | typeof LANGUAGES[number][] }), onSubmit: ({ value }) => saved(value) });
      const common = { name: "language", label: "Contest language", options: LANGUAGES,
        format: (value: typeof LANGUAGES[number]) => value.label,
        getId: (value: typeof LANGUAGES[number]) => value.id,
      };
      return <methods.AppForm><methods.Form >
        <methods.AppField name="language" validators={{ onChange: ({ value }) => (multiple ? Array.isArray(value) && value.length === 2 : !Array.isArray(value) && value.id === "ko") ? undefined : "Choose an eligible contest language." }}>{field => multiple ? <field.AutocompleteMultiInput {...common} /> : <field.AutocompleteInput {...common} />}</methods.AppField>
        <Button type="submit">Save registration</Button>
        <Button onClick={() => methods.setFieldValue("language", multiple ? [LANGUAGES[0], LANGUAGES[2]] : LANGUAGES[2])}>Use eligible selection</Button>
      </methods.Form></methods.AppForm>;
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
      const methods = usePaperForm({ defaultValues: { language: null as (typeof LANGUAGES)[number] | null } });
      return <methods.AppForm><methods.AppField name="language" validators={{ onBlur: ({ value }) => !value ? "Choose an option." : undefined }}>{field => <field.AutocompleteInput label="Language" required options={LANGUAGES} format={(option) => option.label} getId={(option) => option.id} />}</methods.AppField><button>Next</button></methods.AppForm>;
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
      const methods = usePaperForm({ defaultValues: { tags: ["fiction"] } });
      return <methods.AppForm><methods.AppField name="tags">{field => <field.TagsInput label="Tags" options={["fiction", "history"]} maxSelections={1} />}</methods.AppField></methods.AppForm>;
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
