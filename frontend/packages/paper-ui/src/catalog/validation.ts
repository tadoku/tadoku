import {
  CATALOG_CATEGORIES,
  CATALOG_DOCUMENT_KINDS,
  CATALOG_LIFECYCLES,
  type CatalogFixture,
  type CatalogRegistryInput,
  type CatalogValidationIssue,
  type CatalogValidationResult,
} from "./schema";
import {
  categoryForKindIsValid,
  expectedRoutePattern,
  validateDocumentMetadata,
  validateStableDocument,
} from "./document-validation";
import { validateFixtureInto } from "./fixture-validation";
import { addIssue, ID_PATTERN, isNonEmptyString, ROUTE_PATTERN } from "./validation-primitives";

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
