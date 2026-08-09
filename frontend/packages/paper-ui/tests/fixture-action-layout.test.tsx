import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { catalogRegistry } from "../src/catalog";

const sharedFormFixtureIds = [
  "textarea.reading-notes",
  "select.language",
  "checkbox.public-entry",
  "radio-select.viewport",
  "radio-group.format",
  "amount.progress",
  "autocomplete.language",
  "multi-autocomplete.languages",
  "tags.entry",
] as const;

function renderFixture(id: string) {
  const fixture = catalogRegistry.fixtures.find((candidate) => candidate.id === id);
  expect(fixture, `catalogue fixture ${id}`).toBeDefined();

  render(<>{fixture?.render()}</>);
}

function expectSubmitActionRow(label: string) {
  const submit = screen.getByRole("button", { name: label });

  expect(submit).toHaveAttribute("type", "submit");
  expect(submit.parentElement).toHaveClass("paper-fixture-row");
}

describe("form fixture action layout", () => {
  it("keeps the recommended Input submit action from stretching with its form grid", () => {
    renderFixture("input.recommended");

    expectSubmitActionRow("Save log");
  });

  it.each(sharedFormFixtureIds)(
    "keeps the %s submit action from stretching with the shared form grid",
    (fixtureId) => {
      renderFixture(fixtureId);

      expectSubmitActionRow("Save entry");
    },
  );

  it("keeps the Logging v2 submit action from stretching with its form grid", () => {
    renderFixture("experiment.logging-v2-entry");

    expectSubmitActionRow("Review entry");
  });

  it("preserves the narrow-viewport fixture as an intentional full-width button", () => {
    renderFixture("button.narrow");

    const button = screen.getByRole("button", {
      name: "Add this finished book to the August reading log",
    });
    expect(button).toHaveClass("paper-button--full-width");
    expect(button.parentElement).not.toHaveClass("paper-fixture-row");
  });
});
