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
  { name: "name", type: "string", required: true, description: "Path in the enclosing useForm values. Initialize it through defaultValues, then use reset to load another record." },
  { name: "label", type: "string", required: true, description: "Persistent visible label; do not replace it with a placeholder." },
  { name: "hint", type: "string", description: "Help shown below the control and connected through aria-describedby. Optional hints do not shift adjacent controls out of alignment." },
  { name: "required", type: "boolean", defaultValue: "false (RadioGroup: true)", description: "Marks the decision as required and validates it through React Hook Form." },
];
export const validationProp: Prop = { name: "rules", type: "RegisterOptions", description: "React Hook Form validation, including custom recovery messages. Use noValidate on the form so inline errors handle submission." };
export const optionsProp: Prop = { name: "options", type: "readonly Option<string>[]", required: true, description: "Objects with value, label, and optional disabled. Persist value; show label to the user." };
export const disabledProp: Prop = { name: "disabled", type: "boolean", defaultValue: "false", description: "Prevents editing. Explain why the value is unavailable in nearby text." };
export const formSetup = "Import from paper-ui and load paper-ui/styles.css once at the application root. Wrap the form in FormProvider, initialize useForm defaultValues, and submit with methods.handleSubmit. The examples show validation, reset, and locally observable submitted state; replace the local confirmation with your save handler.";
