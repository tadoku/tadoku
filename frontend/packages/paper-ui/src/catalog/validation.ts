import {
  CATALOG_CATEGORIES,
  CATALOG_DOCUMENT_KINDS,
  CATALOG_LIFECYCLES,
  COMPONENT_PAGE_SECTION_KEYS,
  COMPONENT_CATEGORIES,
  FIXTURE_DENSITIES,
  FIXTURE_THEMES,
  REQUIRED_COMPONENT_SECTION_KEYS,
  type CatalogCategory,
  type CatalogDocument,
  type CatalogFixture,
  type CatalogKind,
  type CatalogRegistryInput,
  type CatalogValidationIssue,
  type CatalogValidationResult,
} from "./schema";

const ID_PATTERN = /^[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
const ROUTE_PATTERN = /^\/[a-z0-9]+(?:[/-][a-z0-9]+)*$/;
const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

const FIXTURE_NON_DETERMINISM_PATTERNS: ReadonlyArray<
  readonly [RegExp, string]
> = [
  [/\bMath\.random\s*\(/, "Math.random"],
  [/\bcrypto\.(?:randomUUID|getRandomValues)\s*\(/, "random crypto APIs"],
  [/\bDate\.now\s*\(/, "Date.now"],
  [/\bnew\s+Date\s*\(\s*\)/, "an unpinned Date"],
  [/\bperformance\.now\s*\(/, "performance.now"],
  [/\b(?:faker|Faker)\b/, "Faker"],
  [/\bfetch\s*\(/, "fetch"],
  [/\b(?:XMLHttpRequest|WebSocket|EventSource)\b/, "a network API"],
  [/\bnavigator\.onLine\b/, "network state"],
];

const FIXTURE_ROUTER_PATTERNS: ReadonlyArray<readonly [RegExp, string]> = [
  [/\buseRouter\s*\(/, "useRouter"],
  [/\b(?:next\/router|next\/navigation|react-router(?:-dom)?)\b/, "a router package"],
  [/\b(?:window|document)\.location\b/, "global location state"],
];

function addIssue(
  issues: CatalogValidationIssue[],
  code: CatalogValidationIssue["code"],
  path: string,
  message: string,
): void {
  issues.push({ code, path, message });
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function isNonEmptyStringArray(value: unknown): value is readonly string[] {
  return (
    Array.isArray(value) &&
    value.length > 0 &&
    value.every((item) => isNonEmptyString(item))
  );
}

function hasDuplicate(values: readonly string[]): string | undefined {
  const seen = new Set<string>();
  return values.find((value) => {
    if (seen.has(value)) return true;
    seen.add(value);
    return false;
  });
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

function categoryForKindIsValid(
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

function expectedRoutePattern(document: CatalogDocument): RegExp {
  switch (document.kind) {
    case "foundation":
      return /^\/foundations\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "component":
      return /^\/components\/(?:actions|forms|navigation|feedback|overlays|data-display)\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "pattern":
      return /^\/patterns\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "experiment":
      return /^\/experiments\/[a-z0-9]+(?:-[a-z0-9]+)*$/;
    case "governance":
      return /^\/(?:contributing|changelog)$/;
  }
}

function validateStringList(
  issues: CatalogValidationIssue[],
  value: unknown,
  path: string,
  { required = false }: { readonly required?: boolean } = {},
): void {
  if (!Array.isArray(value) || !value.every((item) => isNonEmptyString(item))) {
    addIssue(issues, "missing-metadata", path, "must be an array of non-empty strings");
    return;
  }
  if (required && value.length === 0) {
    addIssue(issues, "missing-metadata", path, "must contain at least one entry");
  }
  const duplicate = hasDuplicate(value);
  if (duplicate !== undefined) {
    addIssue(
      issues,
      "duplicate-reference",
      path,
      `contains duplicate entry "${duplicate}"`,
    );
  }
}

function validateDocumentMetadata(
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

function validateStableDocument(
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

function validateFixtureInto(
  fixture: CatalogFixture,
  path: string,
  issues: CatalogValidationIssue[],
): void {
  if (!fixture || typeof fixture !== "object") {
    addIssue(issues, "invalid-fixture", path, "must be an object");
    return;
  }
  if (!isNonEmptyString(fixture.id) || !ID_PATTERN.test(fixture.id)) {
    addIssue(issues, "invalid-id", `${path}.id`, "must be a stable lowercase ID");
  }
  (["name", "description"] as const).forEach((key) => {
    if (!isNonEmptyString(fixture[key])) {
      addIssue(issues, "missing-metadata", `${path}.${key}`, "must be non-empty");
    }
  });
  validateStringList(issues, fixture.tags, `${path}.tags`, { required: true });

  if (
    !Array.isArray(fixture.themes) ||
    fixture.themes.length === 0 ||
    fixture.themes.some(
      (theme) => !(FIXTURE_THEMES as readonly unknown[]).includes(theme),
    )
  ) {
    addIssue(issues, "invalid-fixture", `${path}.themes`, "contains an unsupported theme");
  } else if (hasDuplicate(fixture.themes)) {
    addIssue(issues, "duplicate-reference", `${path}.themes`, "contains a duplicate theme");
  }
  if (
    !Array.isArray(fixture.densities) ||
    fixture.densities.length === 0 ||
    fixture.densities.some(
      (density) => !(FIXTURE_DENSITIES as readonly unknown[]).includes(density),
    )
  ) {
    addIssue(
      issues,
      "invalid-fixture",
      `${path}.densities`,
      "contains an unsupported density",
    );
  } else if (hasDuplicate(fixture.densities)) {
    addIssue(
      issues,
      "duplicate-reference",
      `${path}.densities`,
      "contains a duplicate density",
    );
  }

  if (!Array.isArray(fixture.viewports) || fixture.viewports.length === 0) {
    addIssue(issues, "invalid-fixture", `${path}.viewports`, "must not be empty");
  } else {
    const viewportIds: string[] = [];
    fixture.viewports.forEach((viewport, viewportIndex) => {
      const viewportPath = `${path}.viewports[${viewportIndex}]`;
      if (!isNonEmptyString(viewport?.id) || !ID_PATTERN.test(viewport.id)) {
        addIssue(issues, "invalid-id", `${viewportPath}.id`, "must be a stable lowercase ID");
      } else {
        viewportIds.push(viewport.id);
      }
      if (!isNonEmptyString(viewport?.label)) {
        addIssue(issues, "missing-metadata", `${viewportPath}.label`, "must be non-empty");
      }
      if (!Number.isInteger(viewport?.width) || viewport.width <= 0) {
        addIssue(issues, "invalid-fixture", `${viewportPath}.width`, "must be a positive integer");
      }
      if (!Number.isInteger(viewport?.height) || viewport.height <= 0) {
        addIssue(issues, "invalid-fixture", `${viewportPath}.height`, "must be a positive integer");
      }
    });
    const duplicateViewport = hasDuplicate(viewportIds);
    if (duplicateViewport) {
      addIssue(
        issues,
        "duplicate-reference",
        `${path}.viewports`,
        `contains duplicate viewport "${duplicateViewport}"`,
      );
    }
  }

  if (fixture.deterministic !== true) {
    addIssue(
      issues,
      "invalid-fixture-marker",
      `${path}.deterministic`,
      "must be explicitly marked true after deterministic review",
    );
  }
  if (typeof fixture.render !== "function") {
    addIssue(issues, "invalid-fixture", `${path}.render`, "must be a render function");
    return;
  }

  const source = Function.prototype.toString.call(fixture.render);
  FIXTURE_NON_DETERMINISM_PATTERNS.forEach(([pattern, label]) => {
    if (pattern.test(source)) {
      addIssue(
        issues,
        "non-deterministic-fixture",
        `${path}.render`,
        `must not read ${label}; pass fixed values through the fixture instead`,
      );
    }
  });
  FIXTURE_ROUTER_PATTERNS.forEach(([pattern, label]) => {
    if (pattern.test(source)) {
      addIssue(
        issues,
        "router-dependent-fixture",
        `${path}.render`,
        `must not read ${label}; fixtures are router-neutral`,
      );
    }
  });
}

export function validateCatalogFixture(
  fixture: CatalogFixture,
): CatalogValidationResult {
  const issues: CatalogValidationIssue[] = [];
  validateFixtureInto(fixture, "fixture", issues);
  return { valid: issues.length === 0, issues };
}

export function validateCatalogRegistry(
  registry: CatalogRegistryInput,
): CatalogValidationResult {
  const issues: CatalogValidationIssue[] = [];
  if (!registry || typeof registry !== "object") {
    return {
      valid: false,
      issues: [{ code: "invalid-registry", path: "registry", message: "must be an object" }],
    };
  }
  if (!Array.isArray(registry.documents) || !Array.isArray(registry.fixtures)) {
    return {
      valid: false,
      issues: [
        {
          code: "invalid-registry",
          path: "registry",
          message: "documents and fixtures must be arrays",
        },
      ],
    };
  }

  const documentIds = new Map<string, number>();
  const fixtureIds = new Map<string, number>();
  const routes = new Map<string, string>();
  const slugs = new Map<string, string>();
  const aliases = new Map<string, string>();

  registry.fixtures.forEach((fixture, index) => {
    const path = `fixtures[${index}]`;
    validateFixtureInto(fixture, path, issues);
    const previous = fixtureIds.get(fixture?.id);
    if (previous !== undefined) {
      addIssue(
        issues,
        "duplicate-fixture-id",
        `${path}.id`,
        `duplicates fixtures[${previous}].id`,
      );
    } else if (isNonEmptyString(fixture?.id)) {
      fixtureIds.set(fixture.id, index);
    }
  });

  registry.documents.forEach((document, index) => {
    const path = `documents[${index}]`;
    if (!document || typeof document !== "object") {
      addIssue(issues, "invalid-document", path, "must be an object");
      return;
    }
    if (!isNonEmptyString(document.id) || !ID_PATTERN.test(document.id)) {
      addIssue(issues, "invalid-id", `${path}.id`, "must be a stable lowercase ID");
    } else {
      const previous = documentIds.get(document.id);
      if (previous !== undefined) {
        addIssue(
          issues,
          "duplicate-document-id",
          `${path}.id`,
          `duplicates documents[${previous}].id`,
        );
      } else {
        documentIds.set(document.id, index);
      }
    }

    if (!(CATALOG_DOCUMENT_KINDS as readonly unknown[]).includes(document.kind)) {
      addIssue(issues, "invalid-kind", `${path}.kind`, "contains an unsupported kind");
    }
    if (!(CATALOG_LIFECYCLES as readonly unknown[]).includes(document.lifecycle)) {
      addIssue(
        issues,
        "invalid-lifecycle",
        `${path}.lifecycle`,
        "must be Experimental, Stable, or Deprecated",
      );
    }
    if (!(CATALOG_CATEGORIES as readonly unknown[]).includes(document.category)) {
      addIssue(issues, "invalid-category", `${path}.category`, "is not a catalogue category");
    } else if (
      (CATALOG_DOCUMENT_KINDS as readonly unknown[]).includes(document.kind) &&
      !categoryForKindIsValid(document.kind, document.category)
    ) {
      addIssue(
        issues,
        "invalid-category",
        `${path}.category`,
        `category "${document.category}" is not valid for kind "${document.kind}"`,
      );
    }

    if (
      !isNonEmptyString(document.route) ||
      !ROUTE_PATTERN.test(document.route) ||
      ((CATALOG_DOCUMENT_KINDS as readonly unknown[]).includes(document.kind) &&
        !expectedRoutePattern(document).test(document.route))
    ) {
      addIssue(
        issues,
        "invalid-route",
        `${path}.route`,
        "does not match the canonical route for its kind and category",
      );
    } else {
      const previousRoute = routes.get(document.route);
      if (previousRoute) {
        addIssue(
          issues,
          "duplicate-route",
          `${path}.route`,
          `duplicates canonical route owned by "${previousRoute}"`,
        );
      } else {
        routes.set(document.route, document.id);
      }
      const routeParts = document.route.split("/");
      const slug = routeParts[routeParts.length - 1] ?? "";
      const previousSlug = slugs.get(slug);
      if (previousSlug) {
        addIssue(
          issues,
          "duplicate-slug",
          `${path}.route`,
          `slug "${slug}" is already owned by "${previousSlug}"`,
        );
      } else {
        slugs.set(slug, document.id);
      }
    }

    validateDocumentMetadata(document, path, issues);
    validateStableDocument(document, path, issues);

    document.aliases?.forEach((alias: string, aliasIndex: number) => {
      const aliasPath = `${path}.aliases[${aliasIndex}]`;
      if (!ROUTE_PATTERN.test(alias)) {
        addIssue(issues, "invalid-route", aliasPath, "must be an absolute normalized route");
      }
      const owner = aliases.get(alias) ?? routes.get(alias);
      if (owner) {
        addIssue(issues, "duplicate-alias", aliasPath, `route is already owned by "${owner}"`);
      } else {
        aliases.set(alias, document.id);
      }
    });
  });

  registry.documents.forEach((document, documentIndex) => {
    if (!document || typeof document !== "object") return;
    document.fixtureIds?.forEach((fixtureId: string, fixtureIndex: number) => {
      if (!fixtureIds.has(fixtureId)) {
        addIssue(
          issues,
          "unknown-fixture-reference",
          `documents[${documentIndex}].fixtureIds[${fixtureIndex}]`,
          `fixture "${fixtureId}" is not registered`,
        );
      }
    });
    document.dependencies?.documents?.forEach(
      (documentId: string, dependencyIndex: number) => {
      if (!documentIds.has(documentId)) {
        addIssue(
          issues,
          "unknown-document-reference",
          `documents[${documentIndex}].dependencies.documents[${dependencyIndex}]`,
          `document "${documentId}" is not registered`,
        );
      }
      },
    );

    if (document.lifecycle === "Deprecated") {
      const replacementId = document.migration?.replacementDocumentId;
      if (!isNonEmptyString(replacementId)) {
        addIssue(
          issues,
          "invalid-deprecation",
          `documents[${documentIndex}].migration.replacementDocumentId`,
          "Deprecated documents must name a replacement document",
        );
      } else if (replacementId === document.id || !documentIds.has(replacementId)) {
        addIssue(
          issues,
          "unknown-document-reference",
          `documents[${documentIndex}].migration.replacementDocumentId`,
          "replacement must reference a different registered document",
        );
      }
      if (!isNonEmptyString(document.migration?.removalGuidance)) {
        addIssue(
          issues,
          "invalid-deprecation",
          `documents[${documentIndex}].migration.removalGuidance`,
          "Deprecated documents must include planned removal guidance",
        );
      }
    }
  });

  aliases.forEach((owner, alias) => {
    const canonicalOwner = routes.get(alias);
    if (canonicalOwner && canonicalOwner !== owner) {
      addIssue(
        issues,
        "duplicate-alias",
        `documents[${documentIds.get(owner) ?? 0}].aliases`,
        `alias "${alias}" collides with canonical route owned by "${canonicalOwner}"`,
      );
    }
  });

  const redirects = registry.redirects ?? [];
  if (!Array.isArray(redirects)) {
    addIssue(issues, "invalid-registry", "redirects", "must be an array when provided");
  } else {
    const redirectSources = new Set<string>();
    redirects.forEach((redirect, index) => {
      const path = `redirects[${index}]`;
      if (!redirect || typeof redirect !== "object") {
        addIssue(issues, "invalid-redirect", path, "must be an object");
        return;
      }
      if (!ROUTE_PATTERN.test(redirect.from) || !ROUTE_PATTERN.test(redirect.to)) {
        addIssue(issues, "invalid-redirect", path, "from and to must be normalized routes");
      }
      if (redirect.from === redirect.to) {
        addIssue(issues, "invalid-redirect", path, "must not redirect a route to itself");
      }
      if (routes.has(redirect.from) || aliases.has(redirect.from) || redirectSources.has(redirect.from)) {
        addIssue(issues, "duplicate-alias", `${path}.from`, "redirect source is already owned");
      }
      if (!routes.has(redirect.to)) {
        addIssue(
          issues,
          "unknown-document-reference",
          `${path}.to`,
          "redirect target must be a canonical document route",
        );
      }
      redirectSources.add(redirect.from);
    });
  }

  return { valid: issues.length === 0, issues };
}
