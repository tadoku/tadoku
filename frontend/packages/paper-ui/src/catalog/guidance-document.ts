import { defineCatalogDocument, type CatalogDocument } from "./schema";

export const REVIEW_DATE = "2026-09-05";
export const PACKAGE_VERSION = "0.1.0";

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

export function guidanceDocument(options: GuidanceDocumentOptions): CatalogDocument {
  return defineCatalogDocument({
    ...options,
    category: options.kind === "foundation" ? "foundations" : "governance",
    aliases: [],
    lifecycle: "Experimental",
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
      knownConstraints: ["Review these roles in the consuming screen; a foundation sample cannot certify an application’s accessibility."],
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
