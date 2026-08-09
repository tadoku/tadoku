import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { render, screen } from "@testing-library/react";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it } from "vitest";
import { Select } from "../src";

const formsCss = readFileSync(
  resolve("src/components/forms/forms.css"),
  "utf8",
);

function SelectFixture() {
  const methods = useForm({ defaultValues: { language: "zh" } });

  return (
    <FormProvider {...methods}>
      <Select
        name="language"
        label="Language"
        options={[
          {
            value: "zh",
            label: "Chinese with an intentionally long option label",
          },
        ]}
      />
    </FormProvider>
  );
}

describe("Select chevron visual contract", () => {
  it("reserves token-based space and balances the chevron against the text inset", () => {
    render(<SelectFixture />);

    const select = screen.getByRole("combobox", { name: "Language" });
    const control = select.parentElement;
    const chevron = control?.querySelector("svg");

    expect(control).toHaveClass("paper-select__control");
    expect(chevron).toHaveClass("paper-select__icon");
    expect(chevron).toHaveAttribute("aria-hidden", "true");

    expect(formsCss).toMatch(
      /\.paper-select\s*\{[^}]*appearance:\s*none;[^}]*padding-inline-end:\s*calc\(var\(--paper-field-padding-inline\) \* 2 \+ var\(--paper-icon-size-compact\)\);/s,
    );
    expect(formsCss).toMatch(
      /\.paper-select__icon\s*\{[^}]*inset-inline-end:\s*var\(--paper-field-padding-inline\);[^}]*pointer-events:\s*none;/s,
    );
  });

  it("keeps the chevron above the focused Select stacking layer", () => {
    expect(formsCss).toMatch(
      /\.paper-select__icon\s*\{[^}]*z-index:\s*2;/s,
    );
  });
});
