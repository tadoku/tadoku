import type { ReactNode } from "react";

export const CATALOG_DOCUMENT_KINDS = [
  "foundation",
  "component",
  "pattern",
  "experiment",
  "governance",
] as const;

export const CATALOG_LIFECYCLES = [
  "Experimental",
  "Stable",
  "Deprecated",
] as const;

export const COMPONENT_CATEGORIES = [
  "actions",
  "forms",
  "navigation",
  "feedback",
  "overlays",
  "data-display",
] as const;

export const CATALOG_CATEGORIES = [
  "foundations",
  ...COMPONENT_CATEGORIES,
  "patterns",
  "experiments",
  "governance",
] as const;

export const FIXTURE_THEMES = ["light", "dark"] as const;
export const FIXTURE_DENSITIES = ["comfortable", "compact"] as const;

/**
 * The order is part of the public documentation contract. Keep additions in
 * `optionalSections`; changing this list changes the Stable-page contract.
 */
export const REQUIRED_COMPONENT_SECTION_KEYS = [
  "overview",
  "whenToUse",
  "whenNotToUse",
  "choosingBetween",
  "anatomy",
  "recommendedExample",
  "variants",
  "statesAndAdaptation",
  "behavior",
  "contentGuidance",
  "accessibility",
  "implementation",
  "apiReference",
  "relatedPatterns",
  "migration",
  "lifecycle",
] as const;

/**
 * The small set of topics that may appear in a consumer-facing component
 * guide. The larger required-section contract records catalogue evidence;
 * authors opt into only the topics whose copy is useful on the page.
 */
export const COMPONENT_PAGE_SECTION_KEYS = [
  "usage",
  "examples",
  "variantsAndStates",
  "behavior",
  "contentGuidance",
  "accessibility",
] as const;

export type CatalogKind = (typeof CATALOG_DOCUMENT_KINDS)[number];
export type CatalogLifecycle = (typeof CATALOG_LIFECYCLES)[number];
export type ComponentCategory = (typeof COMPONENT_CATEGORIES)[number];
export type CatalogCategory = (typeof CATALOG_CATEGORIES)[number];
export type FixtureTheme = (typeof FIXTURE_THEMES)[number];
export type FixtureDensity = (typeof FIXTURE_DENSITIES)[number];
export type RequiredComponentSectionKey =
  (typeof REQUIRED_COMPONENT_SECTION_KEYS)[number];
export type ComponentPageSectionKey =
  (typeof COMPONENT_PAGE_SECTION_KEYS)[number];

export interface CatalogGuidance {
  readonly whenToUse: readonly string[];
  readonly whenNotToUse: readonly string[];
  readonly content: readonly string[];
  readonly commonMistakes: readonly string[];
}

export interface CatalogAccessibility {
  readonly requirements: readonly string[];
  readonly keyboard: readonly string[];
  readonly knownConstraints: readonly string[];
}

export interface CatalogApiContract {
  readonly react: readonly string[];
  readonly cssClasses: readonly string[];
  readonly publicTypes: readonly string[];
  readonly defaults: readonly string[];
  readonly invalidCombinations: readonly string[];
}

export interface CatalogDependencies {
  /** Stable IDs of other catalogue documents. */
  readonly documents: readonly string[];
  /** Published package names used by the documented public contract. */
  readonly packages: readonly string[];
}

export interface CatalogMigration {
  readonly legacy: readonly string[];
  readonly notes: readonly string[];
  /** Required for Deprecated documents and resolved against this registry. */
  readonly replacementDocumentId?: string;
  /** Required for Deprecated documents. Include timing and an exit path. */
  readonly removalGuidance?: string;
}

export interface CatalogChangelogEntry {
  /** ISO 8601 calendar date (YYYY-MM-DD). */
  readonly date: string;
  readonly note: string;
}

export interface CatalogDocumentationSection {
  readonly heading: string;
  readonly content: readonly string[];
}

export type RequiredComponentSections = {
  readonly [Key in RequiredComponentSectionKey]: CatalogDocumentationSection;
};

export interface OptionalComponentSection extends CatalogDocumentationSection {
  /** Places the section without replacing or reordering a required section. */
  readonly after: RequiredComponentSectionKey;
}

export interface ComponentDocumentationSections {
  /**
   * Partial by design: authors can draft a page while Experimental. Runtime
   * registry validation requires every key before it can become Stable.
   */
  readonly required: Partial<RequiredComponentSections>;
  /** Ordered, intentionally curated topics rendered on the public page. */
  readonly pageSections: readonly ComponentPageSectionKey[];
  readonly optional?: readonly OptionalComponentSection[];
}

export interface CatalogDocument {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly kind: CatalogKind;
  readonly category: CatalogCategory;
  readonly aliases: readonly string[];
  readonly summary: string;
  readonly keywords: readonly string[];
  readonly lifecycle: CatalogLifecycle;
  /** ISO 8601 calendar date (YYYY-MM-DD). */
  readonly reviewDate: string;
  readonly sourcePath: string;
  readonly packageVersion: string;
  readonly guidance: CatalogGuidance;
  readonly accessibility: CatalogAccessibility;
  readonly api: CatalogApiContract;
  readonly fixtureIds: readonly string[];
  readonly dependencies: CatalogDependencies;
  readonly migration: CatalogMigration;
  readonly changelog: readonly CatalogChangelogEntry[];
  /** Stable identifiers for rendered semantic or behavioral tests. */
  readonly behaviorTestIds: readonly string[];
  /** Component pages store the ordered instructional contract here. */
  readonly sections?: ComponentDocumentationSections;
}

export interface CatalogFixtureViewport {
  readonly id: string;
  readonly label: string;
  readonly width: number;
  readonly height: number;
}

export interface CatalogFixture {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly tags: readonly string[];
  readonly themes: readonly FixtureTheme[];
  readonly densities: readonly FixtureDensity[];
  readonly viewports: readonly CatalogFixtureViewport[];
  /**
   * An explicit review marker. Runtime validation also checks the render
   * function for known time, randomness, network, and router dependencies.
   */
  readonly deterministic: true;
  /** Complete copyable consumer example for the catalogue Code view. */
  readonly code?: string;
  readonly render: () => ReactNode;
}

export interface CatalogRedirect {
  readonly from: string;
  readonly to: string;
}

export interface CatalogRegistryInput {
  readonly documents: readonly CatalogDocument[];
  readonly fixtures: readonly CatalogFixture[];
  readonly redirects?: readonly CatalogRedirect[];
}

export interface CatalogRegistry {
  readonly documents: readonly CatalogDocument[];
  readonly fixtures: readonly CatalogFixture[];
  readonly redirects: readonly CatalogRedirect[];
}

export type CatalogValidationCode =
  | "invalid-registry"
  | "invalid-document"
  | "invalid-fixture"
  | "invalid-redirect"
  | "invalid-id"
  | "duplicate-document-id"
  | "duplicate-fixture-id"
  | "duplicate-route"
  | "duplicate-slug"
  | "duplicate-alias"
  | "duplicate-reference"
  | "invalid-kind"
  | "invalid-lifecycle"
  | "invalid-category"
  | "invalid-route"
  | "invalid-date"
  | "missing-metadata"
  | "missing-stable-section"
  | "missing-stable-fixture"
  | "missing-stable-behavior-test"
  | "missing-stable-accessibility"
  | "invalid-section"
  | "unknown-fixture-reference"
  | "unknown-document-reference"
  | "invalid-deprecation"
  | "invalid-fixture-marker"
  | "non-deterministic-fixture"
  | "router-dependent-fixture";

export interface CatalogValidationIssue {
  readonly code: CatalogValidationCode;
  readonly path: string;
  readonly message: string;
}

export interface CatalogValidationResult {
  readonly valid: boolean;
  readonly issues: readonly CatalogValidationIssue[];
}

export class CatalogValidationError extends Error {
  readonly issues: readonly CatalogValidationIssue[];

  constructor(issues: readonly CatalogValidationIssue[]) {
    const details = issues
      .map((issue) => `${issue.path}: ${issue.message}`)
      .join("\n");
    super(`Invalid Paper catalogue registry:\n${details}`);
    this.name = "CatalogValidationError";
    this.issues = issues;
  }
}

export function defineCatalogDocument<Document extends CatalogDocument>(
  document: Document,
): Document {
  return document;
}

export function defineCatalogFixture<Fixture extends CatalogFixture>(
  fixture: Fixture,
): Fixture {
  return fixture;
}
