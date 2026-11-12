import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { FilterMenu, type FilterMenuHandle, type FilterMenuSection } from "./FilterMenu";

const tags = ["audiobook", "dense", "fiction", "manga", "news", "podcast", "graded", "novel", "drama"];

const baseSections: readonly FilterMenuSection[] = [
  {
    id: "activity",
    label: "Activity",
    values: [
      { value: "reading", label: "Reading", points: 1240.5 },
      { value: "listening", label: "Listening", points: 830 },
      { value: "speaking", label: "Speaking", points: 0 },
    ],
    selected: [],
  },
  {
    id: "tag",
    label: "Tags",
    values: tags.map((tag, index) => ({ value: tag, label: tag, points: 100 - index })),
    selected: [],
  },
];

function Controlled({
  sections = baseSections,
  onChange = vi.fn(),
  onClear = vi.fn(),
}: {
  sections?: readonly FilterMenuSection[];
  onChange?: (sectionId: string, values: readonly string[]) => void;
  onClear?: () => void;
}) {
  const [selected, setSelected] = useState<Record<string, readonly string[]>>(
    Object.fromEntries(sections.map((section) => [section.id, section.selected])),
  );
  const count = Object.values(selected).flat().length;
  return (
    <FilterMenu
      sections={sections.map((section) => ({ ...section, selected: selected[section.id] ?? [] }))}
      onChange={(sectionId, values) => {
        onChange(sectionId, values);
        setSelected((current) => ({ ...current, [sectionId]: values }));
      }}
      onClear={() => {
        onClear();
        setSelected({});
      }}
      resultLabel={`${83 - count} logs`}
    />
  );
}

const menuTrigger = () =>
  document.querySelector<HTMLElement>(".paper-filter-menu__trigger--menu")!;
const sheetTrigger = () =>
  document.querySelector<HTMLElement>(".paper-filter-menu__trigger--sheet")!;

describe("FilterMenu on desktop", () => {
  it("keeps the menu open while several values are ticked and reports every selection", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Controlled onChange={onChange} />);

    await user.click(menuTrigger());
    const menu = await screen.findByRole("menu");
    await user.click(within(menu).getByRole("menuitemcheckbox", { name: /Reading/ }));
    await user.click(within(menu).getByRole("menuitemcheckbox", { name: /Listening/ }));

    expect(screen.getByRole("menu")).toBeInTheDocument();
    expect(within(menu).getByRole("menuitemcheckbox", { name: /Reading/ })).toHaveAttribute("aria-checked", "true");
    expect(onChange).toHaveBeenLastCalledWith("activity", ["reading", "listening"]);
    expect(menuTrigger()).toHaveAccessibleName("Filters (2)");
  });

  it("removes a value when it is unticked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <Controlled
        onChange={onChange}
        sections={[{ ...baseSections[0], selected: ["reading", "listening"] }]}
      />,
    );

    await user.click(menuTrigger());
    await user.click(await screen.findByRole("menuitemcheckbox", { name: /Reading/ }));
    expect(onChange).toHaveBeenLastCalledWith("activity", ["listening"]);
  });

  it.each([
    ["Escape", async (user: ReturnType<typeof userEvent.setup>) => user.keyboard("{Escape}")],
    ["an outside press", async (user: ReturnType<typeof userEvent.setup>) => user.click(document.body)],
  ])("closes on %s and returns focus to the trigger", async (_name, dismiss) => {
    // Base UI returns focus after an outside press only where focus() reads preventScroll, which jsdom does not.
    const focus = HTMLElement.prototype.focus;
    vi.spyOn(HTMLElement.prototype, "focus").mockImplementation(function (this: HTMLElement, options) {
      void options?.preventScroll;
      focus.call(this, options);
    });
    const user = userEvent.setup();
    render(<Controlled />);

    await user.click(menuTrigger());
    await screen.findByRole("menu");
    await dismiss(user);

    await waitFor(() => expect(screen.queryByRole("menu")).not.toBeInTheDocument());
    expect(menuTrigger()).toHaveFocus();
  });

  it("names each value with its points and hides values without points", async () => {
    const user = userEvent.setup();
    render(<Controlled />);

    await user.click(menuTrigger());
    const menu = await screen.findByRole("menu");
    expect(within(menu).getByRole("menuitemcheckbox", { name: /Reading/ })).toHaveTextContent("1,240.5");
    expect(within(menu).queryByRole("menuitemcheckbox", { name: /Speaking/ })).not.toBeInTheDocument();
  });

  it("keeps a chosen value listed and checked when it has no points", async () => {
    const user = userEvent.setup();
    render(<Controlled sections={[{ ...baseSections[0], selected: ["speaking"] }]} />);

    await user.click(menuTrigger());
    expect(await screen.findByRole("menuitemcheckbox", { name: /Speaking/ })).toHaveAttribute("aria-checked", "true");
  });

  it("shows the first values, keeps chosen later values, and reveals the rest from Show all", async () => {
    const user = userEvent.setup();
    render(<Controlled sections={[{ ...baseSections[1], selected: ["drama"] }]} />);

    await user.click(menuTrigger());
    const menu = await screen.findByRole("menu");
    const names = () => within(menu).getAllByRole("menuitemcheckbox").map((item) => item.textContent);
    expect(names()).toHaveLength(7);
    expect(within(menu).getByRole("menuitemcheckbox", { name: /drama/ })).toHaveAttribute("aria-checked", "true");
    expect(within(menu).queryByRole("menuitemcheckbox", { name: /graded/ })).not.toBeInTheDocument();

    await user.click(within(menu).getByRole("menuitem", { name: "Show all 9" }));
    expect(screen.getByRole("menu")).toBeInTheDocument();
    expect(names()).toHaveLength(9);
    expect(within(menu).queryByRole("menuitem", { name: /Show all/ })).not.toBeInTheDocument();
  });

  it("clears one section from its Clear and leaves the others", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <Controlled
        onChange={onChange}
        sections={[
          { ...baseSections[0], selected: ["reading"] },
          { ...baseSections[1], selected: ["manga"] },
        ]}
      />,
    );

    await user.click(menuTrigger());
    const activity = within(await screen.findByRole("group", { name: "Activity" }));
    await user.click(activity.getByRole("menuitem", { name: "Clear Activity" }));

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenLastCalledWith("activity", []);
    expect(screen.getByRole("menu")).toBeInTheDocument();
    expect(menuTrigger()).toHaveAccessibleName("Filters (1)");
    expect(within(screen.getByRole("group", { name: "Tags" })).getByRole("menuitem", { name: "Clear Tags" }))
      .toBeInTheDocument();
  });

  it("clears everything from Clear all, closes and returns focus to the trigger", async () => {
    const user = userEvent.setup();
    const onClear = vi.fn();
    render(<Controlled onClear={onClear} sections={[{ ...baseSections[0], selected: ["reading"] }]} />);

    await user.click(menuTrigger());
    const menu = await screen.findByRole("menu");
    expect(menu).toHaveTextContent("82 logs");
    await user.click(within(menu).getByRole("menuitem", { name: "Clear all" }));

    expect(onClear).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(screen.queryByRole("menu")).not.toBeInTheDocument());
    expect(menuTrigger()).toHaveFocus();
    expect(menuTrigger()).toHaveAccessibleName("Filters");
  });

  it("offers no Clear or Clear all while nothing is chosen", async () => {
    const user = userEvent.setup();
    render(<Controlled />);

    await user.click(menuTrigger());
    const menu = await screen.findByRole("menu");
    expect(within(menu).queryByRole("menuitem", { name: /Clear/ })).not.toBeInTheDocument();
    expect(menu).toHaveTextContent("83 logs");
  });
});

describe("FilterMenu on a phone", () => {
  it("opens a sheet that applies choices as they are ticked and closes from its result button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Controlled onChange={onChange} />);

    await user.click(sheetTrigger());
    const sheet = await screen.findByRole("dialog", { name: "Filters" });
    await user.click(within(sheet).getByRole("checkbox", { name: /Listening/ }));

    expect(onChange).toHaveBeenLastCalledWith("activity", ["listening"]);
    expect(within(sheet).getByRole("checkbox", { name: /Listening/ })).toBeChecked();
    await user.click(within(sheet).getByRole("button", { name: "Show 82 logs" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(sheetTrigger()).toHaveFocus();
    expect(sheetTrigger()).toHaveAccessibleName("Filters (1)");
  });

  it("traps focus inside the sheet", async () => {
    const user = userEvent.setup();
    render(<Controlled />);

    await user.click(sheetTrigger());
    const sheet = await screen.findByRole("dialog", { name: "Filters" });
    for (let step = 0; step < 20; step += 1) {
      await user.tab();
      await waitFor(() => expect(sheet).toContainElement(document.activeElement as HTMLElement));
    }
  });

  it("lists more values before Show all than the desktop menu", async () => {
    const user = userEvent.setup();
    render(<Controlled sections={[baseSections[1]]} />);

    await user.click(sheetTrigger());
    const sheet = await screen.findByRole("dialog", { name: "Filters" });
    expect(within(sheet).getAllByRole("checkbox")).toHaveLength(8);
    await user.click(within(sheet).getByRole("button", { name: "Show all 9" }));
    expect(within(sheet).getAllByRole("checkbox")).toHaveLength(9);
  });

  it("clears one section, or everything and returns focus to the trigger", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const onClear = vi.fn();
    render(
      <Controlled
        onChange={onChange}
        onClear={onClear}
        sections={[
          { ...baseSections[0], selected: ["reading"] },
          { ...baseSections[1], selected: ["manga"] },
        ]}
      />,
    );

    await user.click(sheetTrigger());
    const sheet = await screen.findByRole("dialog", { name: "Filters" });
    await user.click(within(sheet).getByRole("button", { name: "Clear Tags" }));
    expect(onChange).toHaveBeenLastCalledWith("tag", []);
    expect(within(sheet).getByRole("checkbox", { name: /Reading/ })).toBeChecked();

    await user.click(within(sheet).getByRole("button", { name: "Clear all" }));
    expect(onClear).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(sheetTrigger()).toHaveFocus();
  });
});

describe("FilterMenu focus handle", () => {
  it("lets the screen return focus to the trigger after clearing filters elsewhere", () => {
    const ref = createRef<FilterMenuHandle>();
    render(<FilterMenu ref={ref} sections={baseSections} onChange={vi.fn()} onClear={vi.fn()} resultLabel="83 logs" />);

    ref.current?.focus();
    expect(document.activeElement).toHaveAccessibleName("Filters");
  });
});
