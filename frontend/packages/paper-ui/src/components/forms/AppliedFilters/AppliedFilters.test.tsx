import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
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

function RemovableFilters({ initial, onRemove = vi.fn() }: { initial: string[]; onRemove?: (label: string) => void }) {
  const [applied, setApplied] = useState(initial);
  return (
    <AppliedFilters
      variant="removable"
      filters={applied.map((label) => ({
        id: label,
        label,
        onRemove: () => {
          onRemove(label);
          setApplied((current) => current.filter((item) => item !== label));
        },
      }))}
      onClear={() => setApplied([])}
      resultLabel={`${83 - applied.length} logs of 127`}
    />
  );
}

describe("AppliedFilters removable", () => {
  it("removes only the value whose chip is pressed", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    render(<RemovableFilters initial={["Listening", "audiobook"]} onRemove={onRemove} />);

    await user.click(screen.getByRole("button", { name: "Remove Listening" }));

    expect(onRemove).toHaveBeenCalledExactlyOnceWith("Listening");
    expect(screen.queryByRole("button", { name: "Remove Listening" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove audiobook" })).toBeInTheDocument();
  });

  it("keeps focus on the chips after one is removed", async () => {
    const user = userEvent.setup();
    render(<RemovableFilters initial={["Listening", "audiobook", "dense"]} />);

    screen.getByRole("button", { name: "Remove audiobook" }).focus();
    await user.keyboard("{Enter}");
    expect(screen.getByRole("button", { name: "Remove dense" })).toHaveFocus();

    await user.keyboard("{Enter}");
    expect(screen.getByRole("button", { name: "Remove Listening" })).toHaveFocus();
  });

  it("shows the result count and one Clear all with the chips", async () => {
    const user = userEvent.setup();
    render(<RemovableFilters initial={["Listening"]} />);

    const group = screen.getByRole("group", { name: "Applied filters" });
    expect(group).toHaveTextContent("82 logs of 127");
    expect(screen.getAllByRole("button", { name: "Clear all" })).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "Clear all" }));
    expect(screen.queryByRole("group", { name: "Applied filters" })).not.toBeInTheDocument();
  });
});
