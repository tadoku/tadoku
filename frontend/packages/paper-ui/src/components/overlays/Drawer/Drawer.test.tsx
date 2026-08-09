import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { Drawer } from "./Drawer";

function ControlledDrawer({ onChange = vi.fn() }: { onChange?: (open: boolean) => void }) {
  const [open, setOpen] = useState(false);

  return (
    <Drawer
      open={open}
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen);
        onChange(nextOpen);
      }}
      trigger={<button type="button">Browse entries</button>}
      title="Reading entries"
      description="Choose an entry to review."
      closeLabel="Close entries"
    >
      <a href="#first">First entry</a>
      <button type="button">Last action</button>
    </Drawer>
  );
}

describe("Drawer", () => {
  it("supports controlled state, accessible naming, Escape dismissal, and focus restoration", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ControlledDrawer onChange={onChange} />);

    const trigger = screen.getByRole("button", { name: "Browse entries" });
    await user.click(trigger);

    const drawer = screen.getByRole("dialog", { name: "Reading entries" });
    expect(drawer).toHaveAccessibleDescription("Choose an entry to review.");
    expect(drawer).toHaveAttribute("data-placement", "end");
    expect(onChange).toHaveBeenLastCalledWith(true);

    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(onChange).toHaveBeenLastCalledWith(false);
    expect(trigger).toHaveFocus();
  });

  it("contains focus, exposes a labeled close control, and dismisses from the backdrop", async () => {
    const user = userEvent.setup();
    render(<ControlledDrawer />);

    await user.click(screen.getByRole("button", { name: "Browse entries" }));
    const close = screen.getByRole("button", { name: "Close entries" });
    expect(close.querySelector("svg")).not.toBeNull();

    expect(document.querySelectorAll("[data-base-ui-focus-guard]")).toHaveLength(2);
    expect(
      Array.from(document.querySelectorAll("[data-base-ui-inert]")).length,
    ).toBeGreaterThan(0);

    await user.click(document.querySelector<HTMLElement>(".paper-drawer__backdrop")!);
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("supports start placement and keeps its portal in the trigger owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <Drawer
        defaultOpen={false}
        placement="start"
        trigger={<button type="button">Open filters</button>}
        title="Filters"
      >
        <button type="button">Apply filters</button>
      </Drawer>,
      { container, baseElement: frameDocument.body },
    );

    await user.click(
      within(frameDocument.body).getByRole("button", { name: "Open filters" }),
    );
    const drawer = within(frameDocument.body).getByRole("dialog", {
      name: "Filters",
    });
    expect(drawer).toHaveAttribute("data-placement", "start");
    expect(document.body.querySelector('[role="dialog"]')).toBeNull();

    unmount();
    frame.remove();
  });
});
