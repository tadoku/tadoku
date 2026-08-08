import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, useRef, useState, type Ref } from "react";
import { describe, expect, it, vi } from "vitest";
import { Modal } from "../src";

function ComposableModal({
  onOpenChange = vi.fn(),
  triggerRef,
}: {
  onOpenChange?: (open: boolean) => void;
  triggerRef?: Ref<HTMLButtonElement>;
}) {
  const [open, setOpen] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <Modal
      trigger={
        <button
          ref={triggerRef}
          className="application-search-trigger"
          type="button"
        >
          <span>Search Paper</span>
          <kbd>Ctrl K</kbd>
        </button>
      }
      title="Search Paper"
      description="Search the Paper catalogue."
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
        setOpen(nextOpen);
      }}
      initialFocus={inputRef}
      footer={null}
    >
      <label>
        Search documents
        <input ref={inputRef} type="search" />
      </label>
    </Modal>
  );
}

describe("Modal composition", () => {
  it("composes an application trigger and supports controlled state", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    const triggerRef = createRef<HTMLButtonElement>();
    render(
      <ComposableModal
        onOpenChange={onOpenChange}
        triggerRef={triggerRef}
      />,
    );

    const trigger = screen.getByRole("button", { name: /Search Paper/u });
    expect(trigger).toHaveClass("application-search-trigger");
    expect(trigger).not.toHaveClass("paper-button");
    expect(trigger).toHaveAttribute("aria-haspopup", "dialog");
    expect(triggerRef.current).toBe(trigger);

    await user.click(trigger);
    expect(onOpenChange).toHaveBeenLastCalledWith(true);
    expect(screen.getByRole("dialog", { name: "Search Paper" })).toHaveAccessibleDescription(
      "Search the Paper catalogue.",
    );
  });

  it("keeps Base UI focus and dismissal behavior in a footerless composition", async () => {
    const user = userEvent.setup();
    render(<ComposableModal />);

    const trigger = screen.getByRole("button", { name: /Search Paper/u });
    await user.click(trigger);

    await waitFor(() => {
      expect(screen.getByRole("searchbox", { name: "Search documents" })).toHaveFocus();
    });
    expect(document.querySelector(".paper-modal__footer")).toBeNull();
    const closeButton = screen.getByRole("button", { name: "Close" });
    expect(closeButton).toBeInTheDocument();

    await user.tab();
    await waitFor(() => expect(closeButton).toHaveFocus());

    const backdrop = document.querySelector<HTMLElement>(".paper-modal__backdrop");
    expect(backdrop).not.toBeNull();
    await user.click(backdrop!);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(trigger);
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });
});
