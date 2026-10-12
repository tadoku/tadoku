import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { AppliedFilters } from "./AppliedFilters";

describe("AppliedFilters", () => {
  it("renders nothing while no filter is applied", () => {
    const { container } = render(<AppliedFilters filters={[]} onClear={vi.fn()} />);

    expect(container).toBeEmptyDOMElement();
  });

  it("names the group and lists every applied filter as text", () => {
    render(<AppliedFilters filters={["Japanese", "Name: Anton"]} onClear={vi.fn()} />);

    expect(screen.getByRole("group", { name: "Applied filters" })).toHaveTextContent(
      "Japanese · Name: Anton",
    );
  });

  it("clears every filter from one keyboard-operable action", async () => {
    const user = userEvent.setup();
    const clear = vi.fn();
    render(<AppliedFilters filters={["2025", "Japanese"]} onClear={clear} />);

    await user.tab();
    expect(screen.getByRole("button", { name: "Clear all" })).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(clear).toHaveBeenCalledTimes(1);
  });

  it("accepts localized names for the group and its action", () => {
    render(
      <AppliedFilters
        filters={["Japonais"]}
        onClear={vi.fn()}
        label="Filtres appliqués"
        clearLabel="Tout effacer"
      />,
    );

    expect(screen.getByRole("group", { name: "Filtres appliqués" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tout effacer" })).toBeInTheDocument();
  });
});
