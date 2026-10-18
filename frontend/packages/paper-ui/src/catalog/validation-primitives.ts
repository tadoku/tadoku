import type { CatalogValidationIssue } from "./schema";

export const ID_PATTERN = /^[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
export const ROUTE_PATTERN = /^\/[a-z0-9]+(?:[/-][a-z0-9]+)*$/;

export function addIssue(
  issues: CatalogValidationIssue[],
  code: CatalogValidationIssue["code"],
  path: string,
  message: string,
): void {
  issues.push({ code, path, message });
}

export function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

export function hasDuplicate(values: readonly string[]): string | undefined {
  const seen = new Set<string>();
  return values.find((value) => {
    if (seen.has(value)) return true;
    seen.add(value);
    return false;
  });
}

export function validateStringList(
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
