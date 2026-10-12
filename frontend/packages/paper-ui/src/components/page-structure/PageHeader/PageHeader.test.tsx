import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { NavigationLinkProps } from "../../navigation";
import { PageHeader } from "./PageHeader";

describe("PageHeader", () => {
  it("names the page with its only level-one heading and keeps the lead outside it", () => {
    render(
      <PageHeader
        title="Contests"
        lead="Official rounds bring the Tadoku community together throughout the year."
      />,
    );

    expect(screen.getAllByRole("heading")).toHaveLength(1);
    expect(screen.getByRole("heading", { level: 1 })).toHaveAccessibleName("Contests");
    expect(
      screen.getByText("Official rounds bring the Tadoku community together throughout the year."),
    ).toBeVisible();
  });

  it("shows status beside the title without adding it to the heading name", () => {
    render(
      <PageHeader
        title="Our Sunday reading group"
        status={<span>Unlisted</span>}
      />,
    );

    expect(screen.getByRole("heading", { level: 1 })).toHaveAccessibleName("Our Sunday reading group");
    expect(screen.getByText("Unlisted")).toBeVisible();
  });

  it("renders the back link through the consumer's link renderer, before the heading", () => {
    const renderLink = vi.fn(({ href, ...props }: NavigationLinkProps) => (
      <a {...props} href={`/app${href}`} data-router-link="" />
    ));
    render(
      <PageHeader
        title="Join 2026 Round 5"
        backLink={{ href: "/contests/round5", label: "Back to 2026 Round 5" }}
        renderLink={renderLink}
      />,
    );

    const link = screen.getByRole("link", { name: "Back to 2026 Round 5" });
    expect(link).toHaveAttribute("data-router-link");
    expect(link).toHaveAttribute("href", "/app/contests/round5");
    expect(renderLink).toHaveBeenCalledWith(expect.objectContaining({ href: "/contests/round5" }));
    expect(
      link.compareDocumentPosition(screen.getByRole("heading", { level: 1 })) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("falls back to a native link when the application supplies no renderer", () => {
    render(
      <PageHeader
        title="Account settings"
        backLink={{ href: "/users/anton", label: "Your profile" }}
      />,
    );

    expect(screen.getByRole("link", { name: "Your profile" })).toHaveAttribute("href", "/users/anton");
  });

  it("keeps page actions operable from the keyboard after the back link", async () => {
    const user = userEvent.setup();
    const createContest = vi.fn();
    render(
      <PageHeader
        title="Contests"
        backLink={{ href: "/", label: "Home" }}
        actions={<button type="button" onClick={createContest}>Create contest</button>}
      />,
    );

    await user.tab();
    expect(screen.getByRole("link", { name: "Home" })).toHaveFocus();
    await user.tab();
    expect(screen.getByRole("button", { name: "Create contest" })).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(createContest).toHaveBeenCalledTimes(1);
  });

  it("keeps supporting content, such as a step indicator, inside the header", () => {
    render(
      <PageHeader title="Log activity">
        <p>Log details · 1 of 2</p>
      </PageHeader>,
    );

    expect(screen.getByText("Log details · 1 of 2").closest("header")).toContainElement(
      screen.getByRole("heading", { level: 1 }),
    );
  });
});
