import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  Breadcrumb,
  Navbar,
  Pagination,
  Sidebar,
  Tabbar,
  VerticalTabbar,
  type NavigationLinkRenderer,
} from "../src/components/navigation";
import {
  phaseThreeNavigationDocuments,
  phaseThreeNavigationFixtures,
} from "../src/catalog/phase-three-navigation";
import { REQUIRED_COMPONENT_SECTION_KEYS } from "../src/catalog/schema";

const renderRouterLink: NavigationLinkRenderer = (props) => (
  <a data-router-link="true" {...props} />
);

const navbarItems = [
  { type: "link", id: "home", label: "Home", href: "/", current: true },
  { type: "link", id: "contests", label: "Contests", href: "/contests" },
  {
    type: "dropdown",
    id: "account",
    label: "Account",
    links: [
      { id: "profile", label: "Profile", href: "/profile" },
      { id: "settings", label: "Settings", href: "/settings" },
    ],
  },
] as const;

describe("router-neutral navigation", () => {
  it("renders Navbar links through the consumer adapter and restores mobile trigger focus on Escape", async () => {
    const user = userEvent.setup();
    render(
      <Navbar
        brand={<span>Tadoku</span>}
        brandHref="/"
        navigation={navbarItems}
        renderLink={renderRouterLink}
      />,
    );

    expect(screen.getByRole("link", { name: "Tadoku" })).toHaveAttribute("data-router-link", "true");
    expect(screen.getByRole("link", { name: "Home" })).toHaveAttribute("aria-current", "page");
    const trigger = screen.getByRole("button", { name: "Open menu" });
    await user.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("link", { name: "Profile" })).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(trigger).toHaveFocus();
  });

  it("uses Base UI keyboard behavior for the Navbar account dropdown", async () => {
    const user = userEvent.setup();
    render(<Navbar brand="Tadoku" brandHref="/" navigation={navbarItems} />);

    const trigger = screen.getByRole("button", { name: "Account" });
    trigger.focus();
    await user.keyboard("{Enter}");
    const profile = await screen.findByRole("menuitem", { name: "Profile" });
    expect(profile).toHaveFocus();
    await user.keyboard("{ArrowDown}");
    expect(screen.getByRole("menuitem", { name: "Settings" })).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(trigger).toHaveFocus();
  });

  it("keeps Navbar dropdown portals in the trigger owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <Navbar brand="Tadoku" brandHref="/" navigation={navbarItems} />,
      { container, baseElement: frameDocument.body },
    );

    await user.click(within(frameDocument.body).getByRole("button", { name: "Account" }));
    expect(await within(frameDocument.body).findByRole("menu")).toBeInTheDocument();
    expect(document.body.querySelector('[role="menu"]')).toBeNull();
    unmount();
    frame.remove();
  });

  it("marks Sidebar current state and removes disabled links from keyboard activation", async () => {
    const user = userEvent.setup();
    const disabledSelect = vi.fn();
    render(
      <Sidebar
        currentPath="/admin/logs"
        sections={[{
          id: "admin",
          title: "Admin",
          links: [
            { id: "logs", label: "Logs", href: "/admin/logs" },
            { id: "imports", label: "Imports", href: "/admin/imports", disabled: true, onSelect: disabledSelect },
          ],
        }]}
      />,
    );

    expect(screen.getByRole("link", { name: "Logs" })).toHaveAttribute("aria-current", "page");
    const disabled = screen.getByRole("link", { name: "Imports" });
    expect(disabled).toHaveAttribute("aria-disabled", "true");
    expect(disabled).toHaveAttribute("tabindex", "-1");
    await user.click(disabled);
    expect(disabledSelect).not.toHaveBeenCalled();
  });

  it("scopes Sidebar section heading ids to each component instance", () => {
    const sections = [{
      id: "admin tools",
      title: "Admin",
      links: [{ id: "logs", label: "Logs", href: "/admin/logs" }],
    }];
    const { container } = render(
      <>
        <Sidebar label="Desktop navigation" sections={sections} />
        <Sidebar label="Mobile navigation" sections={sections} />
      </>,
    );

    const headings = Array.from(container.querySelectorAll<HTMLHeadingElement>(".paper-sidebar__title"));
    const headingIds = headings.map((heading) => heading.id);
    expect(new Set(headingIds)).toHaveProperty("size", headingIds.length);
    for (const headingId of headingIds) {
      expect(headingId).toMatch(/^[a-zA-Z][a-zA-Z0-9_-]*$/u);
    }

    for (const section of container.querySelectorAll<HTMLElement>(".paper-sidebar__section")) {
      const labelledBy = section.getAttribute("aria-labelledby");
      expect(labelledBy).toBeTruthy();
      expect(container.querySelectorAll(`[id="${labelledBy}"]`)).toHaveLength(1);
      expect(section.querySelector("h2")).toHaveAttribute("id", labelledBy);
    }
  });

  it("uses ordered Breadcrumb semantics with a non-link current page", () => {
    render(
      <Breadcrumb items={[
        { id: "home", label: "Home", href: "/" },
        { id: "contests", label: "Contests", href: "/contests" },
        { id: "round", label: "August Japanese" },
      ]} />,
    );

    const breadcrumb = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(breadcrumb).getAllByRole("listitem")).toHaveLength(3);
    expect(within(breadcrumb).getByText("August Japanese").closest(".paper-breadcrumb__current")).toHaveAttribute("aria-current", "page");
    expect(within(breadcrumb).queryByRole("link", { name: "August Japanese" })).not.toBeInTheDocument();
  });

  it("keeps horizontal and vertical tab bars as linked navigation", () => {
    const links = [
      { id: "entries", label: "Entries", href: "/profile/entries" },
      { id: "stats", label: "Statistics", href: "/profile/stats", current: true },
    ];
    render(
      <>
        <Tabbar links={links} label="Profile views" renderLink={renderRouterLink} />
        <VerticalTabbar links={links} label="Profile views vertical" renderLink={renderRouterLink} />
      </>,
    );

    for (const navigation of screen.getAllByRole("navigation")) {
      const current = within(navigation).getByRole("link", { name: "Statistics" });
      expect(current).toHaveAttribute("aria-current", "page");
      expect(current).toHaveAttribute("data-router-link", "true");
    }
    expect(screen.queryByRole("tab")).not.toBeInTheDocument();
  });

  it("paginates with callbacks and disables unavailable edge controls", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();
    const { rerender } = render(
      <Pagination totalPages={12} currentPage={5} onPageChange={onPageChange} siblingCount={1} />,
    );

    expect(document.querySelectorAll(".paper-pagination__gap").length).toBeGreaterThan(0);
    expect(screen.getByText("5")).toHaveAttribute("aria-current", "page");
    await user.click(screen.getByRole("button", { name: "Next page" }));
    expect(onPageChange).toHaveBeenCalledWith(6);

    rerender(<Pagination totalPages={12} currentPage={1} onPageChange={onPageChange} />);
    expect(screen.queryByRole("button", { name: "Previous page" })).not.toBeInTheDocument();
    expect(screen.getByText("Previous").closest(".paper-pagination__control")).toHaveAttribute("aria-disabled", "true");
  });

  it("renders pagination hrefs through the router adapter and intercepts only when requested", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();
    render(
      <Pagination
        totalPages={3}
        currentPage={2}
        getHref={(page) => `/entries?page=${page}`}
        onPageChange={onPageChange}
        renderLink={renderRouterLink}
      />,
    );

    const next = screen.getByRole("link", { name: "Next page" });
    expect(next).toHaveAttribute("href", "/entries?page=3");
    expect(next).toHaveAttribute("data-router-link", "true");
    await user.click(screen.getByRole("link", { name: "Page 1" }));
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("ships complete Stable catalog records and deterministic fixtures", () => {
    expect(phaseThreeNavigationDocuments).toHaveLength(6);
    expect(phaseThreeNavigationFixtures).toHaveLength(6);
    expect(phaseThreeNavigationDocuments.map(({ route }) => route)).toEqual([
      "/components/navigation/navbar",
      "/components/navigation/sidebar",
      "/components/navigation/breadcrumb",
      "/components/navigation/tabbar",
      "/components/navigation/vertical-tabbar",
      "/components/navigation/pagination",
    ]);

    for (const document of phaseThreeNavigationDocuments) {
      expect(document.lifecycle).toBe("Stable");
      expect(Object.keys(document.sections?.required ?? {})).toEqual(
        expect.arrayContaining([...REQUIRED_COMPONENT_SECTION_KEYS]),
      );
      expect(document.fixtureIds).toHaveLength(1);
      expect(document.behaviorTestIds.length).toBeGreaterThanOrEqual(2);
    }
    for (const fixture of phaseThreeNavigationFixtures) {
      expect(fixture.deterministic).toBe(true);
      expect(fixture.code).toContain('from "paper-ui"');
    }
  });
});
