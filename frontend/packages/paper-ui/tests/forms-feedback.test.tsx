import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { phaseThreeFormsFeedbackFixtures } from "../src/catalog/phase-three-forms-feedback";
import {
  Button,
  ButtonGroup,
  Flash,
  Loading,
  Surface,
  ToastProvider,
  surfaceClassName,
  useToast,
} from "../src";

describe("form usage examples", () => {
  it.each(["textarea.reading-notes", "select.language", "radio-select.viewport", "radio-group.format", "amount.progress", "autocomplete.language", "multi-autocomplete.languages", "tags.entry"])("demonstrates required validation in %s.empty", async (id) => {
    const fixture = phaseThreeFormsFeedbackFixtures.find((entry) => entry.id === `${id}.empty`);
    const user = userEvent.setup();
    render(<>{fixture!.render()}</>);
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("alert")).not.toBeEmptyDOMElement();
    expect(screen.queryByText("Entry saved")).not.toBeInTheDocument();
  });

  it("saves the submitted values and resets the example", async () => {
    const fixture = phaseThreeFormsFeedbackFixtures.find((entry) => entry.id === "textarea.reading-notes.empty");
    const user = userEvent.setup();
    render(<>{fixture!.render()}</>);
    const input = screen.getByRole("textbox", { name: /Reading notes/u });
    await user.type(input, "Finished chapter three.");
    await user.click(screen.getByRole("button", { name: "Save entry" }));
    expect(await screen.findByRole("status")).toHaveTextContent("Finished chapter three.");
    await user.clear(input);
    expect(screen.getByRole("status")).toHaveTextContent("Finished chapter three.");
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(input).toHaveValue("");
  });
});

function ToastTrigger() {
  const toast = useToast();
  return <Button onClick={() => toast.add({ title: "Entry saved", description: "12 pages added." })}>Show notification</Button>;
}

describe("feedback and action compositions", () => {
  it("uses urgent and polite live-region semantics", () => {
    render(<><Flash title="Saved">Entry updated.</Flash><Flash variant="danger">Could not delete entry.</Flash><Loading label="Loading entries" /></>);
    expect(screen.getAllByRole("status")).toHaveLength(2);
    expect(screen.getByRole("alert")).toHaveTextContent("Could not delete entry.");
    expect(screen.getByText("Loading entries")).toHaveClass("paper-visually-hidden");
  });

  it("preserves links and buttons inside one labelled ButtonGroup", async () => {
    const user = userEvent.setup();
    const selected = vi.fn();
    render(<ButtonGroup label="Log actions" actions={[{ id: "view", label: "View", href: "/logs/1" }, { id: "delete", label: "Delete", variant: "destructive", onSelect: selected }]} />);
    expect(screen.getByRole("group", { name: "Log actions" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "View" })).toHaveAttribute("href", "/logs/1");
    await user.click(screen.getByRole("button", { name: "Delete" }));
    expect(selected).toHaveBeenCalledOnce();
  });

  it("shares surface recipes and renders queued Base UI toasts", async () => {
    const user = userEvent.setup();
    render(<ToastProvider timeout={0}><Surface elevation="floating" accent>Reading summary</Surface><ToastTrigger /></ToastProvider>);
    expect(screen.getByText("Reading summary")).toHaveClass(surfaceClassName({ elevation: "floating", accent: true }));
    await user.click(screen.getByRole("button", { name: "Show notification" }));
    expect(await screen.findByText("Entry saved")).toBeInTheDocument();
    const dismiss = screen.getByRole("button", { name: "Dismiss notification" });
    expect(dismiss.querySelector("svg")).not.toBeNull();
    expect(dismiss).not.toHaveTextContent("×");
  });

  it("keeps its toast viewport in the provider owner document", async () => {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const frameDocument = frame.contentDocument!;
    const container = frameDocument.createElement("div");
    frameDocument.body.append(container);
    const user = userEvent.setup({ document: frameDocument });
    const { unmount } = render(
      <ToastProvider timeout={0}><ToastTrigger /></ToastProvider>,
      { container, baseElement: frameDocument.body },
    );

    await user.click(within(frameDocument.body).getByRole("button", { name: "Show notification" }));
    expect(await within(frameDocument.body).findByText("Entry saved")).toBeInTheDocument();
    expect(document.body).not.toHaveTextContent("Entry saved");
    unmount();
    frame.remove();
  });
});
