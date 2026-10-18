import {
COMPONENT_PAGE_SECTION_KEYS,
defineCatalogDocument,
type CatalogDocument,
type ComponentDocumentationSections,
type RequiredComponentSections,
} from "./schema";
import { siteNavigationFixtures, siteNavigationSpecs } from "./site-navigation";
import { viewNavigationFixtures, viewNavigationSpecs } from "./view-navigation";

const REVIEW_DATE = "2026-09-05";

export interface NavigationDocSpec {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly summary: string;
  readonly fixtureId: string;
  readonly sourcePath: string;
  readonly react: readonly string[];
  readonly types: readonly string[];
  readonly css: readonly string[];
  readonly when: string;
  readonly avoid: string;
  readonly choose: string;
  readonly anatomy: string;
  readonly variants: string;
  readonly states: string;
  readonly behavior: string;
  readonly content: string;
  readonly accessibility: string;
  readonly migration: string;
  readonly packages?: readonly string[];
  readonly example: string;
  readonly props: NonNullable<CatalogDocument["api"]["props"]>;
}

function navigationSections(spec: NavigationDocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [spec.anatomy] },
    recommendedExample: { heading: "Recommended example", content: [spec.example] },
    variants: { heading: "Variants", content: [spec.variants] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [`Import ${spec.react.join(", ")} from paper-ui and load paper-ui/styles.css once at the application entry. Examples use native fragment links so they run without a router. For an application router, pass renderLink={(props) => <AppLink {...props} />}; forward href, className, children, tabIndex, event handlers and aria attributes to the final anchor. Derive currentPath from your router; current on an item explicitly overrides exact href matching.`] },
    apiReference: { heading: "API reference", content: [`See API / Props for required data, defaults, and routing contracts. Public types: ${spec.types.join(", ")}.`] },
    relatedPatterns: { heading: "Related patterns", content: ["Combine navigation components with the application shell and real destination data; do not duplicate routing or authorization policy inside Paper."] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: ["Stable in Paper 0.1.0 with deterministic fixtures and rendered semantic, keyboard, and router-boundary tests."] },
  };
  return { required, pageSections: COMPONENT_PAGE_SECTION_KEYS };
}

function navigationDocument(spec: NavigationDocSpec): CatalogDocument {
  return defineCatalogDocument({
    id: spec.id,
    route: spec.route,
    name: spec.name,
    kind: "component",
    category: "navigation",
    aliases: [],
    summary: spec.summary,
    keywords: [spec.name.toLocaleLowerCase(), "navigation", "router neutral", "paper"],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    sourcePath: spec.sourcePath,
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: [spec.when],
      whenNotToUse: [spec.avoid],
      content: [spec.content],
      commonMistakes: [spec.choose],
    },
    accessibility: {
      requirements: [spec.accessibility],
      keyboard: [spec.behavior],
      knownConstraints: [spec.states],
    },
    api: {
      react: spec.react,
      props: spec.props,
      cssClasses: spec.css,
      publicTypes: spec.types,
      defaults: [
        "Links render as native anchors; renderLink may replace them without changing semantics.",
        ...(spec.name === "Navbar"
          ? ["actions renders end-aligned utilities; mobileNavigation defaults to true and may be disabled when the application provides another narrow navigation pattern."]
          : []),
      ],
      invalidCombinations: [spec.avoid],
    },
    fixtureIds: [spec.fixtureId],
    dependencies: { documents: ["foundation.iconography"], packages: spec.packages ?? ["paper-ui"] },
    migration: { legacy: [`ui ${spec.name}`], notes: [spec.migration] },
    changelog: [{ date: REVIEW_DATE, note: `Published router-neutral ${spec.name} as Stable.` }],
    behaviorTestIds: [`${spec.id}.semantics`, `${spec.id}.keyboard`, `${spec.id}.router-boundary`],
    sections: navigationSections(spec),
  });
}

export const phaseThreeNavigationFixtures = [
  ...siteNavigationFixtures,
  ...viewNavigationFixtures,
] as const;

const specs: readonly NavigationDocSpec[] = [
  ...siteNavigationSpecs,
  ...viewNavigationSpecs,
];

export const phaseThreeNavigationDocuments = specs.map(navigationDocument);
