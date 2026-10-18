import {
defineCatalogDocument,
type CatalogDocument,
type ComponentDocumentationSections,
type RequiredComponentSections,
} from "./schema";
import type { ComponentDetails, DocSpec } from "./forms-feedback-spec";
import {
  buttonGroupAndSurfaceCommonMistakes,
  buttonGroupAndSurfaceDetails,
  buttonGroupFixtures,
  buttonGroupSpec,
  surfaceFixtures,
  surfaceSpec,
} from "./button-group-and-surface";
import {
  feedbackCommonMistakes,
  feedbackDetails,
  flashFixtures,
  flashSpec,
  loadingFixtures,
  loadingSpec,
  toastFixtures,
  toastSpec,
} from "./feedback-components";
import {
  autocompleteCommonMistakes,
  autocompleteDetails,
  autocompleteFixtures,
  autocompleteSpec,
  multiAutocompleteFixtures,
  multiAutocompleteSpec,
  tagsInputFixtures,
  tagsInputSpec,
} from "./form-autocomplete";
import {
  choiceGroupCommonMistakes,
  choiceGroupDetails,
  radioGroupFixtures,
  radioGroupSpec,
  radioSelectFixtures,
  radioSelectNativeFixtures,
  radioSelectSpec,
  toggleSelectFixtures,
  toggleSelectSpec,
} from "./form-choice-groups";
import {
  amountWithUnitFixtures,
  amountWithUnitSpec,
  checkboxFixtures,
  checkboxSpec,
  formFieldCommonMistakes,
  formFieldDetails,
  selectFixtures,
  selectSpec,
  textAreaFixtures,
  textAreaSpec,
} from "./form-fields";

const REVIEW_DATE = "2026-09-05";

const details: ComponentDetails = {
  ...formFieldDetails,
  ...choiceGroupDetails,
  ...autocompleteDetails,
  ...buttonGroupAndSurfaceDetails,
  ...feedbackDetails,
};

function sections(spec: DocSpec): ComponentDocumentationSections {
  const required: RequiredComponentSections = {
    overview: { heading: "Overview", content: [spec.summary] },
    whenToUse: { heading: "When to use", content: [spec.when] },
    whenNotToUse: { heading: "When not to use", content: [spec.avoid] },
    choosingBetween: { heading: "Choose between", content: [spec.choose] },
    anatomy: { heading: "Anatomy", content: [details[spec.id].anatomy] },
    recommendedExample: { heading: "Recommended example", content: [details[spec.id].example] },
    variants: { heading: "Variants", content: [spec.variants ?? details[spec.id].variants] },
    statesAndAdaptation: { heading: "States and adaptation", content: [spec.states ?? details[spec.id].states] },
    behavior: { heading: "Behavior", content: [spec.behavior] },
    contentGuidance: { heading: "Content guidance", content: [spec.content ?? details[spec.id].content] },
    accessibility: { heading: "Accessibility", content: [spec.accessibility] },
    implementation: { heading: "Implementation", content: [details[spec.id].implementation] },
    apiReference: { heading: "API reference", content: [`Public React APIs: ${spec.react.join(", ")}. Public types: ${spec.types.join(", ") || "none"}.`] },
    relatedPatterns: { heading: "Related patterns", content: [`Combine ${spec.name} with Button and the nearest documented form, feedback, or action pattern rather than inventing local structure.`] },
    migration: { heading: "Migration", content: [spec.migration] },
    lifecycle: { heading: "Lifecycle", content: ["Published in Paper 0.1.0. Review the documented limits before application migration; Stable describes API availability, not exhaustive product coverage."] },
  };
  return {
    required,
    pageSections: spec.pageSections ?? ["usage", "examples", "variantsAndStates", "behavior", "contentGuidance", "accessibility"],
  };
}

const commonMistakes: Record<string, string> = {
  ...formFieldCommonMistakes,
  ...choiceGroupCommonMistakes,
  ...autocompleteCommonMistakes,
  ...buttonGroupAndSurfaceCommonMistakes,
  ...feedbackCommonMistakes,
};

function document(spec: DocSpec): CatalogDocument {
  return defineCatalogDocument({
    id: spec.id,
    route: spec.route,
    name: spec.name,
    kind: "component",
    category: spec.category,
    aliases: [],
    summary: spec.summary,
    keywords: [spec.name.toLocaleLowerCase(), spec.category, "paper"],
    lifecycle: "Stable",
    reviewDate: REVIEW_DATE,
    sourcePath: spec.sourcePath,
    packageVersion: "0.1.0",
    guidance: {
      whenToUse: [spec.when],
      whenNotToUse: [spec.avoid],
      content: [spec.content ?? details[spec.id].content],
      commonMistakes: [commonMistakes[spec.id]],
    },
    accessibility: {
      requirements: [spec.accessibility],
      keyboard: [spec.behavior],
      knownConstraints: [details[spec.id].states],
    },
    api: {
      react: spec.react,
      props: details[spec.id].props,
      cssClasses: spec.css ?? [],
      publicTypes: spec.types,
      defaults: ["Theme and density inherit from the nearest Paper root."],
      invalidCombinations: [spec.avoid],
    },
    fixtureIds: [spec.fixtureId, ...(spec.category === "forms" ? [`${spec.fixtureId}.empty`] : []), ...(spec.id === "component.radio-select" ? ["radio-select.native"] : [])],
    dependencies: { documents: ["component.button"], packages: spec.category === "forms" ? ["paper-ui", "react-hook-form"] : ["paper-ui"] },
    migration: { legacy: [`ui ${spec.name}`], notes: [spec.migration] },
    changelog: [{ date: REVIEW_DATE, note: `Revised ${spec.name} with complete runnable examples, state guidance, and concrete prop contracts.` }],
    behaviorTestIds: [`${spec.id}.semantics`, `${spec.id}.behavior`],
    sections: sections(spec),
  });
}

export const phaseThreeFormsFeedbackFixtures = [
  ...toggleSelectFixtures,
  ...textAreaFixtures,
  ...selectFixtures,
  ...checkboxFixtures,
  ...radioSelectFixtures,
  ...radioGroupFixtures,
  ...amountWithUnitFixtures,
  ...autocompleteFixtures,
  ...multiAutocompleteFixtures,
  ...tagsInputFixtures,
  ...radioSelectNativeFixtures,
  ...buttonGroupFixtures,
  ...toastFixtures,
  ...flashFixtures,
  ...loadingFixtures,
  ...surfaceFixtures,
] as const;

const specs: readonly DocSpec[] = [
  toggleSelectSpec,
  textAreaSpec,
  selectSpec,
  checkboxSpec,
  radioSelectSpec,
  radioGroupSpec,
  amountWithUnitSpec,
  autocompleteSpec,
  multiAutocompleteSpec,
  tagsInputSpec,
  buttonGroupSpec,
  flashSpec,
  loadingSpec,
  surfaceSpec,
  toastSpec,
];

export const phaseThreeFormsFeedbackDocuments = specs.map(document);
