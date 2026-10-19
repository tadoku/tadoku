import { useStore } from '@tanstack/react-form';
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { usePaperForm, Button, type TagsInputProps } from '../src';

import { describe, expect, it, vi } from "vitest";

import { phaseThreeFormsFeedbackFixtures } from "../src/catalog/phase-three-forms-feedback";

function TagsForm({
  initialTags = [],
  onSubmit = () => undefined,
  validate,
  ...props
}: Partial<TagsInputProps> & { initialTags?: string[]; validate?: (tags: string[]) => true | string; onSubmit?: (values: unknown) => void }) {
  const methods = usePaperForm({ defaultValues: { tags: initialTags }, onSubmit: ({ value }) => onSubmit(value) });
  return <methods.AppForm>
    <methods.Form>
      <methods.AppField name="tags" validators={{ onChange: ({ value }) => { if (props.required && !value.length) return "Add at least one tag."; const result = validate?.(value); return result === true ? undefined : result; } }}>{field => <field.TagsInput label="Tags" options={["fiction", "history", "manga"]} {...props} />}</methods.AppField>
      <output data-testid="tags">{JSON.stringify(useStore(methods.store, state => state.values["tags"]))}</output>
      <Button type="submit">Save entry</Button>
      <Button onClick={() => methods.reset()}>Reset</Button>
    </methods.Form>
  </methods.AppForm>;
}

describe("TagsInput creation and selection", () => {
  it("demonstrates saving a custom tag in the published styleguide example", async () => {
    const user = userEvent.setup();
    const example = phaseThreeFormsFeedbackFixtures.find(fixture => fixture.id === "tags.entry")!;
    render(<>{example.render()}</>);
    await user.type(screen.getByRole("combobox", { name: /Tags/ }), "book club{Enter}");
    expect(screen.queryByText("Entry saved")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("status")).toHaveTextContent('"book club"');
  });

  it("removes selected tags from suggestions and returns focus after selecting and removing", async () => {
    const user = userEvent.setup();
    render(<TagsForm initialTags={["fiction", "book club"]} />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    expect(screen.getByRole("button", { name: "Remove book club" })).toBeVisible();
    await user.click(input);
    expect(await screen.findByRole("option", { name: "history" })).toBeVisible();
    expect(screen.queryByRole("option", { name: "fiction" })).not.toBeInTheDocument();
    await user.type(input, "hist");
    await user.click(await screen.findByRole("option", { name: "history" }));
    expect(screen.getByTestId("tags")).toHaveTextContent('["fiction","book club","history"]');
    expect(input).toHaveValue("");
    expect(input).toHaveFocus();
    expect(screen.queryByRole("option", { name: "history" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Remove history" }));
    expect(input).toHaveFocus();
    await user.click(input);
    expect(await screen.findByRole("option", { name: "history" })).toBeVisible();
  });

  it("creates trimmed arbitrary tags on Enter without submitting, including empty and duplicate Enter", async () => {
    const user = userEvent.setup();
    const saved = vi.fn();
    render(<TagsForm onSubmit={saved} />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    await user.type(input, "  book club  {Enter}");
    expect(screen.getByTestId("tags")).toHaveTextContent('["book club"]');
    expect(input).toHaveFocus();
    expect(input).toHaveValue("");
    await user.keyboard("{Enter}");
    await user.type(input, " BOOK CLUB {Enter}");
    expect(screen.getByTestId("tags")).toHaveTextContent('["book club"]');
    expect(saved).not.toHaveBeenCalled();
    await user.clear(input);
    await user.keyboard("{Escape}{Enter}");
    expect(saved).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    await waitFor(() => expect(saved).toHaveBeenCalledTimes(1));
    expect(saved.mock.calls[0][0]).toEqual({ tags: ["book club"] });
  });

  it("offers a visible Add suggestion, and keeps highlighted known options selectable by keyboard", async () => {
    const user = userEvent.setup();
    render(<TagsForm />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    await user.type(input, "週末読書");
    await user.click(await screen.findByRole("option", { name: "Add “週末読書”" }));
    expect(screen.getByRole("button", { name: "Remove 週末読書" })).toBeVisible();
    expect(input).toHaveFocus();
    expect(input).toHaveValue("");
    await user.type(input, "hist");
    await user.keyboard("{ArrowDown}{Enter}");
    expect(screen.getByRole("button", { name: "Remove history" })).toBeVisible();
    expect(screen.queryByRole("button", { name: "Remove hist" })).not.toBeInTheDocument();
  });

  it("does not select or create while an IME composition is being confirmed", async () => {
    const user = userEvent.setup();
    const saved = vi.fn();
    render(<TagsForm onSubmit={saved} />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    await user.type(input, "hist");
    await user.keyboard("{ArrowDown}");
    fireEvent.compositionStart(input);
    fireEvent.keyDown(input, { key: "Enter", code: "Enter", isComposing: true });
    expect(screen.getByTestId("tags")).toHaveTextContent("[]");
    expect(saved).not.toHaveBeenCalled();
    fireEvent.compositionEnd(input, { data: "history" });
    await user.clear(input);
    await user.type(input, "読書");
    fireEvent.keyDown(input, { key: "Enter", code: "Enter", keyCode: 229 });
    expect(screen.getByTestId("tags")).toHaveTextContent("[]");
    await user.keyboard("{Enter}");
    expect(screen.getByRole("button", { name: "Remove 読書" })).toBeVisible();
  });

  it("retains input focus at the limit and allows removing a tag before creating another", async () => {
    const user = userEvent.setup();
    const saved = vi.fn();
    render(<TagsForm maxSelections={1} onSubmit={saved} />);
    const input = screen.getByRole("combobox", { name: "Tags" });
    await user.type(input, "weekend{Enter}");
    expect(input).toHaveFocus();
    expect(screen.getByRole("button", { name: "Show tags options" })).toBeDisabled();
    await user.keyboard("another{Enter}");
    expect(screen.getByTestId("tags")).toHaveTextContent('["weekend"]');
    expect(saved).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Remove weekend" }));
    expect(input).toHaveFocus();
    await user.type(input, "novel{Enter}");
    expect(screen.getByRole("button", { name: "Remove novel" })).toBeVisible();
  });

  it("honors RHF validation and reset for custom tags", async () => {
    const user = userEvent.setup();
    const saved = vi.fn();
    render(<TagsForm required onSubmit={saved} validate={tags => tags.includes("book club") || "Add your book club tag."} />);
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    const input = screen.getByRole("combobox", { name: /Tags/ });
    expect(input).toHaveAttribute("aria-invalid", "true");
    await user.type(input, "fiction{Enter}");
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Add your book club tag.");
    await user.type(input, "book club{Enter}");
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    await waitFor(() => expect(saved).toHaveBeenCalledTimes(1));
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.getByTestId("tags")).toHaveTextContent("[]");
  });

  it("disables additions and removal when the whole field is disabled", async () => {
    const user = userEvent.setup();
    render(<TagsForm disabled initialTags={["book club"]} />);
    expect(screen.getByRole("combobox", { name: "Tags" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Remove book club" }));
    expect(screen.getByTestId("tags")).toHaveTextContent('["book club"]');
    expect(screen.getByRole("button", { name: "Show tags options" })).toBeDisabled();
  });
});
