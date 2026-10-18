import { type ReactNode } from "react";
import {
  COMPONENT_PAGE_SECTION_KEYS,
  defineCatalogDocument,
  defineCatalogFixture,
  type CatalogDocument,
  type RequiredComponentSections,
} from "./schema";

export const REVIEW_DATE = "2026-10-12";
const VIEWPORTS = [
  { id: "phone", label: "Phone", width: 360, height: 720 },
  { id: "tablet", label: "Tablet", width: 768, height: 800 },
  { id: "desktop", label: "Desktop", width: 1280, height: 800 },
] as const;
export const EXPERIMENTAL_LIFECYCLE =
  "Experimental in Paper 0.1.0. Extracted from repeated paper-playground screens; no production application consumes it yet.";

export function fixture(input: {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly tags: readonly string[];
  readonly code: string;
  readonly render: () => ReactNode;
}) {
  return defineCatalogFixture({
    ...input,
    themes: ["light", "dark"],
    densities: ["comfortable", "compact"],
    viewports: VIEWPORTS,
    deterministic: true,
  });
}

type ComponentInput = Pick<
  CatalogDocument,
  "id" | "route" | "name" | "category" | "summary" | "keywords" | "sourcePath" | "fixtureIds" | "behaviorTestIds"
> & {
  readonly sections: RequiredComponentSections;
  readonly commonMistakes: readonly string[];
  readonly api: CatalogDocument["api"];
  readonly dependencies?: readonly string[];
  readonly legacy: readonly string[];
};

export function componentDocument({
  sections,
  commonMistakes,
  dependencies = [],
  legacy,
  ...input
}: ComponentInput): CatalogDocument {
  return defineCatalogDocument({
    ...input,
    kind: "component",
    aliases: [],
    lifecycle: "Experimental",
    reviewDate: REVIEW_DATE,
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: sections.whenToUse.content,
      whenNotToUse: sections.whenNotToUse.content,
      content: sections.contentGuidance.content,
      commonMistakes,
    },
    accessibility: {
      requirements: sections.accessibility.content,
      keyboard: sections.behavior.content,
      knownConstraints: sections.statesAndAdaptation.content,
    },
    dependencies: { documents: dependencies, packages: ["paper-ui"] },
    migration: { legacy, notes: sections.migration.content },
    changelog: [{ date: REVIEW_DATE, note: "Added as an Experimental page-composition control." }],
    sections: { required: sections, pageSections: COMPONENT_PAGE_SECTION_KEYS },
  });
}
