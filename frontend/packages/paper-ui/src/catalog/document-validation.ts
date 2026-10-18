import {
  COMPONENT_PAGE_SECTION_KEYS,
  COMPONENT_CATEGORIES,
  REQUIRED_COMPONENT_SECTION_KEYS,
  type CatalogCategory,
  type CatalogDocument,
  type CatalogKind,
  type CatalogValidationIssue,
} from "./schema";
import { addIssue, isNonEmptyString, validateStringList } from "./validation-primitives";

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

function isNonEmptyStringArray(value: unknown): value is readonly string[] {
  return (
    Array.isArray(value) &&
    value.length > 0 &&
    value.every((item) => isNonEmptyString(item))
  );
}

function isCalendarDate(value: unknown): value is string {
  if (typeof value !== "string" || !DATE_PATTERN.test(value)) return false;
  const [year, month, day] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return (
    date.getUTCFullYear() === year &&
    date.getUTCMonth() === month - 1 &&
    date.getUTCDate() === day
  );
}

export function categoryForKindIsValid(
  kind: CatalogKind,
  category: CatalogCategory,
): boolean {
  switch (kind) {
    case "foundation":
      return category === "foundations";
    case "component":
      return (COMPONENT_CATEGORIES as readonly string[]).includes(category);
    case "pattern":
      return category === "patterns";
    case "experiment":
      return category === "experiments";
    case "governance":
      return category === "governance";
  }
}

export function expectedRoutePattern(document: CatalogDocument): RegExp {
  switch (document.kind) {
    case "foundation":
      return /^\/foundations\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "component":
      return /^\/components\/(?:actions|forms|navigation|feedback|overlays|data-display|page-structure)\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "pattern":
      return /^\/patterns\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "experiment":
      return /^\/experiments\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "governance":
      return /^\/(?:contributing|changelog)$/;
  }
}

export function validateDocumentMetadata(
  document: CatalogDocument,
  path: string,
  issues: CatalogValidationIssue[],
): void {
  const strings: ReadonlyArray<readonly [unknown, string]> = [
    [document.name, "name"],
    [document.summary, "summary"],
    [document.sourcePath, "sourcePath"],
    [document.packageVersion, "packageVersion"],
  ];
  strings.forEach(([value, key]) => {
    if (!isNonEmptyString(value)) {
      addIssue(issues, "missing-metadata", `${path}.${key}`, "must be a non-empty string");
    }
  });

  if (!isCalendarDate(document.reviewDate)) {
    addIssue(
      issues,
      "invalid-date",
      `${path}.reviewDate`,
      "must be a real ISO calendar date (YYYY-MM-DD)",
    );
  }

  validateStringList(issues, document.aliases, `${path}.aliases`);
  validateStringList(issues, document.keywords, `${path}.keywords`, { required: true });
  validateStringList(issues, document.fixtureIds, `${path}.fixtureIds`);
  validateStringList(issues, document.behaviorTestIds, `${path}.behaviorTestIds`);

  if (!document.guidance || typeof document.guidance !== "object") {
    addIssue(issues, "missing-metadata", `${path}.guidance`, "is required");
  } else {
    validateStringList(issues, document.guidance.whenToUse, `${path}.guidance.whenToUse`);
    validateStringList(
      issues,
      document.guidance.whenNotToUse,
      `${path}.guidance.whenNotToUse`,
    );
    validateStringList(issues, document.guidance.content, `${path}.guidance.content`);
    validateStringList(
      issues,
      document.guidance.commonMistakes,
      `${path}.guidance.commonMistakes`,
    );
  }

  if (!document.accessibility || typeof document.accessibility !== "object") {
    addIssue(issues, "missing-metadata", `${path}.accessibility`, "is required");
  } else {
    validateStringList(
      issues,
      document.accessibility.requirements,
      `${path}.accessibility.requirements`,
    );
    validateStringList(
      issues,
      document.accessibility.keyboard,
      `${path}.accessibility.keyboard`,
    );
    validateStringList(
      issues,
      document.accessibility.knownConstraints,
      `${path}.accessibility.knownConstraints`,
    );
  }

  if (!document.api || typeof document.api !== "object") {
    addIssue(issues, "missing-metadata", `${path}.api`, "is required");
  } else {
    validateStringList(issues, document.api.react, `${path}.api.react`);
    validateStringList(issues, document.api.cssClasses, `${path}.api.cssClasses`);
    validateStringList(issues, document.api.publicTypes, `${path}.api.publicTypes`);
    validateStringList(issues, document.api.defaults, `${path}.api.defaults`);
    validateStringList(
      issues,
      document.api.invalidCombinations,
      `${path}.api.invalidCombinations`,
    );
  }

  if (!document.dependencies || typeof document.dependencies !== "object") {
    addIssue(issues, "missing-metadata", `${path}.dependencies`, "is required");
  } else {
    validateStringList(
      issues,
      document.dependencies.documents,
      `${path}.dependencies.documents`,
    );
    validateStringList(
      issues,
      document.dependencies.packages,
      `${path}.dependencies.packages`,
    );
  }

  if (!document.migration || typeof document.migration !== "object") {
    addIssue(issues, "missing-metadata", `${path}.migration`, "is required");
  } else {
    validateStringList(issues, document.migration.legacy, `${path}.migration.legacy`);
    validateStringList(issues, document.migration.notes, `${path}.migration.notes`);
  }

  if (!Array.isArray(document.changelog)) {
    addIssue(issues, "missing-metadata", `${path}.changelog`, "must be an array");
  } else {
    document.changelog.forEach((entry, entryIndex) => {
      const entryPath = `${path}.changelog[${entryIndex}]`;
      if (!isCalendarDate(entry?.date)) {
        addIssue(issues, "invalid-date", `${entryPath}.date`, "must be a real ISO date");
      }
      if (!isNonEmptyString(entry?.note)) {
        addIssue(issues, "missing-metadata", `${entryPath}.note`, "must be non-empty");
      }
    });
  }
}

export function validateStableDocument(
  document: CatalogDocument,
  path: string,
  issues: CatalogValidationIssue[],
): void {
  if (document.lifecycle !== "Stable") return;

  if (!Array.isArray(document.fixtureIds) || document.fixtureIds.length === 0) {
    addIssue(
      issues,
      "missing-stable-fixture",
      `${path}.fixtureIds`,
      "Stable documents must reference at least one deterministic fixture",
    );
  }
  if (
    !Array.isArray(document.behaviorTestIds) ||
    document.behaviorTestIds.length === 0
  ) {
    addIssue(
      issues,
      "missing-stable-behavior-test",
      `${path}.behaviorTestIds`,
      "Stable documents must identify rendered semantic or behavioral tests",
    );
  }
  if (
    !document.accessibility ||
    !isNonEmptyStringArray(document.accessibility.requirements)
  ) {
    addIssue(
      issues,
      "missing-stable-accessibility",
      `${path}.accessibility.requirements`,
      "Stable documents must state accessibility requirements",
    );
  }

  if (document.kind !== "component") return;
  if (!document.sections || typeof document.sections.required !== "object") {
    REQUIRED_COMPONENT_SECTION_KEYS.forEach((key) => {
      addIssue(
        issues,
        "missing-stable-section",
        `${path}.sections.required.${key}`,
        `Stable component is missing required section "${key}"`,
      );
    });
    return;
  }

  REQUIRED_COMPONENT_SECTION_KEYS.forEach((key) => {
    const section = document.sections?.required[key];
    if (
      !section ||
      !isNonEmptyString(section.heading) ||
      !isNonEmptyStringArray(section.content)
    ) {
      addIssue(
        issues,
        "missing-stable-section",
        `${path}.sections.required.${key}`,
        `Stable component section "${key}" needs a heading and content`,
      );
    }
  });

  const pageSections = document.sections.pageSections;
  if (!Array.isArray(pageSections) || pageSections.length === 0) {
    addIssue(
      issues,
      "missing-stable-section",
      `${path}.sections.pageSections`,
      "Stable component pages must opt into at least one useful public topic",
    );
  } else {
    const seenPageSections = new Set<string>();
    pageSections.forEach((key, index) => {
      if (!(COMPONENT_PAGE_SECTION_KEYS as readonly string[]).includes(key)) {
        addIssue(
          issues,
          "invalid-section",
          `${path}.sections.pageSections[${index}]`,
          `Unsupported public component topic "${key}"`,
        );
      } else if (seenPageSections.has(key)) {
        addIssue(
          issues,
          "invalid-section",
          `${path}.sections.pageSections[${index}]`,
          `Public component topic "${key}" is duplicated`,
        );
      }
      seenPageSections.add(key);
    });
  }

  Object.keys(document.sections.required).forEach((key) => {
    if (!(REQUIRED_COMPONENT_SECTION_KEYS as readonly string[]).includes(key)) {
      addIssue(
        issues,
        "invalid-section",
        `${path}.sections.required.${key}`,
        "Component-specific sections belong in sections.optional",
      );
    }
  });

  let previousAfter = -1;
  document.sections.optional?.forEach((section, sectionIndex) => {
    const after = REQUIRED_COMPONENT_SECTION_KEYS.indexOf(section.after);
    if (after < previousAfter) {
      addIssue(
        issues,
        "invalid-section",
        `${path}.sections.optional[${sectionIndex}].after`,
        "Optional sections must follow the required instructional sequence",
      );
    }
    previousAfter = after;
  });
}
