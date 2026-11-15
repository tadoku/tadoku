import { useStore } from "@tanstack/react-form";
import { fireEvent, render, screen } from "@testing-library/react";
import type { ClipboardEvent } from "react";
import { describe, expect, it, vi } from "vitest";
import { usePaperForm } from "../src";

interface InputProps {
  readonly disabled?: boolean;
  readonly readOnly?: boolean;
  readonly onPaste?: (event: ClipboardEvent<HTMLInputElement>) => void;
}

function AmountForm({ control, inputProps }: { control: "AmountWithUnit" | "Input"; inputProps?: InputProps }) {
  const form = usePaperForm({ defaultValues: { amount: 0, unit: "pages" } });
  const amount = useStore(form.store, state => state.values.amount);
  return (
    <form.AppForm>
      <form.AppField name="unit">
        {unitField => (
          <form.AppField name="amount">
            {field => control === "AmountWithUnit"
              ? <field.AmountWithUnit label="Amount" units={[{ value: "pages", label: "Pages" }]} unitField={unitField} step="any" min={0} {...inputProps} />
              : <field.Input label="Amount" type="number" step="any" min={0} {...inputProps} />}
          </form.AppField>
        )}
      </form.AppField>
      <output aria-label="Amount value">{JSON.stringify(amount)}</output>
    </form.AppForm>
  );
}

function paste(text: string) {
  return fireEvent.paste(screen.getByRole("spinbutton", { name: "Amount" }), { clipboardData: { getData: () => text } });
}

const amountInput = () => screen.getByRole("spinbutton", { name: "Amount" });

describe.each(["AmountWithUnit", "Input"] as const)("pasting into a number %s", control => {
  it.each([" 12.5", "12.5 ", "\t\n12.5\r\n", " 12.5 "])("trims %j into the field's number", pasted => {
    render(<AmountForm control={control} />);

    expect(paste(pasted)).toBe(false);
    expect(amountInput()).toHaveValue(12.5);
    expect(screen.getByLabelText("Amount value")).toHaveTextContent("12.5");
  });

  it.each(["12.5", "   ", " 12 5 ", " 12pages ", " 0x10 "])("leaves %j to native numeric-input handling", pasted => {
    render(<AmountForm control={control} />);

    expect(paste(pasted)).toBe(true);
    expect(screen.getByLabelText("Amount value")).toHaveTextContent("0");
  });

  it.each([{ disabled: true }, { readOnly: true }])("does not change an input with %j", inputProps => {
    render(<AmountForm control={control} inputProps={inputProps} />);

    paste(" 12.5 ");
    expect(screen.getByLabelText("Amount value")).toHaveTextContent("0");
  });

  it("honors a caller cancelling the paste", () => {
    const onPaste = vi.fn((event: ClipboardEvent<HTMLInputElement>) => event.preventDefault());
    render(<AmountForm control={control} inputProps={{ onPaste }} />);

    paste(" 12.5 ");
    expect(onPaste).toHaveBeenCalledOnce();
    expect(screen.getByLabelText("Amount value")).toHaveTextContent("0");
  });
});
