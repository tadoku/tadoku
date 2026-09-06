import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, it } from "vitest";
import { phaseThreeContentFixtures } from "../src/catalog/phase-three-content";

it("reviews entered reading and returns to editing without losing it", async () => {
  const user = userEvent.setup();
  render(<>{phaseThreeContentFixtures[1].render()}</>);
  await user.clear(screen.getByRole("spinbutton", { name: "Pages" }));
  await user.type(screen.getByRole("spinbutton", { name: "Pages" }), "72");
  await user.click(screen.getByRole("button", { name: "Review entry" }));
  expect(await screen.findByRole("heading", { name: "Review reading" })).toBeVisible();
  expect(screen.getByText(/72 pages of コンビニ人間/)).toBeVisible();
  expect(screen.getByRole("heading", { name: "Review reading" })).toHaveFocus();
  await user.click(screen.getByRole("button", { name: "Edit entry" }));
  expect(screen.getByRole("spinbutton", { name: "Pages" })).toHaveValue(72);
  expect(screen.getByRole("textbox", { name: "Work" })).toHaveFocus();
});
