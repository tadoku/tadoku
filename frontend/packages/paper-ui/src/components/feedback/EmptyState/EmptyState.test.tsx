import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect, useRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { EmptyState } from "./EmptyState";

describe("EmptyState", () => {
  it("places its title at the heading level the screen chooses", () => {
    const { rerender } = render(<EmptyState headingLevel={1} title="Contest not found" />);
    expect(screen.getByRole("heading", { level: 1, name: "Contest not found" })).toBeInTheDocument();

    rerender(<EmptyState headingLevel={3} title="No activity yet" />);
    expect(screen.getByRole("heading", { level: 3, name: "No activity yet" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { level: 1 })).not.toBeInTheDocument();
  });

  it("explains the state and offers a keyboard-operable next step", async () => {
    const user = userEvent.setup();
    const clearFilters = vi.fn();
    render(
      <EmptyState
        headingLevel={2}
        title="No contests match these filters."
        description="Try another title or language, or include all dates."
        actions={<button type="button" onClick={clearFilters}>Clear filters</button>}
      />,
    );

    expect(screen.getByText("Try another title or language, or include all dates.")).toBeVisible();
    await user.tab();
    expect(screen.getByRole("button", { name: "Clear filters" })).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(clearFilters).toHaveBeenCalledTimes(1);
  });

  it("announces a failure when the screen marks the state as an alert", () => {
    render(
      <EmptyState
        role="alert"
        headingLevel={2}
        title="Contests could not be loaded."
        description="Try loading this collection again."
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("Contests could not be loaded.");
  });

  it("lets the screen move focus to the heading after a state change", () => {
    function DeletedLog() {
      const heading = useRef<HTMLHeadingElement>(null);
      useEffect(() => heading.current?.focus(), []);
      return <EmptyState headingLevel={1} headingRef={heading} title="Log deleted" />;
    }
    render(<DeletedLog />);

    expect(screen.getByRole("heading", { name: "Log deleted" })).toHaveFocus();
  });

  it("does not add unfocusable headings to the focus order", () => {
    render(<EmptyState headingLevel={2} title="No questions match that search." />);

    expect(screen.getByRole("heading")).not.toHaveAttribute("tabindex");
  });

  it("requires the screen to choose a heading level", () => {
    // @ts-expect-error The consuming screen decides where the state sits in its outline.
    const unplaced = <EmptyState title="No activity yet" />;

    expect(unplaced).toBeTruthy();
  });
});
