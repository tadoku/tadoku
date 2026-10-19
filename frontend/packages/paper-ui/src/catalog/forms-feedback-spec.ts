import type {
  CatalogDocument,
  ComponentCategory,
  ComponentPageSectionKey,
} from "./schema";

export interface DocSpec {
  readonly id: string;
  readonly route: string;
  readonly name: string;
  readonly category: ComponentCategory;
  readonly summary: string;
  readonly fixtureId: string;
  readonly sourcePath: string;
  readonly react: readonly string[];
  readonly types: readonly string[];
  readonly css?: readonly string[];
  readonly when: string;
  readonly avoid: string;
  readonly choose: string;
  readonly behavior: string;
  readonly accessibility: string;
  readonly migration: string;
  readonly variants?: string;
  readonly states?: string;
  readonly content?: string;
  readonly pageSections?: readonly ComponentPageSectionKey[];
}

export type Prop = NonNullable<CatalogDocument["api"]["props"]>[number];
export type ComponentDetails = Record<string, { example: string; anatomy: string; variants: string; states: string; content: string; implementation: string; props: readonly Prop[] }>;

export const fieldProps: readonly Prop[] = [
  { name: "name", type: "string", required: true, description: "Pass name to form.AppField, not the bound control. Initialize it through usePaperForm defaultValues." },
  { name: "label", type: "string", required: true, description: "Persistent visible label; do not replace it with a placeholder." },
  { name: "hint", type: "string", description: "Help shown below the control and connected through aria-describedby. Optional hints do not shift adjacent controls out of alignment." },
  { name: "required", type: "boolean", defaultValue: "false (RadioGroup: true)", description: "Marks the native control as required; add the required rule to the AppField or form validators." },
];
export const validationProp: Prop = { name: "validators (AppField)", type: "FieldValidators", description: "TanStack validation and recovery messages belong on form.AppField or the form. form.Form submits through handleSubmit and focuses errors." };
export const optionsProp: Prop = { name: "options", type: "readonly Option<string>[]", required: true, description: "Objects with value, label, and optional disabled. Persist value; show label to the user." };
export const disabledProp: Prop = { name: "disabled", type: "boolean", defaultValue: "false", description: "Prevents editing. Explain why the value is unavailable in nearby text." };
export const formSetup = "Import usePaperForm from paper-ui. Initialize typed defaultValues, validators and onSubmit. Wrap form.Form in form.AppForm; bind each named form.AppField to field.Input or another bound control. required and numeric bounds need matching validators. Use reset(values, { keepDefaultValues: true }) to load another draft. The screen owns persistence and server errors.";
