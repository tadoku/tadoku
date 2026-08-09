import {
  CatalogValidationError,
  type CatalogRegistry,
  type CatalogRegistryInput,
} from "./schema";
import { validateCatalogRegistry } from "./validation";
import { phaseTwoDocuments, phaseTwoFixtures } from "./phase-two";
import {
  phaseThreeFormsFeedbackDocuments,
  phaseThreeFormsFeedbackFixtures,
} from "./phase-three-forms-feedback";
import {
  legacyStyleguideRedirects,
  phaseThreeContentFixtures,
  phaseThreeExperimentDocuments,
  phaseThreeFoundationDocuments,
  phaseThreeGovernanceDocuments,
  phaseThreePatternDocuments,
} from "./phase-three-content";
import {
  phaseThreeDataLayoutDocuments,
  phaseThreeDataLayoutFixtures,
} from "./phase-three-data-layout";
import {
  phaseThreeNavigationDocuments,
  phaseThreeNavigationFixtures,
} from "./phase-three-navigation";
import {
  publishedPrimitiveDocuments,
  publishedPrimitiveFixtures,
} from "./published-primitives";

export const foundationDocuments = phaseThreeFoundationDocuments;
export const governanceDocuments = phaseThreeGovernanceDocuments;

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
  documents: [
    ...foundationDocuments,
    ...phaseTwoDocuments,
    ...phaseThreeFormsFeedbackDocuments,
    ...phaseThreeNavigationDocuments,
    ...publishedPrimitiveDocuments,
    ...phaseThreeDataLayoutDocuments,
    ...phaseThreePatternDocuments,
    ...phaseThreeExperimentDocuments,
    ...governanceDocuments,
  ],
  fixtures: [
    ...phaseTwoFixtures,
    ...phaseThreeFormsFeedbackFixtures,
    ...phaseThreeNavigationFixtures,
    ...publishedPrimitiveFixtures,
    ...phaseThreeDataLayoutFixtures,
    ...phaseThreeContentFixtures,
  ],
  redirects: legacyStyleguideRedirects,
});
