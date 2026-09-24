# Composing Tadoku Paper

Paper uses three levels of composition. They describe ownership, not mandatory directory names.

| Level | Owns | Location |
| --- | --- | --- |
| Controls | Reusable appearance, semantics, and interaction without Tadoku business state | `frontend/packages/paper-ui/src/components/` |
| Product patterns | The content order, states, and actions of a recognizable Tadoku task section | Documented in the Paper catalogue; code starts beside its consuming screen |
| Screens | Routes, data fetching, permissions, form state, mutations, and arrangement of task sections | The consuming application |

Screens may import local pattern components and `paper-ui` controls. Pattern components may import `paper-ui`. `paper-ui` must not import an application, router, API client, or product service. An application passes data, links, and callbacks into any extracted pattern component; it still owns their effects.

## When to extract and share

Keep a section in its screen while it has one consumer and a clear local shape. Extract a named component when the section repeats, has meaningful independent states, or becomes hard to understand inside the screen. Name it for the user task or domain concept, such as `ReadingLogSummary`, rather than its layout.

A catalogue pattern describes a product convention; it does not require a React export. Publish a component from `paper-ui` only after the same stable appearance and interaction contract is needed by more than one consumer and its API does not encode application data, routing, permissions, or business rules. Keep product-specific shared code outside `paper-ui` until there is an actual cross-application consumer and an agreed owner.

Put a component in its own file when its API, examples, tests, or consumers warrant it. Small related helpers can share a file. Folder structure should follow responsibility and stay easy to navigate as the application grows.

## Documenting a pattern

For each catalogue pattern, record:

- The task it helps with, when to use it, and when to avoid it.
- Its anatomy and content order, including what each action does.
- Relevant loading, empty, error, unavailable, permission-limited, and success states. Show only states that the task can actually reach.
- Responsive, keyboard, and assistive-technology behavior.
- Which data and effects belong to the consuming screen, and whether the example is guidance or an exported component.
- The controls it composes and at least one consuming screen or journey.

Use realistic, deterministic catalogue fixtures to illustrate the contract. Verify application behavior through the real screen; a fixture alone cannot prove data flow or persistence. When reviewing a new Paper page, ask whether each section is a control, a documented pattern, or screen-owned composition, and whether its current owner matches its data and reuse needs.

The [logging pattern](../../../frontend/packages/paper-ui/src/catalog/phase-three-content.tsx) illustrates this boundary. Its reading-log summary uses `Surface` and `ButtonGroup` to show the work, amount, privacy, and contest state. A consuming logs screen owns the log data, navigation, and save or submission actions. The catalogue example is guidance; it is not a shared logging implementation.
