import Example46 from "./examples/experiment.logging-v2-entry";
import Example46Source from "./examples/experiment.logging-v2-entry.tsx?raw";
import Example45 from "./examples/pattern.logging-summary";
import Example45Source from "./examples/pattern.logging-summary.tsx?raw";
import { guidanceDocument, PACKAGE_VERSION, REVIEW_DATE } from "./guidance-document";
import { defineCatalogDocument, defineCatalogFixture, type CatalogRedirect } from "./schema";
import { visualFoundationDocuments } from "./visual-foundations";

const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

export const phaseThreeFoundationDocuments = [
  guidanceDocument({
    id: "foundation.principles", route: "/foundations/principles", name: "Principles", kind: "foundation",
    summary: "The product and design principles that keep Tadoku calm, legible, and recognizably about reading.",
    keywords: ["principles", "bookplate", "calm", "reading"], sourcePath: "src/catalog/phase-three-content.tsx",
    whenToUse: ["Use these principles to resolve component, content, and product-pattern trade-offs."],
    content: ["Paper should feel editorial rather than ornamental: strong hierarchy, quiet structure, and realistic Tadoku language.", "Prefer semantic native behavior, reveal state without relying on color, and introduce complexity only when it helps a reading task."],
    requirements: ["Legibility, keyboard access, and recognizable state outrank decorative novelty."],
    publicContract: ["The decision log and accepted ADRs are the source of truth."],
  }),
  ...visualFoundationDocuments,
] as const;

export const phaseThreeContentFixtures = [
  defineCatalogFixture({
    id: "pattern.logging-summary", name: "Reading-log summary", description: "A realistic summary with explicit navigation and continuation actions.",
    tags: ["logging", "pattern", "reading"], themes: ["light", "dark"], densities: ["comfortable", "compact"], viewports: VIEWPORTS,
    deterministic: true,
    code: Example45Source,
    render: () => <Example45 />,
  }),
  defineCatalogFixture({
    id: "experiment.logging-v2-entry", name: "Logging v2 entry", description: "A deterministic experiment separating saved reading from contest submission.",
    tags: ["logging", "experiment", "form"], themes: ["light", "dark"], densities: ["comfortable", "compact"], viewports: VIEWPORTS,
    deterministic: true,
    code: Example46Source,
    render: () => <Example46 />,
  }),
] as const;

export const phaseThreePatternDocuments = [
  defineCatalogDocument({
    id: "pattern.logging", route: "/patterns/logging", name: "Logging", kind: "pattern", category: "patterns", aliases: [],
    summary: "Composes Paper primitives into reading summaries, entry actions, and transparent contest state.",
    keywords: ["logging", "reading", "contest", "summary"], lifecycle: "Stable", reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/phase-three-content.tsx", packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: ["Use for product flows that record reading and explain where it contributes."],
      whenNotToUse: ["Do not turn the product flow into a reusable low-level component."],
      content: ["Name the work, amount, language, privacy, and contest state in plain Tadoku language.", "Compose the summary from Paper controls. Keep log data, routes, and save or submission actions in the consuming screen; this example documents a pattern, not an exported component."],
      commonMistakes: ["Hiding whether an entry is merely saved or also submitted to a contest.", "Treating the catalogue fixture's local state as application behavior."],
    },
    accessibility: {
      requirements: ["Keep summaries in document order and expose every action as a named link or button."],
      keyboard: ["All actions follow native link and button operation."], knownConstraints: [],
    },
    api: { react: ["Surface", "ButtonGroup", "Flash"], cssClasses: [], publicTypes: [], defaults: ["Applications own routing, data, and mutation state."], invalidCombinations: ["Patterns must not import application routers into paper-ui.", "The catalogue example is not a shared logging component."] },
    fixtureIds: ["pattern.logging-summary"], dependencies: { documents: ["component.surface", "component.button-group"], packages: ["paper-ui"] },
    migration: { legacy: ["styleguide/pages/logging.tsx"], notes: ["Preserve the realistic information hierarchy, not the legacy local-state implementation."] },
    changelog: [{ date: REVIEW_DATE, note: "Rebuilt logging as a product pattern." }], behaviorTestIds: ["pattern.logging.render"],
  }),
] as const;

export const phaseThreeExperimentDocuments = [
  defineCatalogDocument({
    id: "experiment.logging-v2", route: "/experiments/logging-v2", name: "Logging v2", kind: "experiment", category: "experiments", aliases: [],
    summary: "Explores a clearer split between saving reading, reviewing details, and submitting to contests.",
    keywords: ["logging", "experiment", "contest", "form"], lifecycle: "Experimental", reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/phase-three-content.tsx", packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: ["Use only for evaluating the proposed logging flow."],
      whenNotToUse: ["Do not treat this experiment as a Stable product or component contract."],
      content: ["Keep saved, reviewed, and submitted states explicit in labels and confirmation copy."],
      commonMistakes: ["Promoting experimental layout or copy into applications without product review."],
    },
    accessibility: { requirements: ["Retain visible labels, error relationships, and explicit confirmation state."], keyboard: ["Use native form control and button behavior."], knownConstraints: ["Product validation is not complete."] },
    api: { react: ["Flash", "Button"], cssClasses: ["paper-field", "paper-input"], publicTypes: [], defaults: ["The application owns TanStack Form and submission state."], invalidCombinations: ["Do not expose the experiment as a paper-ui component."] },
    fixtureIds: ["experiment.logging-v2-entry"], dependencies: { documents: ["pattern.logging"], packages: ["paper-ui", "@tanstack/react-form"] },
    migration: { legacy: ["styleguide/pages/logging-v2.tsx"], notes: ["Keep this visibly Experimental until a product decision promotes or removes it."] },
    changelog: [{ date: REVIEW_DATE, note: "Preserved logging v2 as a deterministic experiment." }], behaviorTestIds: ["experiment.logging-v2.render"],
  }),
] as const;

export const phaseThreeGovernanceDocuments = [
  guidanceDocument({
    id: "governance.contributing", route: "/contributing", name: "Contributing", kind: "governance",
    summary: "How Paper changes are proposed, implemented, tested, documented, reviewed, and promoted.",
    keywords: ["contributing", "lifecycle", "review", "deprecation"], sourcePath: "src/catalog/phase-three-content.tsx",
    whenToUse: ["Use before adding an export, token, fixture, route, or lifecycle change."],
    content: ["Every change has a canonical source, deterministic fixture, behavior evidence, migration note, and review date before Stable promotion.", "Experimental contracts may change. Stable contracts require migration-safe changes. Deprecated contracts name a registered replacement, removal timing, and exit path.", "Design history includes the refinement audit, visual studies, decision log, implementation plan, research packet, and accepted ADRs, archived in Git history and linked from the catalogue design history."],
    requirements: ["Review semantics, keyboard behavior, reflow, both themes, both densities, and realistic content before promotion."],
    publicContract: ["The catalogue registry is the source of lifecycle, route, fixture, source, and migration metadata."],
  }),
  guidanceDocument({
    id: "governance.changelog", route: "/changelog", name: "Changelog", kind: "governance",
    summary: "Meaningful package, component, documentation, deployment, and migration changes.",
    keywords: ["changelog", "release", "version", "migration"], sourcePath: "src/catalog/phase-three-content.tsx",
    whenToUse: ["Consult before upgrading Paper or changing an application integration."],
    content: ["0.1.0 establishes semantic foundations, the complete catalogue contract, static delivery at paper.tadoku.app, framework-neutral components, and TypeScript 5.4+-compatible built declarations.", "Current review (unreleased) adds usable foundation examples, a spacing scale and layout utilities, corrects accent-rail perimeter alignment, and adds Experimental page-composition controls (PageHeader, SectionHeading, StatusTag, EmptyState, StatGrid, DescriptionList and AppliedFilters) with the Filter bar pattern. Documentation coverage is not an application migration or release approval."],
    requirements: ["Record behavior, accessibility, API, or migration consequences instead of listing filenames alone."],
    publicContract: ["Release identity is the merged Tadoku commit plus immutable container digest recorded at each deployment gate."],
  }),
] as const;

/** Exact legacy paths; the root route is already canonical and needs no redirect. */
export const legacyStyleguideRedirects = [
  { from: "/color", to: "/foundations/color" },
  { from: "/typography", to: "/foundations/typography" },
  { from: "/branding", to: "/foundations/brand" },
  { from: "/templates", to: "/foundations/layout" },
  { from: "/forms", to: "/components/forms/input" },
  { from: "/buttons", to: "/components/actions/button" },
  { from: "/navigation", to: "/components/navigation/navbar" },
  { from: "/toasts", to: "/components/feedback/toast" },
  { from: "/flash", to: "/components/feedback/flash" },
  { from: "/charts", to: "/components/data-display/heatmap-chart" },
  { from: "/modals", to: "/components/overlays/modal" },
  { from: "/tables", to: "/components/data-display/table" },
  { from: "/breadcrumb", to: "/components/navigation/breadcrumb" },
  { from: "/action-menu", to: "/components/actions/action-menu" },
  { from: "/pagination", to: "/components/navigation/pagination" },
  { from: "/logging", to: "/patterns/logging" },
  { from: "/logging-v2", to: "/experiments/logging-v2" },
] as const satisfies readonly CatalogRedirect[];
