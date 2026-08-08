import { FormProvider, useForm } from "react-hook-form";
import { Button, ButtonGroup, Flash, Input, Surface } from "../index";
import {
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type CatalogRedirect,
} from "./schema";

const REVIEW_DATE = "2026-08-08";
const PACKAGE_VERSION = "0.1.0";
const OWNER = "Tadoku design systems";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;

interface GuidanceDocumentOptions {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly kind: "foundation" | "governance";
  readonly summary: string;
  readonly keywords: readonly string[];
  readonly sourcePath: string;
  readonly whenToUse: readonly string[];
  readonly content: readonly string[];
  readonly requirements: readonly string[];
  readonly publicContract: readonly string[];
}

function guidanceDocument(options: GuidanceDocumentOptions): CatalogDocument {
  return defineCatalogDocument({
    ...options,
    category: options.kind === "foundation" ? "foundations" : "governance",
    aliases: [],
    lifecycle: "Experimental",
    owner: OWNER,
    reviewDate: REVIEW_DATE,
    packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: options.whenToUse,
      whenNotToUse: ["Do not copy legacy presentation values when a semantic Paper role exists."],
      content: options.content,
      commonMistakes: ["Treating a visual sample as a new token or component contract."],
    },
    accessibility: {
      requirements: options.requirements,
      keyboard: [],
      knownConstraints: ["The comprehensive WCAG and forced-colors audit remains tracked separately."],
    },
    api: {
      react: [],
      cssClasses: [],
      publicTypes: [],
      defaults: options.publicContract,
      invalidCombinations: ["Private palette values and package source paths are not consumer APIs."],
    },
    fixtureIds: [],
    dependencies: { documents: [], packages: ["paper-ui"] },
    migration: {
      legacy: [],
      notes: ["Apply this contract when an application completes its one-time Paper migration."],
    },
    changelog: [{ date: REVIEW_DATE, note: "Rebuilt from canonical Paper sources for the complete catalogue." }],
    behaviorTestIds: [],
  });
}

export const phaseThreeFoundationDocuments = [
  guidanceDocument({
    id: "foundation.principles", route: "/foundations/principles", name: "Principles", kind: "foundation",
    summary: "The product and design principles that keep Tadoku calm, legible, and recognizably about reading.",
    keywords: ["principles", "bookplate", "calm", "reading"], sourcePath: "docs/wip/tadoku-paper/decision-log.md",
    whenToUse: ["Use these principles to resolve component, content, and product-pattern trade-offs."],
    content: ["Paper should feel editorial rather than ornamental: strong hierarchy, quiet structure, and realistic Tadoku language.", "Prefer semantic native behavior, reveal state without relying on color, and introduce complexity only when it helps a reading task."],
    requirements: ["Legibility, keyboard access, and recognizable state outrank decorative novelty."],
    publicContract: ["The decision log and accepted ADRs are the source of truth."],
  }),
  guidanceDocument({
    id: "foundation.color", route: "/foundations/color", name: "Color", kind: "foundation",
    summary: "Semantic surface, text, action, status, rule, focus, and chart roles for light and dark themes.",
    keywords: ["color", "theme", "tokens", "contrast"], sourcePath: "src/foundations/tokens.css",
    whenToUse: ["Choose a semantic role for every product color."],
    content: ["Warm paper neutrals carry structure while ink violet is reserved for action, focus, and small annotation-like accents.", "Status and chart colors always need text, shape, pattern, or position as a second cue."],
    requirements: ["Use semantic aliases and preserve text/action contrast in both themes."],
    publicContract: ["Public custom properties use the --paper-color-* families."],
  }),
  guidanceDocument({
    id: "foundation.typography", route: "/foundations/typography", name: "Typography", kind: "foundation",
    summary: "Merriweather editorial hierarchy with Open Sans for interface copy and dense controls.",
    keywords: ["typography", "fonts", "hierarchy", "prose"], sourcePath: "src/foundations/fonts.css",
    whenToUse: ["Use named type roles instead of choosing ad hoc sizes or weights."],
    content: ["Merriweather is limited to display, page, and section hierarchy. Open Sans carries component titles, labels, metadata, and body copy.", "Self-hosted 400, 600, and 700 weights keep emphasis honest and delivery framework-independent."],
    requirements: ["Respect zoom, user font settings, and readable line lengths."],
    publicContract: ["Use paper-type-display, paper-type-page, paper-type-section, paper-type-component, paper-type-label, and paper-type-metadata."],
  }),
  guidanceDocument({
    id: "foundation.spacing-and-density", route: "/foundations/spacing-and-density", name: "Spacing and density", kind: "foundation",
    summary: "A shared spacing rhythm with comfortable 44px and compact 36px control contracts.",
    keywords: ["spacing", "density", "compact", "comfortable"], sourcePath: "src/foundations/tokens.css",
    whenToUse: ["Apply one density at an application or bounded preview root."],
    content: ["Comfortable is the product default; compact is for information-dense administration without shrinking icon-only targets below an accessible size."],
    requirements: ["Density may change spacing and type roles but not semantics, names, or focus visibility."],
    publicContract: ["Set data-density to comfortable or compact at a root boundary."],
  }),
  guidanceDocument({
    id: "foundation.shape-and-borders", route: "/foundations/shape-and-borders", name: "Shape and borders", kind: "foundation",
    summary: "Square geometry, quiet rules, straight accent rails, and deliberate lower edges.",
    keywords: ["shape", "borders", "rules", "accent rail"], sourcePath: "src/foundations/tokens.css",
    whenToUse: ["Use borders to explain structure or interaction, never to decorate every region."],
    content: ["Static surfaces use a quiet one-pixel rule. Pale controls receive a subtle two-pixel lower edge; filled actions keep a stronger action edge.", "A colored accent rail replaces the left hairline and overlays its endpoints so no mitered joins appear."],
    requirements: ["Focus rings remain distinct from borders and status rails."],
    publicContract: ["Use paper-accent-rail and the public field/action recipes."],
  }),
  guidanceDocument({
    id: "foundation.elevation", route: "/foundations/elevation", name: "Elevation", kind: "foundation",
    summary: "Flat ordinary surfaces with hard-offset depth for floating and rare showcase layers.",
    keywords: ["elevation", "shadow", "overlay", "surface"], sourcePath: "src/foundations/tokens.css",
    whenToUse: ["Use floating for transient layers and showcase only for deliberate demonstrations."],
    content: ["Ordinary cards remain flat. Floating uses a three-pixel hard offset; showcase uses five pixels and should stay rare."],
    requirements: ["Elevation must not be the only cue for hierarchy or interactivity."],
    publicContract: ["Use paper-elevation-flat, paper-elevation-floating, or paper-elevation-showcase."],
  }),
  guidanceDocument({
    id: "foundation.iconography", route: "/foundations/iconography", name: "Iconography", kind: "foundation",
    summary: "A curated Heroicons grammar for actions, navigation, status, and confirmation.",
    keywords: ["icons", "heroicons", "labels", "size"], sourcePath: "src/icons/index.ts",
    whenToUse: ["Use an icon when it improves recognition or saves space without obscuring the action."],
    content: ["Outline icons represent navigation and actions; solid icons represent confirmation and status. Sizes are compact 16px, default 20px, prominent 24px, and 48px only for empty-state illustration."],
    requirements: ["Decorative icons are hidden; icon-only controls have stable accessible names."],
    publicContract: ["Import curated symbols and iconClassName from paper-ui/icons."],
  }),
  guidanceDocument({
    id: "foundation.motion", route: "/foundations/motion", name: "Motion", kind: "foundation",
    summary: "Quick, standard, and deliberate motion roles with reduced-motion safety.",
    keywords: ["motion", "transition", "reduced motion"], sourcePath: "src/foundations/tokens.css",
    whenToUse: ["Use motion to explain a state change, not to decorate static content."],
    content: ["Quick feedback is 120ms, standard transitions are 180ms, and deliberate movement is 240ms. Components suppress nonessential animation when reduced motion is requested."],
    requirements: ["No required information may depend on animation completing or being perceived."],
    publicContract: ["Use --paper-motion-quick, --paper-motion-standard, and --paper-motion-deliberate."],
  }),
  guidanceDocument({
    id: "foundation.layout", route: "/foundations/layout", name: "Layout", kind: "foundation",
    summary: "Responsive page width, prose measure, stack, cluster, and overflow guidance.",
    keywords: ["layout", "responsive", "measure", "stack"], sourcePath: "src/foundations/base.css",
    whenToUse: ["Compose pages from document flow first, then add grids only when relationships require them."],
    content: ["Keep reading copy to a legible measure, let controls wrap before they overflow, and place horizontal data inside an explicitly labelled scroll region."],
    requirements: ["Layouts reflow at 320 CSS pixels without losing content or two-dimensional scrolling."],
    publicContract: ["Paper utilities cover stacks, clusters, fixture rows, and readable measures."],
  }),
  guidanceDocument({
    id: "foundation.brand", route: "/foundations/brand", name: "Brand", kind: "foundation",
    summary: "Canonical Cut Meter, wordmarks, favicons, and monochrome-first lockup rules.",
    keywords: ["brand", "logo", "cut meter", "favicon"], sourcePath: "src/assets/brand/README.md",
    whenToUse: ["Use the packaged canonical asset that matches the surface and available width."],
    content: ["The Cut Meter is proven in solid monochrome first, reversed second, and accented only as an optional third expression.", "Never filter, redraw, stretch, recolor individual bars, or reconstruct the diagonal cut."],
    requirements: ["Use meaningful alternate text only when the surrounding wordmark does not already name Tadoku."],
    publicContract: ["Import brand files through paper-ui/assets/brand/*."],
  }),
] as const;

function ReadingLogPatternFixture() {
  return (
    <Surface as="article" accent>
      <p className="paper-type-metadata">AUGUST JAPANESE</p>
      <h3 className="paper-type-component">Today’s reading</h3>
      <p>48 pages of コンビニ人間 · private notes saved</p>
      <ButtonGroup actions={[
        { id: "view", label: "View log", href: "#view-log", variant: "outline" },
        { id: "continue", label: "Add reading", onSelect: () => undefined },
      ]} />
    </Surface>
  );
}

function LoggingExperimentFixture() {
  const methods = useForm({ defaultValues: { title: "コンビニ人間", pages: "48" } });
  return (
    <FormProvider {...methods}>
      <form className="paper-fixture-form" onSubmit={methods.handleSubmit(() => undefined)}>
        <Flash title="Experimental flow">Contest submission remains a separate confirmation step.</Flash>
        <Input name="title" label="Work" />
        <Input name="pages" label="Pages" inputMode="numeric" />
        <Button type="submit">Review entry</Button>
      </form>
    </FormProvider>
  );
}

export const phaseThreeContentFixtures = [
  defineCatalogFixture({
    id: "pattern.logging-summary", name: "Reading-log summary", description: "A realistic summary with explicit navigation and continuation actions.",
    tags: ["logging", "pattern", "reading"], themes: ["light", "dark"], densities: ["comfortable", "compact"], viewports: VIEWPORTS,
    deterministic: true,
    code: `import { ButtonGroup, Surface } from "paper-ui";\n\n<Surface as="article" accent>\n  <h3>Today’s reading</h3>\n  <p>48 pages of コンビニ人間</p>\n  <ButtonGroup actions={actions} />\n</Surface>`,
    render: () => <ReadingLogPatternFixture />,
  }),
  defineCatalogFixture({
    id: "experiment.logging-v2-entry", name: "Logging v2 entry", description: "A deterministic experiment separating saved reading from contest submission.",
    tags: ["logging", "experiment", "form"], themes: ["light", "dark"], densities: ["comfortable", "compact"], viewports: VIEWPORTS,
    deterministic: true,
    code: `import { Button, Flash, Input } from "paper-ui";\nimport { FormProvider, useForm } from "react-hook-form";\n\n<FormProvider {...methods}>\n  <form onSubmit={methods.handleSubmit(reviewEntry)}>\n    <Flash title="Experimental flow">Contest submission is separate.</Flash>\n    <Input name="title" label="Work" />\n    <Input name="pages" label="Pages" inputMode="numeric" />\n    <Button type="submit">Review entry</Button>\n  </form>\n</FormProvider>`,
    render: () => <LoggingExperimentFixture />,
  }),
] as const;

export const phaseThreePatternDocuments = [
  defineCatalogDocument({
    id: "pattern.logging", route: "/patterns/logging", name: "Logging", kind: "pattern", category: "patterns", aliases: [],
    summary: "Composes Paper primitives into reading summaries, entry actions, and transparent contest state.",
    keywords: ["logging", "reading", "contest", "summary"], lifecycle: "Stable", owner: OWNER, reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/phase-three-content.tsx", packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: ["Use for product flows that record reading and explain where it contributes."],
      whenNotToUse: ["Do not turn the product flow into a reusable low-level component."],
      content: ["Name the work, amount, language, privacy, and contest state in plain Tadoku language."],
      commonMistakes: ["Hiding whether an entry is merely saved or also submitted to a contest."],
    },
    accessibility: {
      requirements: ["Keep summaries in document order and expose every action as a named link or button."],
      keyboard: ["All actions follow native link and button operation."], knownConstraints: [],
    },
    api: { react: ["Surface", "ButtonGroup", "Flash"], cssClasses: [], publicTypes: [], defaults: ["Applications own routing, data, and mutation state."], invalidCombinations: ["Patterns must not import application routers into paper-ui."] },
    fixtureIds: ["pattern.logging-summary"], dependencies: { documents: ["component.surface", "component.button-group"], packages: ["paper-ui"] },
    migration: { legacy: ["styleguide/pages/logging.tsx"], notes: ["Preserve the realistic information hierarchy, not the legacy local-state implementation."] },
    changelog: [{ date: REVIEW_DATE, note: "Rebuilt logging as a product pattern." }], behaviorTestIds: ["pattern.logging.render"],
  }),
] as const;

export const phaseThreeExperimentDocuments = [
  defineCatalogDocument({
    id: "experiment.logging-v2", route: "/experiments/logging-v2", name: "Logging v2", kind: "experiment", category: "experiments", aliases: [],
    summary: "Explores a clearer split between saving reading, reviewing details, and submitting to contests.",
    keywords: ["logging", "experiment", "contest", "form"], lifecycle: "Experimental", owner: OWNER, reviewDate: REVIEW_DATE,
    sourcePath: "src/catalog/phase-three-content.tsx", packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: ["Use only for evaluating the proposed logging flow."],
      whenNotToUse: ["Do not treat this experiment as a Stable product or component contract."],
      content: ["Keep saved, reviewed, and submitted states explicit in labels and confirmation copy."],
      commonMistakes: ["Promoting experimental layout or copy into applications without product review."],
    },
    accessibility: { requirements: ["Retain visible labels, error relationships, and explicit confirmation state."], keyboard: ["Use native form control and button behavior."], knownConstraints: ["Product validation is not complete."] },
    api: { react: ["Flash", "Button"], cssClasses: ["paper-field", "paper-input"], publicTypes: [], defaults: ["The application owns React Hook Form and submission state."], invalidCombinations: ["Do not expose the experiment as a paper-ui component."] },
    fixtureIds: ["experiment.logging-v2-entry"], dependencies: { documents: ["pattern.logging"], packages: ["paper-ui", "react-hook-form"] },
    migration: { legacy: ["styleguide/pages/logging-v2.tsx"], notes: ["Keep this visibly Experimental until a product decision promotes or removes it."] },
    changelog: [{ date: REVIEW_DATE, note: "Preserved logging v2 as a deterministic experiment." }], behaviorTestIds: ["experiment.logging-v2.render"],
  }),
] as const;

export const phaseThreeGovernanceDocuments = [
  guidanceDocument({
    id: "governance.contributing", route: "/contributing", name: "Contributing", kind: "governance",
    summary: "How Paper changes are proposed, implemented, tested, documented, reviewed, and promoted.",
    keywords: ["contributing", "lifecycle", "ownership", "review", "deprecation"], sourcePath: "docs/wip/tadoku-paper/implementation-plan.md",
    whenToUse: ["Use before adding an export, token, fixture, route, or lifecycle change."],
    content: ["Every change has an owner, canonical source, deterministic fixture, behavior evidence, migration note, and review date before Stable promotion.", "Experimental contracts may change. Stable contracts require migration-safe changes. Deprecated contracts name a registered replacement, removal timing, and exit path.", "Design history includes the refinement audit, visual studies, decision log, implementation plan, research packet, and accepted ADRs under docs/wip/tadoku-paper."],
    requirements: ["Review semantics, keyboard behavior, reflow, both themes, both densities, and realistic content before promotion."],
    publicContract: ["The catalogue registry is the source of lifecycle, ownership, route, fixture, source, and migration metadata."],
  }),
  guidanceDocument({
    id: "governance.changelog", route: "/changelog", name: "Changelog", kind: "governance",
    summary: "Meaningful package, component, documentation, deployment, and migration changes.",
    keywords: ["changelog", "release", "version", "migration"], sourcePath: "docs/wip/tadoku-paper/research",
    whenToUse: ["Consult before upgrading Paper or changing an application integration."],
    content: ["0.1.0 establishes semantic foundations, the complete catalogue contract, static delivery at paper.tadoku.app, framework-neutral components, and TypeScript 4.9-compatible built declarations.", "TypeScript 7.0.2 performs Paper typechecks while the TypeScript 6 compatibility API remains available to ESLint and declaration tooling."],
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
