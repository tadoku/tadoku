import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { SectionHeading } from "./SectionHeading";

describe("SectionHeading", () => {
  it.each([2, 3, 4, 5, 6] as const)("renders the requested level-%i heading", (level) => {
    render(<SectionHeading level={level} title="Contest history" />);

    expect(screen.getByRole("heading", { level, name: "Contest history" })).toBeInTheDocument();
    expect(screen.getAllByRole("heading")).toHaveLength(1);
  });

  it("labels its enclosing section through the heading id", () => {
    render(
      <section aria-labelledby="history-heading">
        <SectionHeading level={2} id="history-heading" title="Contest history" />
      </section>,
    );

    expect(screen.getByRole("region", { name: "Contest history" })).toBeInTheDocument();
  });

  it("adds description and status without renaming the heading", () => {
    render(
      <SectionHeading
        level={2}
        title="Current contest"
        description="2026 Round 5"
        status={<span>Live now</span>}
      />,
    );

    expect(screen.getByRole("heading", { level: 2 })).toHaveAccessibleName("Current contest");
    expect(screen.getByText("2026 Round 5")).toBeVisible();
    expect(screen.getByText("Live now")).toBeVisible();
  });

  it("keeps section actions in the tab order", async () => {
    const user = userEvent.setup();
    render(
      <SectionHeading
        level={2}
        title="Contest history"
        actions={<a href="/contests/official">All contests</a>}
      />,
    );

    await user.tab();
    expect(screen.getByRole("link", { name: "All contests" })).toHaveFocus();
  });

  it("leaves the level-one page heading to PageHeader", () => {
    // @ts-expect-error SectionHeading starts at level 2; PageHeader owns the page's h1.
    const pageHeading = <SectionHeading level={1} title="Contests" />;

    expect(pageHeading).toBeTruthy();
  });
});
