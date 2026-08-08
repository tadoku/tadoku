import {
  CatalogValidationError,
  defineCatalogDocument,
  type CatalogDocument,
  type CatalogRegistry,
  type CatalogRegistryInput,
} from "./schema";
import { validateCatalogRegistry } from "./validation";

const REVIEW_DATE = "2026-08-08";
const PACKAGE_VERSION = "0.1.0";
const OWNER = "Tadoku design systems";

interface StubDocumentOptions {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly kind: "foundation" | "governance";
  readonly summary: string;
  readonly keywords: readonly string[];
  readonly sourcePath: string;
}

function stubDocument(options: StubDocumentOptions): CatalogDocument {
  return defineCatalogDocument({
    ...options,
    category: options.kind === "foundation" ? "foundations" : "governance",
    aliases: [],
    lifecycle: "Experimental",
    owner: OWNER,
    reviewDate: REVIEW_DATE,
    packageVersion: PACKAGE_VERSION,
    guidance: {
      whenToUse: ["Use this guidance when designing or reviewing Paper interfaces."],
      whenNotToUse: ["Do not use this page as a substitute for a component contract."],
      content: ["Prefer concise, specific Tadoku language."],
      commonMistakes: ["Do not copy legacy styling decisions without reviewing their intent."],
    },
    accessibility: {
      requirements: ["Apply the guidance without reducing semantic or keyboard access."],
      keyboard: [],
      knownConstraints: ["The Phase 1 page is an initial catalogue stub."],
    },
    api: {
      react: [],
      cssClasses: [],
      publicTypes: [],
      defaults: ["Canonical values come from paper-ui package sources."],
      invalidCombinations: [],
    },
    fixtureIds: [],
    dependencies: { documents: [], packages: ["paper-ui"] },
    migration: {
      legacy: [],
      notes: ["Legacy guidance remains authoritative until its consuming application migrates."],
    },
    changelog: [{ date: REVIEW_DATE, note: "Added the initial Phase 1 catalogue record." }],
    behaviorTestIds: [],
  });
}

export const foundationDocuments = [
  stubDocument({
    id: "foundation.principles",
    route: "/foundations/principles",
    name: "Principles",
    kind: "foundation",
    summary: "The product and design principles that guide Tadoku Paper decisions.",
    keywords: ["principles", "bookplate", "design"],
    sourcePath: "src/catalog/registry.ts",
  }),
  stubDocument({
    id: "foundation.color",
    route: "/foundations/color",
    name: "Color",
    kind: "foundation",
    summary: "Semantic surface, text, action, status, rule, and focus color roles.",
    keywords: ["color", "theme", "tokens"],
    sourcePath: "src/foundations/tokens.css",
  }),
  stubDocument({
    id: "foundation.typography",
    route: "/foundations/typography",
    name: "Typography",
    kind: "foundation",
    summary: "The Merriweather and Open Sans hierarchy used throughout Paper.",
    keywords: ["typography", "fonts", "hierarchy"],
    sourcePath: "src/foundations/fonts.css",
  }),
  stubDocument({
    id: "foundation.spacing-and-density",
    route: "/foundations/spacing-and-density",
    name: "Spacing and density",
    kind: "foundation",
    summary: "Spacing roles and the comfortable and compact density contracts.",
    keywords: ["spacing", "density", "compact", "comfortable"],
    sourcePath: "src/foundations/tokens.css",
  }),
  stubDocument({
    id: "foundation.shape-and-borders",
    route: "/foundations/shape-and-borders",
    name: "Shape and borders",
    kind: "foundation",
    summary: "Square geometry, quiet rules, accent rails, and interactive lower edges.",
    keywords: ["shape", "borders", "rules", "accent rail"],
    sourcePath: "src/foundations/tokens.css",
  }),
  stubDocument({
    id: "foundation.elevation",
    route: "/foundations/elevation",
    name: "Elevation",
    kind: "foundation",
    summary: "Flat surfaces and the restrained hard-offset shadows used by overlays.",
    keywords: ["elevation", "shadow", "overlay"],
    sourcePath: "src/foundations/tokens.css",
  }),
  stubDocument({
    id: "foundation.iconography",
    route: "/foundations/iconography",
    name: "Iconography",
    kind: "foundation",
    summary: "Curated icon roles, weights, sizes, and accessible labeling guidance.",
    keywords: ["icons", "heroicons", "size"],
    sourcePath: "src/icons/index.ts",
  }),
  stubDocument({
    id: "foundation.motion",
    route: "/foundations/motion",
    name: "Motion",
    kind: "foundation",
    summary: "Quick, standard, and deliberate motion roles with reduced-motion safety.",
    keywords: ["motion", "transition", "reduced motion"],
    sourcePath: "src/foundations/tokens.css",
  }),
  stubDocument({
    id: "foundation.layout",
    route: "/foundations/layout",
    name: "Layout",
    kind: "foundation",
    summary: "Page width, prose measure, spacing, and responsive composition guidance.",
    keywords: ["layout", "responsive", "measure"],
    sourcePath: "src/foundations/base.css",
  }),
  stubDocument({
    id: "foundation.brand",
    route: "/foundations/brand",
    name: "Brand",
    kind: "foundation",
    summary: "Cut Meter, wordmark, favicon, and monochrome-first brand usage.",
    keywords: ["brand", "logo", "cut meter", "favicon"],
    sourcePath: "src/assets/brand/index.ts",
  }),
] as const;

export const governanceDocuments = [
  stubDocument({
    id: "governance.contributing",
    route: "/contributing",
    name: "Contributing",
    kind: "governance",
    summary: "How to propose, implement, test, document, and promote Paper changes.",
    keywords: ["contributing", "review", "stable"],
    sourcePath: "src/catalog/registry.ts",
  }),
  stubDocument({
    id: "governance.changelog",
    route: "/changelog",
    name: "Changelog",
    kind: "governance",
    summary: "Meaningful package, component, documentation, and migration changes.",
    keywords: ["changelog", "release", "version"],
    sourcePath: "src/catalog/registry.ts",
  }),
] as const;

export function createCatalogRegistry(input: CatalogRegistryInput): CatalogRegistry {
  const registry: CatalogRegistry = {
    documents: input.documents,
    fixtures: input.fixtures,
    redirects: input.redirects ?? [],
  };
  const result = validateCatalogRegistry(registry);
  if (!result.valid) throw new CatalogValidationError(result.issues);
  return registry;
}

export const catalogRegistry = createCatalogRegistry({
  documents: [...foundationDocuments, ...governanceDocuments],
  fixtures: [],
  redirects: [
    { from: "/color", to: "/foundations/color" },
    { from: "/typography", to: "/foundations/typography" },
    { from: "/branding", to: "/foundations/brand" },
    { from: "/templates", to: "/foundations/layout" },
  ],
});
