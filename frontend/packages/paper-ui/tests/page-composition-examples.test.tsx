import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { catalogRegistry } from "../src/catalog";

function example(id: string) {
  const fixture = catalogRegistry.fixtures.find((item) => item.id === id);
  if (!fixture) throw new Error(`Missing catalogue fixture ${id}`);
  return render(<>{fixture.render()}</>);
}

describe("page composition examples teach observable outcomes", () => {
  it("names the contest page once and keeps its back link, status and actions outside the heading", () => {
    example("page-header.contest-detail");

    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByRole("heading", { level: 1 })).toHaveAccessibleName("2026 Round 5");
    expect(screen.getByRole("link", { name: "Back to official contests" })).toBeInTheDocument();
    expect(screen.getByText("Live now")).toBeVisible();
    expect(screen.getByRole("link", { name: "Join contest" })).toBeInTheDocument();
  });

  it("nests profile section headings in outline order", () => {
    example("section-heading.profile");

    expect(screen.getAllByRole("heading").map((heading) => [heading.tagName, heading.textContent])).toEqual([
      ["H2", "Anton’s activity"],
      ["H3", "Immersion rhythm"],
    ]);
    expect(screen.getByRole("region", { name: "Anton’s activity" })).toContainElement(
      screen.getByRole("heading", { level: 3 }),
    );
  });

  it("labels every status tone with text", () => {
    example("status-tag.tones");

    for (const label of ["Unlisted", "Scheduled", "Live now", "Draft", "Unavailable"]) {
      expect(screen.getByText(label)).toBeVisible();
    }
  });

  it("recovers from a search without results and returns focus to the search", async () => {
    const user = userEvent.setup();
    example("empty-state.no-results");

    expect(screen.getByRole("heading", { level: 2, name: /No contests match/ })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Clear search" }));

    const search = screen.getByRole("searchbox", { name: "Find a contest" });
    expect(search).toHaveFocus();
    expect(search).toHaveValue("");
    expect(screen.queryByRole("heading", { name: /No contests match/ })).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("3 contests");
  });

  it("offers the first entry from an empty year inside its section", () => {
    example("empty-state.first-activity");

    const section = screen.getByRole("region", { name: "Anton’s activity" });
    expect(within(section).getByRole("heading", { level: 3 })).toHaveAccessibleName("No activity recorded in 2026");
    expect(within(section).getByRole("link", { name: "Log activity" })).toBeInTheDocument();
  });

  it("pairs reading-year figures with their labels", () => {
    example("stat-grid.reading-year");

    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual([
      "Personal score",
      "Entries",
      "Languages",
      "Contests",
    ]);
    expect(screen.getAllByRole("definition")[1]).toHaveTextContent("12 active days");
  });

  it("presents contest facts in both layouts as terms and values", () => {
    example("description-list.contest-facts");

    expect(screen.getAllByRole("term")).toHaveLength(6);
    expect(screen.getAllByRole("definition")).toHaveLength(6);
    expect(screen.getByRole("region", { name: "Current contest" })).toHaveTextContent("Submitted entries");
  });

  it("summarizes collapsed filters and clears them back to the toggle", async () => {
    const user = userEvent.setup();
    example("pattern.filter-bar");

    const toggle = screen.getByRole("button", { name: "Filters" });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("combobox", { name: "Language" })).not.toBeInTheDocument();

    await user.click(toggle);
    await user.selectOptions(screen.getByRole("combobox", { name: "Language" }), "Spanish");
    await user.click(toggle);

    expect(toggle).toHaveAccessibleName("Filters (1)");
    expect(screen.getByRole("group", { name: "Applied filters" })).toHaveTextContent("Spanish");
    expect(screen.getByRole("status")).toHaveTextContent("2 of 5 participants");

    await user.click(screen.getByRole("button", { name: "Clear all" }));
    expect(toggle).toHaveFocus();
    expect(toggle).toHaveAccessibleName("Filters");
    expect(screen.queryByRole("group", { name: "Applied filters" })).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("5 of 5 participants");
  });

  it("explains a filtered standings view without matches and recovers from it", async () => {
    const user = userEvent.setup();
    example("pattern.filter-bar");

    const toggle = screen.getByRole("button", { name: "Filters" });
    await user.click(toggle);
    await user.type(screen.getByRole("searchbox", { name: "Find a participant" }), "Zed");

    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 3, name: "No participants match these filters." })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(toggle).toHaveFocus();
    expect(screen.getByRole("table")).toBeInTheDocument();
  });
});
