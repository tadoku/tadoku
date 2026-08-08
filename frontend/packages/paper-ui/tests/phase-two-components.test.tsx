import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import {
  ActionMenu,
  Button,
  Input,
  Modal,
  buttonClassName,
} from "../src";

function InputForm({ onSubmit = vi.fn() }: { onSubmit?: (value: unknown) => void }) {
  const methods = useForm<{ title: string }>({ defaultValues: { title: "" } });
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={methods.handleSubmit(onSubmit)}>
        <Input
          name="title"
          label="Log title"
          hint="Use a recognizable name."
          rules={{ required: "Enter a log title." }}
          required
        />
        <Button type="submit">Save log</Button>
      </form>
    </FormProvider>
  );
}

describe("Button", () => {
  it("defaults to a non-submitting native button and shares its recipe", async () => {
    const user = userEvent.setup();
    const submit = vi.fn((event: React.FormEvent) => event.preventDefault());
    render(
      <form onSubmit={submit}>
        <Button variant="outline">Review log</Button>
        <a className={buttonClassName({ variant: "outline" })} href="#logs">
          View logs
        </a>
      </form>,
    );

    const button = screen.getByRole("button", { name: "Review log" });
    expect(button).toHaveAttribute("type", "button");
    expect(button.className).toBe(buttonClassName({ variant: "outline" }));
    expect(screen.getByRole("link", { name: "View logs" })).not.toHaveAttribute(
      "role",
      "button",
    );
    await user.click(button);
    expect(submit).not.toHaveBeenCalled();
  });

  it("keeps its name, reports busy, and blocks repeat activation while loading", async () => {
    const user = userEvent.setup();
    const activate = vi.fn();
    render(
      <Button loading loadingLabel="Saving reading log" onClick={activate}>
        Save log
      </Button>,
    );

    const button = screen.getByRole("button", { name: /Save log/u });
    expect(button).toHaveAccessibleName("Save log");
    expect(button).toHaveAccessibleDescription("Saving reading log");
    expect(button).toHaveAttribute("aria-busy", "true");
    expect(button).toBeDisabled();
    await user.click(button);
    expect(activate).not.toHaveBeenCalled();
  });
});

describe("Input", () => {
  it("associates label and hint, then submits through React Hook Form", async () => {
    const user = userEvent.setup();
    const submit = vi.fn();
    render(<InputForm onSubmit={submit} />);

    const input = screen.getByRole("textbox", { name: "Log title" });
    expect(input).toHaveAccessibleDescription("Use a recognizable name.");
    await user.type(input, "August reading");
    await user.click(screen.getByRole("button", { name: "Save log" }));
    expect(submit).toHaveBeenCalledWith(
      { title: "August reading" },
      expect.anything(),
    );
  });

  it("associates a recoverable validation error", async () => {
    const user = userEvent.setup();
    render(<InputForm />);
    await user.click(screen.getByRole("button", { name: "Save log" }));

    await screen.findByRole("alert");
    const input = screen.getByRole("textbox", { name: "Log title" });
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAccessibleDescription(
      "Use a recognizable name. Enter a log title.",
    );
    expect(screen.getByRole("alert")).toHaveTextContent("Enter a log title.");
  });
});

describe("Modal", () => {
  it("opens with dialog semantics, closes with Escape, and returns focus", async () => {
    const user = userEvent.setup();
    render(
      <Modal
        triggerLabel="Review deletion"
        title="Delete this log?"
        description="This cannot be undone."
      >
        <p>August reading log</p>
      </Modal>,
    );

    const trigger = screen.getByRole("button", { name: "Review deletion" });
    await user.click(trigger);
    expect(screen.getByRole("dialog", { name: "Delete this log?" })).toHaveAccessibleDescription(
      "This cannot be undone.",
    );
    const iconClose = document.querySelector(".paper-modal__icon-close");
    expect(iconClose?.querySelector("svg")).not.toBeNull();
    expect(iconClose).not.toHaveTextContent("×");
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("runs an explicit primary action and closes", async () => {
    const user = userEvent.setup();
    const remove = vi.fn();
    render(
      <Modal
        triggerLabel="Review deletion"
        title="Delete this log?"
        closeLabel="Keep log"
        action={{ label: "Delete log", variant: "destructive", onAction: remove }}
      >
        <p>August reading log</p>
      </Modal>,
    );

    await user.click(screen.getByRole("button", { name: "Review deletion" }));
    await user.click(screen.getByRole("button", { name: "Delete log" }));
    expect(remove).toHaveBeenCalledOnce();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("keeps its portal in the trigger owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <Modal triggerLabel="Open details" title="Log details">
        <p>Japanese · 120 pages</p>
      </Modal>,
      { container, baseElement: frameDocument.body },
    );

    await user.click(within(frameDocument.body).getByRole("button", { name: "Open details" }));
    expect(within(frameDocument.body).getByRole("dialog", { name: "Log details" })).toBeInTheDocument();
    expect(document.body.querySelector('[role="dialog"]')).toBeNull();
    unmount();
    frame.remove();
  });
});

describe("ActionMenu", () => {
  it("supports keyboard selection and skips disabled items", async () => {
    const user = userEvent.setup();
    const edit = vi.fn();
    const duplicate = vi.fn();
    render(
      <ActionMenu
        label="Log actions"
        items={[
          { id: "edit", label: "Edit log", onSelect: edit },
          {
            id: "duplicate",
            label: "Duplicate log",
            disabled: true,
            onSelect: duplicate,
          },
        ]}
      />,
    );

    const trigger = screen.getByRole("button", { name: "Log actions" });
    trigger.focus();
    await user.keyboard("{Enter}");
    const editItem = await screen.findByRole("menuitem", { name: "Edit log" });
    expect(screen.getByRole("menuitem", { name: "Duplicate log" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(editItem).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(edit).toHaveBeenCalledOnce();
    expect(duplicate).not.toHaveBeenCalled();
    expect(trigger).toHaveFocus();
  });

  it("keeps its floating menu in the trigger owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <ActionMenu
        label="Entry actions"
        items={[{ id: "edit", label: "Edit entry", onSelect: vi.fn() }]}
      />,
      { container, baseElement: frameDocument.body },
    );

    await user.click(within(frameDocument.body).getByRole("button", { name: "Entry actions" }));
    expect(await within(frameDocument.body).findByRole("menu")).toBeInTheDocument();
    expect(document.body.querySelector('[role="menu"]')).toBeNull();
    unmount();
    frame.remove();
  });
});
