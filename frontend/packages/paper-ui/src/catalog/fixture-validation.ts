import {
  FIXTURE_DENSITIES,
  FIXTURE_THEMES,
  type CatalogFixture,
  type CatalogValidationIssue,
} from "./schema";
import {
  addIssue,
  hasDuplicate,
  ID_PATTERN,
  isNonEmptyString,
  validateStringList,
} from "./validation-primitives";

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

export function validateFixtureInto(
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
