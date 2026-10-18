import Example32 from "./examples/flash.statuses";
import Example32Source from "./examples/flash.statuses.tsx?raw";
import Example33 from "./examples/loading.entries";
import Example33Source from "./examples/loading.entries.tsx?raw";
import Example31 from "./examples/toast.entry-saved";
import Example31Source from "./examples/toast.entry-saved.tsx?raw";
import { defineCatalogFixture } from "./schema";
import type { ComponentDetails, DocSpec } from "./forms-feedback-spec";

export const feedbackDetails: ComponentDetails = {
  "component.flash": {
    example: "Compare all four statuses with identical anatomy. The warning includes a real in-preview recovery link; title and text explain the status without relying on the rail color.",
    anatomy: "A colored square rail, optional decorative icon, title/body, and optional action form one persistent message.", variants: "Information, success, warning, and danger encode distinct outcomes.", states: "visible=false unmounts the message; caller-owned actions keep native semantics.", content: "Name the issue and its recovery.",
    implementation: "Import Flash from paper-ui. Pass body content as children and a real Paper button or link as action. Visibility is caller-controlled; Flash does not add its own dismiss state.",
    props: [{ name: "variant", type: '"information" | "success" | "warning" | "danger"', defaultValue: '"information"', description: "Danger uses role=alert; the other variants use role=status." }, { name: "title", type: "string", description: "Short outcome or problem statement above the body." }, { name: "children", type: "ReactNode", description: "Consequence and recovery copy." }, { name: "action", type: "ReactNode", description: "Caller-owned link or button; no action behavior is inferred." }, { name: "icon", type: "ReactNode", description: "Optional decorative icon, hidden from assistive technology." }, { name: "visible", type: "boolean", defaultValue: "true", description: "False returns null and removes the message from the document." }],
  },
  "component.loading": {
    example: "Compare all three spinner sizes with a visible description of the pending reading operation. The accessible label remains present even with reduced motion enabled.",
    anatomy: "A decorative spinner sits inside a role=status wrapper with a visually hidden label. Visible loading copy must be supplied by the application.",
    variants: "small fits inline operations; default fits a section; large gives a pending page or region stronger emphasis. Size does not indicate percentage progress.",
    states: "Indeterminate only. Reduced motion stops the animation while retaining the shape and label. Unmount Loading when content or an error replaces the pending state.",
    content: "Name what is loading: Loading reading entries. For long waits, also show visible explanatory text so sighted users know which operation is pending.",
    implementation: "Import Loading from paper-ui. Render conditionally while waiting; mark the region being updated aria-busy when appropriate. Keep one useful status announcement per operation.",
    props: [{ name: "label", type: "string", defaultValue: '"Loading"', description: "Visually hidden accessible status text; supply visible copy alongside the spinner when needed." }, { name: "size", type: '"small" | "default" | "large"', defaultValue: '"default"', description: "Visual spinner size, independent of the loading state." }],
  },
  "component.toast": {
    example: "Trigger a confirmation or failure, add several notifications to inspect the queue, then dismiss one or all. This teaching preview uses timeout=0; choose a finite timeout in application setup unless persistence is intentional.",
    anatomy: "One provider owns the queue and viewport; each notification contains a title, optional description/action, and named dismiss button.",
    variants: "low priority announces politely; high interrupts and belongs to urgent failures. The type field is metadata and does not currently select a Paper color variant.",
    states: "ToastProvider defaults follow Base UI: a five-second timeout and three visible toasts. timeout=0 disables expiry. Hover/focus pauses expiry; closing removes a notification through the manager. Keep durable recovery beside the affected content.",
    content: "Use a completed outcome for confirmation, such as Entry saved. An error should explain what remains safe and where to retry. Do not put the only copy of a critical recovery action in a timed message.",
    implementation: "Mount ToastProvider once above components calling useToast. The hook must run inside that provider, not in the same component that creates it. Use add for a new message, update for an existing id, close for dismissal, and promise for loading/success/error transitions. ToastContainer is a migration alias for the provider.",
    props: [{ name: "ToastProvider.timeout", type: "number", defaultValue: "5000", description: "Milliseconds before dismissal; 0 disables automatic dismissal." }, { name: "ToastProvider.limit", type: "number", defaultValue: "3", description: "Maximum notifications visible at once." }, { name: "add(options)", type: "(ToastOptions) => string", description: "Returns an id. Options include title, description, priority, timeout, actionProps, onClose, and onRemove." }, { name: "close(id?)", type: "(string?) => void", description: "Dismiss one notification, or all when the id is omitted." }, { name: "update(id, options)", type: "(string, ToastUpdateOptions) => void", description: "Change an existing notification without adding another." }, { name: "promise(promise, options)", type: "<Value>(Promise<Value>, ToastPromiseOptions<Value>) => Promise<Value>", description: "Supply loading, success, and error messages or callbacks for one asynchronous operation." }],
  },
};

export const feedbackCommonMistakes: Record<string, string> = {
  "component.flash": "Do not use a danger rail for neutral context; danger also interrupts assistive technology with an alert.",
  "component.loading": "Do not assume label is visible: it is visually hidden. Supply visible waiting copy when users need more than a spinner.",
  "component.toast": "Do not call useToast above its provider or duplicate providers for every action. Mount one provider around the application area."
};

export const flashFixtures = [
  defineCatalogFixture({ ...{
  "id": "flash.statuses",
  "name": "Feedback statuses",
  "description": "Compare the supported presentations with consistent reading-log content.",
  "tags": [
    "flash",
    "statuses"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example32Source, render: () => <Example32 /> }),
] as const;

export const flashSpec: DocSpec = { id: "component.flash", route: "/components/feedback/flash", name: "Flash", category: "feedback", summary: "Places persistent contextual feedback beside the content it describes.", fixtureId: "flash.statuses", sourcePath: "src/components/feedback/feedback.tsx", react: ["Flash"], types: ["FlashProps", "FlashVariant"], when: "Use for page or section feedback that should remain until the context changes, such as a saved result, a contest warning, or a failed page-level operation.", avoid: "Use Toast for brief asynchronous confirmation and an inline field error when the user must repair one value. Do not put required recovery only in a Flash that may disappear with its surrounding content.", choose: "Flash stays beside the content it explains. Toast is transient and global; field errors identify a specific control and its recovery.", variants: "Information covers neutral context, success confirms a completed outcome, and warning calls for review before proceeding. Danger is reserved for an urgent failure and changes the live-region role from status to alert.", states: "Setting visible=false removes the message from the document. Actions are caller-owned links or buttons, so their label, disabled state, and navigation behavior must remain explicit.", behavior: "Danger feedback is announced as an alert; information, success, and warning use a polite status announcement. Changing visible to false unmounts the message instead of merely hiding it.", content: "Write the title as the outcome or issue, then use the body for consequence and recovery. Keep any action label specific\u2014Review fields or Try again instead of OK.", accessibility: "Pair the colored edge with a title and message text, because color does not name the status. Provide actions as real links or buttons and reserve danger for messages that warrant an interrupting alert.", migration: "Map error to danger, remove Next Link ownership, and pass any action as router-owned React content.", pageSections: ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"] };

export const loadingFixtures = [
  defineCatalogFixture({ ...{
  "id": "loading.entries",
  "name": "Loading sizes and visible context",
  "description": "Compare the supported presentations with consistent reading-log content.",
  "tags": [
    "loading",
    "entries"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example33Source, render: () => <Example33 /> }),
] as const;

export const loadingSpec: DocSpec = { id: "component.loading", route: "/components/feedback/loading", name: "Loading", category: "feedback", summary: "Communicates indeterminate progress with a named, reduced-motion-safe status.", fixtureId: "loading.entries", sourcePath: "src/components/feedback/feedback.tsx", react: ["Loading"], types: ["LoadingProps"], when: "Use while a bounded region is waiting and no meaningful progress value exists.", avoid: "Use native progress for measurable work and retain existing content when optimistic updates are safer.", choose: "Loading indicates indeterminate work; Button loading prevents repeat action while keeping its name.", behavior: "The status label is available to assistive technology and animation stops under reduced motion.", accessibility: "Supply a label that names what is loading and do not rely on spinner motion alone.", migration: "Replace legacy Loading and remove application-owned spinner SVGs." };

export const toastFixtures = [
  defineCatalogFixture({ ...{
  "id": "toast.entry-saved",
  "name": "Queued notifications",
  "description": "Run the actions and inspect their visible result.",
  "tags": [
    "toast",
    "entry",
    "saved"
  ],
  "themes": [
    "light",
    "dark"
  ],
  "densities": [
    "comfortable",
    "compact"
  ],
  "viewports": [
    {
      "id": "phone",
      "label": "Phone",
      "width": 360,
      "height": 720
    },
    {
      "id": "tablet",
      "label": "Tablet",
      "width": 768,
      "height": 800
    },
    {
      "id": "desktop",
      "label": "Desktop",
      "width": 1280,
      "height": 800
    }
  ],
  "deterministic": true
}, code: Example31Source, render: () => <Example31 /> }),
] as const;

export const toastSpec: DocSpec = { id: "component.toast", route: "/components/feedback/toast", name: "Toast", category: "feedback", summary: "Queues brief asynchronous notifications with polite/urgent priority and explicit dismissal.", fixtureId: "toast.entry-saved", sourcePath: "src/components/feedback/feedback.tsx", react: ["ToastProvider", "useToast", "ToastContainer"], types: ["ToastProviderProps"], when: "Use for brief confirmation or failure that is not tied to one visible field or section.", avoid: "Use Flash for persistent contextual feedback and never put required recovery only in an expiring toast.", choose: "Toast is transient global feedback; Flash remains in context until it is no longer relevant.", behavior: "Base UI owns queueing, live announcements, timeout pause, swipe, and stacking behind the Paper provider.", accessibility: "Choose low or high priority intentionally, keep a visible dismiss action, and repeat critical recovery in context.", migration: "Replace react-toastify setup with ToastProvider and useToast; ToastContainer remains a migration alias." };
