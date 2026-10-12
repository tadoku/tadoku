import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { DescriptionList } from "./DescriptionList";

const facts = [
  { label: "Round begins", value: "Sep 1, 2026, 12:00 AM UTC" },
  { label: "Moderators", value: <a href="/users/kai">Kai</a> },
];

describe("DescriptionList", () => {
  it.each(["stacked", "inline"] as const)("keeps term and description pairs grouped in the %s layout", (layout) => {
    render(<DescriptionList layout={layout} items={facts} />);

    const terms = screen.getAllByRole("term");
    const definitions = screen.getAllByRole("definition");
    expect(terms.map((term) => term.textContent)).toEqual(["Round begins", "Moderators"]);
    expect(definitions[0]).toHaveTextContent("Sep 1, 2026, 12:00 AM UTC");
    terms.forEach((term, index) => {
      const group = term.parentElement!;
      expect(group.tagName).toBe("DIV");
      expect(group.parentElement?.tagName).toBe("DL");
      expect(group).toContainElement(definitions[index]);
    });
  });

  it("keeps linked values in the keyboard order", async () => {
    const user = userEvent.setup();
    render(<DescriptionList items={facts} />);

    await user.tab();
    expect(screen.getByRole("link", { name: "Kai" })).toHaveFocus();
  });
});
