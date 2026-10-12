import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatGrid } from "./StatGrid";

const stats = [
  { label: "Personal score", value: "1,240" },
  { label: "Entries", value: "18", detail: "12 active days" },
  { label: "Languages", value: "2", detail: "Japanese leads" },
];

describe("StatGrid", () => {
  it("pairs every label with its value in one description list", () => {
    render(<StatGrid stats={stats} />);

    const terms = screen.getAllByRole("term");
    const definitions = screen.getAllByRole("definition");
    expect(terms.map((term) => term.textContent)).toEqual(["Personal score", "Entries", "Languages"]);
    expect(definitions).toHaveLength(3);
    expect(definitions[0]).toHaveTextContent("1,240");

    terms.forEach((term, index) => {
      const group = term.parentElement!;
      expect(group.tagName).toBe("DIV");
      expect(group.parentElement?.tagName).toBe("DL");
      expect(group).toContainElement(definitions[index]);
      expect(group.children).toHaveLength(2);
    });
  });

  it("keeps supporting detail with the value it explains", () => {
    render(<StatGrid stats={stats} />);

    const entries = screen.getAllByRole("definition")[1];
    expect(entries).toHaveTextContent("18");
    expect(entries).toHaveTextContent("12 active days");
  });
});
