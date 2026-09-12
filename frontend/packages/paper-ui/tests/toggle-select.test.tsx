import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { Button, ToggleSelect } from "../src";

function ModifierForm({ disabled = false, validate = false, onSubmit = vi.fn() }) {
  const methods = useForm({ defaultValues: { modifiers: [] as string[] } });
  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        <ToggleSelect
          name="modifiers"
          label="Score modifiers"
          hint="Choose one if it applies."
          disabled={disabled}
          rules={validate ? { validate: (value) => value.length > 0 || "Choose a modifier." } : undefined}
          options={[
            { value: "manga", label: "Manga", description: "×0.2" },
            { value: "comic", label: "Comic", description: "×0.1" },
            { value: "two-column", label: "Two column", description: "×2", disabled: true },
          ]}
        />
        <output aria-label="Current modifiers">{JSON.stringify(methods.watch("modifiers"))}</output>
        <Button onClick={() => methods.reset()}>Reset</Button>
        <Button type="submit">Save entry</Button>
      </form>
    </FormProvider>
  );
}

describe("ToggleSelect", () => {
  it("selects and clears a single optional value by keyboard without submitting or moving focus", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    render(<ModifierForm onSubmit={submit} />);
    const group = screen.getByRole("group", { name: "Score modifiers" });
    const manga = within(group).getByRole("button", { name: "Manga ×0.2", pressed: false });
    const comic = within(group).getByRole("button", { name: "Comic ×0.1", pressed: false });

    await user.tab();
    expect(manga).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(manga).toHaveAttribute("aria-pressed", "true");
    expect(manga).toHaveFocus();
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent('["manga"]');
    await user.keyboard(" ");
    expect(manga).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent("[]");

    await user.click(manga);
    await user.tab();
    expect(comic).toHaveFocus();
    await user.keyboard(" ");
    expect(comic).toHaveAttribute("aria-pressed", "true");
    expect(manga).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent('["comic"]');
    expect(submit).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(submit).toHaveBeenCalledWith({ modifiers: ["comic"] }, expect.anything());
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(comic).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent("[]");
  });

  it("keeps unavailable choices disabled and focuses the first enabled choice on a validation error", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    render(<ModifierForm validate onSubmit={submit} />);
    const disabledChoice = screen.getByRole("button", { name: "Two column ×2" });
    expect(disabledChoice).toBeDisabled();
    await user.click(disabledChoice);
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent("[]");
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Choose a modifier.");
    expect(screen.getByRole("button", { name: "Manga ×0.2" })).toHaveFocus();
    expect(screen.getByRole("group", { name: "Score modifiers" })).toHaveAccessibleDescription("Choose one if it applies. Choose a modifier.");
    expect(submit).not.toHaveBeenCalled();
  });

  it("disables every choice when the field is unavailable", async () => {
    const user = userEvent.setup();
    render(<ModifierForm disabled />);
    for (const button of within(screen.getByRole("group")).getAllByRole("button")) {
      expect(button).toBeDisabled();
      await user.click(button);
    }
    expect(screen.getByLabelText("Current modifiers")).toHaveTextContent("[]");
    await user.tab();
    expect(screen.getByRole("button", { name: "Reset" })).toHaveFocus();
  });
});
