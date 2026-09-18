import { Combobox } from "@base-ui/react/combobox";
import {
  useId,
  useMemo,
  useRef,
  useState,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";
import {
  useController,
  useFormContext,
  useFormState,
  type FieldValues,
  type RegisterOptions,
} from "react-hook-form";
import { CheckIcon, ChevronDownIcon, XMarkIcon, iconClassName } from "../../icons";

export interface Option<Value = string> {
  readonly value: Value;
  readonly label: string;
  readonly description?: string;
  readonly disabled?: boolean;
}

export interface OptionGroup<Value = string> {
  readonly label: string;
  readonly options: readonly Option<Value>[];
}

interface FieldContract {
  readonly name: string;
  readonly label: string;
  readonly hint?: string;
  readonly required?: boolean;
  readonly rules?: RegisterOptions<FieldValues, string>;
}

function useField(name: string) {
  const form = useFormContext();
  if (!form) throw new Error("Paper form controls require react-hook-form FormProvider");
  const formState = useFormState({ control: form.control, name });
  const error = form.getFieldState(name, formState).error;
  return { form, error };
}

function FieldFrame({
  id,
  label,
  hint,
  error,
  required,
  children,
}: {
  readonly id: string;
  readonly label: string;
  readonly hint?: string;
  readonly error?: string;
  readonly required?: boolean;
  readonly children: ReactNode;
}) {
  return (
    <div className={`paper-field${error ? " paper-field--invalid" : ""}`}>
      <label className="paper-field__label" htmlFor={id} id={`${id}-label`}>
        {label}
        {required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}
      </label>
      {children}
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error}</p> : null}
    </div>
  );
}

function describedBy(id: string, hint?: string, error?: unknown, external?: string): string | undefined {
  return [hint ? `${id}-hint` : undefined, error ? `${id}-error` : undefined, external]
    .filter(Boolean)
    .join(" ") || undefined;
}

export interface TextAreaProps
  extends FieldContract,
    Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, "name" | "onChange" | "onBlur" | "required"> {}

export function TextArea({
  name,
  label,
  hint,
  rules,
  required,
  id: providedId,
  ...props
}: TextAreaProps) {
  const id = providedId ?? `paper-textarea-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "This field is required." : undefined),
  });
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <textarea
        {...props}
        {...registration}
        id={id}
        required={required}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        className="paper-input paper-textarea"
      />
    </FieldFrame>
  );
}

export interface SelectProps
  extends FieldContract,
    Omit<SelectHTMLAttributes<HTMLSelectElement>, "name" | "onChange" | "onBlur" | "required"> {
  readonly options: readonly Option<string>[];
  readonly groups?: readonly OptionGroup<string>[];
  readonly placeholder?: string;
}

export function Select({
  name,
  label,
  hint,
  rules,
  required,
  options,
  groups,
  placeholder,
  id: providedId,
  ...props
}: SelectProps) {
  const id = providedId ?? `paper-select-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "Choose an option." : undefined),
  });
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <span className="paper-select__control">
        <select
          {...props}
          {...registration}
          id={id}
          required={required}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
          className="paper-input paper-select"
        >
          {placeholder ? <option value="">{placeholder}</option> : null}
          {groups
            ? groups.map((group) => (
                <optgroup key={group.label} label={group.label}>
                  {group.options.map((option) => (
                    <option key={option.value} value={option.value} disabled={option.disabled}>
                      {option.label}
                    </option>
                  ))}
                </optgroup>
              ))
            : options.map((option) => (
                <option key={option.value} value={option.value} disabled={option.disabled}>
                  {option.label}
                </option>
              ))}
        </select>
        <ChevronDownIcon
          className={iconClassName("compact", "paper-select__icon")}
          aria-hidden="true"
        />
      </span>
    </FieldFrame>
  );
}

export interface CheckboxProps
  extends FieldContract,
    Omit<InputHTMLAttributes<HTMLInputElement>, "name" | "type" | "onChange" | "onBlur" | "required"> {}

export function Checkbox({
  name,
  label,
  hint,
  rules,
  required,
  id: providedId,
  ...props
}: CheckboxProps) {
  const id = providedId ?? `paper-checkbox-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: props.disabled,
    required: rules?.required ?? (required ? "Select this option to continue." : undefined),
  });
  return (
    <fieldset className={`paper-choice-field${error ? " paper-field--invalid" : ""}`}>
      <label className="paper-choice" htmlFor={id}>
        <input
          {...props}
          {...registration}
          id={id}
          type="checkbox"
          required={required}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        />
        <span>{label}</span>
      </label>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error.message?.toString()}</p> : null}
    </fieldset>
  );
}

export const RADIO_SELECT_VARIANTS = ["default", "segmented"] as const;

export type RadioSelectVariant = (typeof RADIO_SELECT_VARIANTS)[number];

export interface RadioSelectProps extends FieldContract {
  readonly options: readonly Option<string>[];
  readonly disabled?: boolean;
  readonly variant?: RadioSelectVariant;
}

export function RadioSelect({
  name,
  label,
  hint,
  rules,
  required,
  options,
  disabled,
  variant = "default",
}: RadioSelectProps) {
  const id = `paper-radio-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    disabled: disabled,
    required: rules?.required ?? (required ? "Choose an option." : undefined),
  });
  const segmented = variant === "segmented";
  return (
    <fieldset
      className={`paper-choice-field paper-radio-select paper-radio-select--${variant}${error ? " paper-field--invalid" : ""}`}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
      disabled={disabled}
    >
      <legend className="paper-field__label">{label}{required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}</legend>
      <div className="paper-choice-list">
        {options.map((option) => (
          <label className={segmented ? "paper-radio-select__segment" : "paper-choice"} key={option.value}>
            <input
              {...registration}
              type="radio"
              value={option.value}
              disabled={option.disabled}
              required={required}
              aria-invalid={error ? true : undefined}
            />
            <span className={segmented ? "paper-radio-select__label" : undefined}>{option.label}</span>
            {segmented ? (
              <CheckIcon
                className={iconClassName("compact", "paper-radio-select__check")}
                aria-hidden="true"
              />
            ) : null}
          </label>
        ))}
      </div>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error.message?.toString()}</p> : null}
    </fieldset>
  );
}

export interface RadioGroupOption<Value = string> extends Option<Value> {
  readonly description: string;
}

export interface RadioGroupProps<Value = string>
  extends Omit<FieldContract, "rules"> {
  readonly options: readonly RadioGroupOption<Value>[];
  readonly rules?: RegisterOptions<FieldValues, string>;
}

export function RadioGroup<Value extends string = string>({
  name,
  label,
  hint,
  required = true,
  rules,
  options,
}: RadioGroupProps<Value>) {
  const id = `paper-radio-group-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const registration = form.register(name, {
    ...rules,
    required: rules?.required ?? (required ? "Choose an option." : undefined),
  });
  return (
    <fieldset
      className={`paper-choice-field${error ? " paper-field--invalid" : ""}`}
      aria-describedby={describedBy(id, hint, error)}
      aria-invalid={error ? true : undefined}
    >
      <legend className="paper-field__label">{label}{required ? <span className="paper-field__required" aria-hidden="true">*</span> : null}</legend>
      <div className="paper-radio-cards">
        {options.map((option) => (
          <label className="paper-radio-card" key={option.value}>
            <input
              {...registration}
              type="radio"
              value={option.value}
              required={required}
              disabled={option.disabled}
              aria-invalid={error ? true : undefined}
            />
            <span><strong>{option.label}</strong><small>{option.description}</small></span>
          </label>
        ))}
      </div>
      {hint ? <p className="paper-field__hint" id={`${id}-hint`}>{hint}</p> : null}
      {error ? <p className="paper-field__error" id={`${id}-error`} role="alert">{error.message?.toString()}</p> : null}
    </fieldset>
  );
}

export interface AmountWithUnitProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "name" | "type"> {
  readonly name: string;
  readonly label: string;
  readonly units: readonly Option<string>[];
  readonly unitsLabel?: string;
  readonly hint?: string;
}

export function AmountWithUnit({
  name,
  label,
  units,
  unitsLabel,
  hint,
  required,
  min,
  max,
  disabled,
  ...props
}: AmountWithUnitProps) {
  const id = `paper-amount-${useId().replace(/:/gu, "")}`;
  const amountName = `${name}Value`;
  const unitName = `${name}Unit`;
  const { form, error: amountError } = useField(amountName);
  const unitState = useFormState({ control: form.control, name: unitName });
  const unitError = form.getFieldState(unitName, unitState).error;
  const error = amountError ?? unitError;
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <div className="paper-compound-field" role="group" aria-labelledby={`${id}-label`}>
        <input
          {...props}
          {...form.register(amountName, {
            valueAsNumber: true,
            disabled,
            required: required ? "Enter an amount." : undefined,
            min: min === undefined ? undefined : { value: min, message: `Enter at least ${min}.` },
            max: max === undefined ? undefined : { value: max, message: `Enter no more than ${max}.` },
          })}
          required={required}
          min={min}
          max={max}
          disabled={disabled}
          id={id}
          type="number"
          className="paper-input"
          aria-invalid={amountError ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        />
        <select
          {...form.register(unitName, { disabled })}
          disabled={disabled}
          aria-label={unitsLabel ?? `Unit for ${label.toLocaleLowerCase()}`}
          className="paper-input paper-compound-field__unit"
          aria-invalid={unitError ? true : undefined}
          aria-describedby={describedBy(id, hint, error, props["aria-describedby"])}
        >
          {units.map((unit) => <option key={unit.value} value={unit.value} disabled={unit.disabled}>{unit.label}</option>)}
        </select>
      </div>
    </FieldFrame>
  );
}

interface AutocompleteContract<Value> extends FieldContract {
  readonly options: readonly Value[];
  readonly format: (option: Value) => string;
  readonly getId: (option: Value) => string;
  readonly match?: (option: Value, query: string) => boolean;
  readonly maxResults?: number;
  readonly disabled?: boolean;
  readonly placeholder?: string;
}

function useFilteredOptions<Value>(
  options: readonly Value[],
  query: string,
  format: (option: Value) => string,
  match?: (option: Value, query: string) => boolean,
  maxResults = 50,
) {
  return useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase();
    return options
      .filter((option) => !normalized || (match ? match(option, query) : format(option).toLocaleLowerCase().includes(normalized)))
      .slice(0, maxResults);
  }, [format, match, maxResults, options, query]);
}

function ComboboxPopup<Value>({
  options,
  getId,
  format,
  portalContainer,
}: {
  readonly options: readonly Value[];
  readonly getId: (option: Value) => string;
  readonly format: (option: Value) => string;
  readonly portalContainer: HTMLElement | null;
}) {
  return (
    <Combobox.Portal container={portalContainer}>
      <Combobox.Positioner className="paper-combobox__positioner" sideOffset={4}>
        <Combobox.Popup className="paper-combobox__popup paper-elevation-floating">
          <Combobox.Empty className="paper-combobox__empty">No matching options.</Combobox.Empty>
          <Combobox.List>
            {options.map((option, index) => (
              <Combobox.Item
                key={getId(option)}
                value={option}
                index={index}
                className="paper-combobox__item"
              >
                <Combobox.ItemIndicator className="paper-combobox__indicator" aria-hidden="true">
                  <CheckIcon className={iconClassName("compact")} />
                </Combobox.ItemIndicator>
                <span>{format(option)}</span>
              </Combobox.Item>
            ))}
          </Combobox.List>
        </Combobox.Popup>
      </Combobox.Positioner>
    </Combobox.Portal>
  );
}

export type AutocompleteInputProps<Value> = AutocompleteContract<Value>;

export function AutocompleteInput<Value>({
  name,
  label,
  hint,
  required,
  rules,
  options,
  format,
  getId,
  match,
  maxResults,
  disabled,
  placeholder,
}: AutocompleteInputProps<Value>) {
  const id = `paper-autocomplete-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const { field } = useController({ name, control: form.control, rules: { ...rules, required: rules?.required ?? (required ? "Choose an option." : undefined) } });
  const [query, setQuery] = useState("");
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const filtered = useFilteredOptions(options, query, format, match, maxResults);
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <Combobox.Root
        items={options}
        filteredItems={filtered}
        value={(field.value as Value | null) ?? null}
        onValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else field.onChange(value);
        }}
        onInputValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else setQuery(value);
        }}
        itemToStringLabel={format}
        isItemEqualToValue={(left, right) => getId(left) === getId(right)}
        disabled={disabled}
      >
        <div className="paper-combobox__input-group">
          <Combobox.Input
            id={id}
            onBlur={field.onBlur}
            aria-required={required || undefined}
            placeholder={placeholder}
            className="paper-input"
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
              field.ref(node);
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger className="paper-combobox__trigger" aria-label={`Show ${label.toLocaleLowerCase()} options`}>
            <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
          </Combobox.Trigger>
        </div>
        <ComboboxPopup options={filtered} getId={getId} format={format} portalContainer={portalContainer} />
      </Combobox.Root>
    </FieldFrame>
  );
}

export interface AutocompleteMultiInputProps<Value> extends AutocompleteContract<Value> {
  readonly maxSelections?: number;
}

export function AutocompleteMultiInput<Value>({
  name,
  label,
  hint,
  required,
  rules,
  options,
  format,
  getId,
  match,
  maxResults,
  maxSelections,
  disabled,
  placeholder,
}: AutocompleteMultiInputProps<Value>) {
  const id = `paper-multiautocomplete-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const { field } = useController({ name, control: form.control, defaultValue: [], rules: { ...rules, required: rules?.required ?? (required ? "Choose at least one option." : undefined) } });
  const values = (field.value ?? []) as Value[];
  const [query, setQuery] = useState("");
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const filtered = useFilteredOptions(options, query, format, match, maxResults);
  const atLimit = maxSelections !== undefined && values.length >= maxSelections;
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <Combobox.Root
        multiple
        items={options}
        filteredItems={filtered}
        value={values}
        onValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else if (maxSelections === undefined || value.length <= maxSelections || value.length < values.length) field.onChange(value);
        }}
        onInputValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else setQuery(value);
        }}
        itemToStringLabel={format}
        isItemEqualToValue={(left, right) => getId(left) === getId(right)}
        disabled={disabled}
      >
        <Combobox.Chips className="paper-combobox__chips">
          {values.map((value) => (
            <Combobox.Chip key={getId(value)} className="paper-combobox__chip">
              {format(value)}
              <Combobox.ChipRemove aria-label={`Remove ${format(value)}`}>
                <XMarkIcon className={iconClassName("compact")} aria-hidden="true" />
              </Combobox.ChipRemove>
            </Combobox.Chip>
          ))}
          <Combobox.Input
            id={id}
            onBlur={field.onBlur}
            aria-required={required || undefined}
            placeholder={atLimit ? "Maximum selections reached" : placeholder}
            className="paper-combobox__chip-input"
            disabled={disabled || atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
              field.ref(node);
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger disabled={disabled || atLimit} className="paper-combobox__trigger" aria-label={`Show ${label.toLocaleLowerCase()} options`}>
            <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
          </Combobox.Trigger>
        </Combobox.Chips>
        <ComboboxPopup options={filtered} getId={getId} format={format} portalContainer={portalContainer} />
      </Combobox.Root>
    </FieldFrame>
  );
}

export type TagsInputProps = Omit<
  AutocompleteMultiInputProps<string>,
  "format" | "getId"
>;

export function TagsInput({
  name,
  label,
  hint,
  required,
  rules,
  options,
  match,
  maxResults = 50,
  maxSelections,
  disabled,
  placeholder = "Type to add a tag",
}: TagsInputProps) {
  const id = `paper-tags-${useId().replace(/:/gu, "")}`;
  const { form, error } = useField(name);
  const { field } = useController({ name, control: form.control, defaultValue: [], rules: { ...rules, required: rules?.required ?? (required ? "Add at least one tag." : undefined) } });
  const values = (field.value ?? []) as string[];
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const chipsRef = useRef<HTMLDivElement>(null);
  const composing = useRef(false);
  const atLimit = maxSelections !== undefined && values.length >= maxSelections;
  const tagKey = (value: string) => value.trim().toLocaleLowerCase();
  const candidate = query.trim();
  const knownTags = [...new Map(options.filter(value => value.trim()).map(value => [tagKey(value), value.trim()])).values()];
  const alreadySelected = values.some(value => tagKey(value) === tagKey(candidate));
  const existingTag = knownTags.find(value => tagKey(value) === tagKey(candidate));
  const canCreate = Boolean(candidate) && !alreadySelected && !existingTag;
  const suggestions = knownTags.filter(value =>
    !values.some(selected => tagKey(selected) === tagKey(value)) &&
    (!candidate || (match ? match(value, query) : tagKey(value).includes(tagKey(candidate))))
  ).slice(0, maxResults);
  const items = canCreate ? [...suggestions, candidate] : suggestions;

  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error?.message?.toString()}>
      <Combobox.Root
        multiple
        items={items}
        filteredItems={items}
        value={values}
        inputValue={query}
        open={open && !atLimit}
        onOpenChange={setOpen}
        onValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else if (maxSelections === undefined || value.length <= maxSelections || value.length < values.length) {
            field.onChange(value);
            if (value.length > values.length) {
              setQuery("");
              setOpen(false);
            }
          }
        }}
        onInputValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else setQuery(value);
        }}
        disabled={disabled}
      >
        <Combobox.Chips ref={chipsRef} className="paper-combobox__chips paper-tags-input">
          {values.map(value => (
            <Combobox.Chip key={value} className="paper-combobox__chip">
              {value}
              <Combobox.ChipRemove aria-label={`Remove ${value}`}>
                <XMarkIcon className={iconClassName("compact")} aria-hidden="true" />
              </Combobox.ChipRemove>
            </Combobox.Chip>
          ))}
          <Combobox.Input
            id={id}
            onBlur={field.onBlur}
            onCompositionStart={() => { composing.current = true; }}
            onCompositionEnd={() => { composing.current = false; }}
            onKeyDown={event => {
              if (event.key !== "Enter") return;
              event.preventDefault();
              if (composing.current || event.nativeEvent.isComposing || event.keyCode === 229 || atLimit) {
                event.preventBaseUIHandler();
                return;
              }
              // Base UI owns highlighted-option selection; plain Enter adds the typed tag.
              if (event.currentTarget.getAttribute("aria-activedescendant")) return;
              event.preventBaseUIHandler();
              if (candidate && !alreadySelected) {
                field.onChange([...values, existingTag ?? candidate]);
                setQuery("");
                setOpen(false);
              }
            }}
            aria-required={required || undefined}
            placeholder={atLimit ? "Maximum tags reached" : placeholder}
            className="paper-combobox__chip-input"
            disabled={disabled}
            readOnly={atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={node => {
              field.ref(node);
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger disabled={disabled || atLimit} className="paper-combobox__trigger" aria-label={`Show ${label.toLocaleLowerCase()} options`}>
            <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
          </Combobox.Trigger>
        </Combobox.Chips>
        <Combobox.Portal container={portalContainer}>
          <Combobox.Positioner anchor={chipsRef} align="start" className="paper-combobox__positioner" sideOffset={4}>
            <Combobox.Popup className="paper-combobox__popup paper-elevation-floating">
              <Combobox.Empty className="paper-combobox__empty">{alreadySelected ? "This tag is already added." : "Type to add a tag."}</Combobox.Empty>
              <Combobox.List>
                {items.map((value, index) => (
                  <Combobox.Item key={value} value={value} index={index} className="paper-combobox__item paper-tags-input__option">
                    {canCreate && value === candidate ? `Add “${value}”` : value}
                  </Combobox.Item>
                ))}
              </Combobox.List>
            </Combobox.Popup>
          </Combobox.Positioner>
        </Combobox.Portal>
      </Combobox.Root>
    </FieldFrame>
  );
}
