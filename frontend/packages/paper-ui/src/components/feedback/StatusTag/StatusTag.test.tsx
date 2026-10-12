import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatusTag } from "./StatusTag";

describe("StatusTag", () => {
  it.each([
    ["neutral", "Draft"],
    ["information", "Scheduled"],
    ["success", "Live now"],
    ["warning", "Registration closing"],
    ["danger", "Unavailable"],
  ] as const)("shows %s status as visible text through the shared recipe", (tone, label) => {
    render(<StatusTag tone={tone}>{label}</StatusTag>);

    const tag = screen.getByText(label);
    expect(tag).toBeVisible();
    expect(tag).toHaveClass("paper-status-tag", `paper-status-tag--${tone}`);
  });

  it("defaults to the neutral tone", () => {
    render(<StatusTag>Unlisted</StatusTag>);

    expect(screen.getByText("Unlisted")).toHaveClass("paper-status-tag--neutral");
  });

  it("labels static state without becoming a live region", () => {
    render(<StatusTag tone="success">Live now</StatusTag>);

    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("requires visible text so status never relies on colour alone", () => {
    // @ts-expect-error A status tag must contain its visible status text.
    const colourOnly = <StatusTag tone="success" />;

    expect(colourOnly).toBeTruthy();
  });
});
