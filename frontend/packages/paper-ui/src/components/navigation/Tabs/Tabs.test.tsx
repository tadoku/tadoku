import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { Tabs } from "./Tabs";

function Example({ onChange = vi.fn() }: { onChange?: (value: string) => void }) {
  return (
    <Tabs.Root defaultValue="overview" onValueChange={onChange}>
      <Tabs.List aria-label="Project sections">
        <Tabs.Tab value="overview">Overview</Tabs.Tab>
        <Tabs.Tab value="history" disabled>
          History
        </Tabs.Tab>
        <Tabs.Tab value="settings">Settings</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="overview">Overview panel</Tabs.Panel>
      <Tabs.Panel value="history">History panel</Tabs.Panel>
      <Tabs.Panel value="settings">Settings panel</Tabs.Panel>
    </Tabs.Root>
  );
}

describe("Tabs", () => {
  it("connects tabs and panels with the expected accessibility relationships", () => {
    render(<Example />);

    const list = screen.getByRole("tablist", { name: "Project sections" });
    const overview = screen.getByRole("tab", { name: "Overview" });
    const overviewPanel = screen.getByRole("tabpanel", { name: "Overview" });

    expect(list).toHaveAttribute("data-orientation", "horizontal");
    expect(overview).toHaveAttribute("aria-selected", "true");
    expect(overview).toHaveAttribute("tabindex", "0");
    expect(overview.id).not.toBe("");
    expect(overviewPanel.id).not.toBe("");
    expect(overview).toHaveAttribute("aria-controls", overviewPanel.id);
    expect(overviewPanel).toHaveAttribute("aria-labelledby", overview.id);
  });

  it("uses roving focus and automatic selection while skipping disabled tabs", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Example onChange={onChange} />);

    const overview = screen.getByRole("tab", { name: "Overview" });
    const settings = screen.getByRole("tab", { name: "Settings" });
    const history = screen.getByRole("tab", { name: "History" });
    await user.click(overview);

    await user.keyboard("{ArrowRight}");
    expect(settings).toHaveFocus();
    expect(settings).toHaveAttribute("aria-selected", "true");
    expect(overview).toHaveAttribute("tabindex", "-1");
    expect(history).toHaveAttribute("aria-disabled", "true");
    expect(history).toHaveAttribute("tabindex", "-1");
    expect(onChange.mock.lastCall?.[0]).toBe("settings");

    await user.keyboard("{Home}");
    expect(overview).toHaveFocus();
    await user.keyboard("{End}");
    expect(settings).toHaveFocus();
    await user.keyboard("{ArrowLeft}");
    expect(overview).toHaveFocus();
  });

  it("supports controlled selection", async () => {
    const user = userEvent.setup();

    function ControlledExample() {
      const [value, setValue] = useState("one");
      return (
        <Tabs.Root value={value} onValueChange={setValue}>
          <Tabs.List aria-label="Controlled tabs">
            <Tabs.Tab value="one">One</Tabs.Tab>
            <Tabs.Tab value="two">Two</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="one">First</Tabs.Panel>
          <Tabs.Panel value="two">Second</Tabs.Panel>
        </Tabs.Root>
      );
    }

    render(<ControlledExample />);
    await user.click(screen.getByRole("tab", { name: "Two" }));
    expect(screen.getByRole("tab", { name: "Two" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("tabpanel", { name: "Two" })).toBeVisible();
  });

  it("keeps inactive panels mounted and hidden by default, with an opt-out", async () => {
    const user = userEvent.setup();
    const { rerender } = render(<Example />);

    const historyPanel = document.querySelector<HTMLElement>(
      '.paper-tabs__panel[data-value="history"]',
    );
    expect(historyPanel).toBeInTheDocument();
    expect(historyPanel).toHaveAttribute("hidden");

    await user.click(screen.getByRole("tab", { name: "Settings" }));
    expect(document.querySelector('.paper-tabs__panel[data-value="overview"]')).toHaveAttribute(
      "hidden",
    );

    rerender(
      <Tabs.Root defaultValue="one">
        <Tabs.List aria-label="Unmounting tabs">
          <Tabs.Tab value="one">One</Tabs.Tab>
          <Tabs.Tab value="two">Two</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="one">First</Tabs.Panel>
        <Tabs.Panel value="two" keepMounted={false}>
          Second
        </Tabs.Panel>
      </Tabs.Root>,
    );
    expect(document.querySelector('.paper-tabs__panel[data-value="two"]')).not.toBeInTheDocument();
  });

  it("forwards custom ids, classes, refs, and vertical orientation", () => {
    const tabRef = { current: null as HTMLButtonElement | null };
    render(
      <Tabs.Root orientation="vertical" className="custom-root" defaultValue="one">
        <Tabs.List className="custom-list" aria-label="Vertical tabs">
          <Tabs.Tab ref={tabRef} id="explicit-tab" className="custom-tab" value="one">
            One
          </Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel id="explicit-panel" className="custom-panel" value="one">
          First
        </Tabs.Panel>
      </Tabs.Root>,
    );

    expect(screen.getByRole("tablist")).toHaveAttribute("aria-orientation", "vertical");
    expect(screen.getByRole("tab")).toHaveClass("paper-tabs__tab", "custom-tab");
    expect(screen.getByRole("tabpanel")).toHaveClass("paper-tabs__panel", "custom-panel");
    expect(tabRef.current).toBe(screen.getByRole("tab"));
  });

  it("uses vertical arrow keys and skips disabled tabs in vertical orientation", async () => {
    const user = userEvent.setup();
    render(
      <Tabs.Root orientation="vertical" defaultValue="one">
        <Tabs.List aria-label="Vertical keyboard tabs">
          <Tabs.Tab value="one">One</Tabs.Tab>
          <Tabs.Tab value="two" disabled>
            Two
          </Tabs.Tab>
          <Tabs.Tab value="three">Three</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="one">First</Tabs.Panel>
        <Tabs.Panel value="two">Second</Tabs.Panel>
        <Tabs.Panel value="three">Third</Tabs.Panel>
      </Tabs.Root>,
    );

    await user.click(screen.getByRole("tab", { name: "One" }));
    await user.keyboard("{ArrowDown}");
    expect(screen.getByRole("tab", { name: "Three" })).toHaveFocus();
    await user.keyboard("{ArrowUp}");
    expect(screen.getByRole("tab", { name: "One" })).toHaveFocus();
  });
});
