export {
  CATALOG_CATEGORIES,
  CATALOG_DOCUMENT_KINDS,
  CATALOG_LIFECYCLES,
  CatalogValidationError,
  COMPONENT_CATEGORIES,
  defineCatalogDocument,
  defineCatalogFixture,
  FIXTURE_DENSITIES,
  FIXTURE_THEMES,
  REQUIRED_COMPONENT_SECTION_KEYS,
} from "./schema";
export type {
  CatalogAccessibility,
  CatalogApiContract,
  CatalogCategory,
  CatalogChangelogEntry,
  CatalogDependencies,
  CatalogDocument,
  CatalogDocumentationSection,
  CatalogFixture,
  CatalogFixtureViewport,
  CatalogGuidance,
  CatalogKind,
  CatalogLifecycle,
  CatalogMigration,
  CatalogRedirect,
  CatalogRegistry,
  CatalogRegistryInput,
  CatalogValidationCode,
  CatalogValidationIssue,
  CatalogValidationResult,
  ComponentCategory,
  ComponentDocumentationSections,
  FixtureDensity,
  FixtureTheme,
  OptionalComponentSection,
  RequiredComponentSectionKey,
  RequiredComponentSections,
} from "./schema";
export {
  catalogRegistry,
  createCatalogRegistry,
  foundationDocuments,
  governanceDocuments,
} from "./registry";
export { validateCatalogFixture, validateCatalogRegistry } from "./validation";
export {
  phaseThreeNavigationDocuments,
  phaseThreeNavigationFixtures,
} from "./phase-three-navigation";
export {
  phaseThreeDataLayoutDocuments,
  phaseThreeDataLayoutFixtures,
} from "./phase-three-data-layout";
