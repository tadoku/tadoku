import { Combobox } from "@base-ui/react/combobox";
import { useId, useMemo, useRef, useState } from "react";
import { CheckIcon, ChevronDownIcon, XMarkIcon, iconClassName } from "../../icons";
import { surface } from "../../foundations/recipes";
import { cn } from "../../foundations/variants";
import { FieldFrame, describedBy, useField, type FieldContract } from "./field";
import { combobox, control } from "./styles";

const slots = combobox();
const popupClassName = cn(surface({ tone: "overlay", elevation: "floating" }), slots.popup());

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
      <Combobox.Positioner className={slots.positioner()} sideOffset={4}>
        <Combobox.Popup className={popupClassName}>
          <Combobox.Empty className={slots.empty()}>No matching options.</Combobox.Empty>
          <Combobox.List>
            {options.map((option, index) => (
              <Combobox.Item
                key={getId(option)}
                value={option}
                index={index}
                className={slots.item()}
              >
                <Combobox.ItemIndicator className={slots.indicator()} aria-hidden="true">
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
  label,
  hint,
  required,
  options,
  format,
  getId,
  match,
  maxResults,
  disabled,
  placeholder,
}: AutocompleteInputProps<Value>) {
  const id = `paper-autocomplete-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<Value | null>();
  const [query, setQuery] = useState("");
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const filtered = useFilteredOptions(options, query, format, match, maxResults);
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
      <Combobox.Root
        items={options}
        filteredItems={filtered}
        value={(field.state.value as Value | null) ?? null}
        onValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else field.handleChange(value);
        }}
        onInputValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else setQuery(value);
        }}
        itemToStringLabel={format}
        isItemEqualToValue={(left, right) => getId(left) === getId(right)}
        disabled={disabled}
      >
        <div className={slots.group()}>
          <Combobox.Input
            id={id}
            name={field.name}
            onBlur={field.handleBlur}
            aria-required={required || undefined}
            placeholder={placeholder}
            className={cn(control({ interactive: !error, invalid: Boolean(error) }), slots.input())}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger className={slots.trigger()} aria-label={`Show ${label.toLocaleLowerCase()} options`}>
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
  label,
  hint,
  required,
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
  const { field, error } = useField<Value[]>();
  const values = (field.state.value ?? []) as Value[];
  const [query, setQuery] = useState("");
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const filtered = useFilteredOptions(options, query, format, match, maxResults);
  const atLimit = maxSelections !== undefined && values.length >= maxSelections;
  return (
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
      <Combobox.Root
        multiple
        items={options}
        filteredItems={filtered}
        value={values}
        onValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else if (maxSelections === undefined || value.length <= maxSelections || value.length < values.length) field.handleChange(value);
        }}
        onInputValueChange={(value, details) => {
          if (details.reason === "escape-key") details.cancel();
          else setQuery(value);
        }}
        itemToStringLabel={format}
        isItemEqualToValue={(left, right) => getId(left) === getId(right)}
        disabled={disabled}
      >
        <Combobox.Chips className={combobox({ invalid: Boolean(error) }).chips()}>
          {values.map((value) => (
            <Combobox.Chip key={getId(value)} className={slots.chip()}>
              {format(value)}
              <Combobox.ChipRemove className={slots.chipRemove()} aria-label={`Remove ${format(value)}`}>
                <XMarkIcon className={iconClassName("compact")} aria-hidden="true" />
              </Combobox.ChipRemove>
            </Combobox.Chip>
          ))}
          <Combobox.Input
            id={id}
            name={field.name}
            onBlur={field.handleBlur}
            aria-required={required || undefined}
            placeholder={atLimit ? "Maximum selections reached" : placeholder}
            className={slots.chipInput()}
            disabled={disabled || atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger disabled={disabled || atLimit} className={cn(slots.trigger(), slots.chipTrigger())} aria-label={`Show ${label.toLocaleLowerCase()} options`}>
            <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
          </Combobox.Trigger>
        </Combobox.Chips>
        <ComboboxPopup options={filtered} getId={getId} format={format} portalContainer={portalContainer} />
      </Combobox.Root>
    </FieldFrame>
  );
}

export interface TagsInputProps extends Omit<AutocompleteMultiInputProps<string>, "format" | "getId"> {
  /** Called with the typed text whenever it changes, including "" after a tag is added, to fetch remote suggestions. */
  readonly onQueryChange?: (query: string) => void;
  /** The visible label of a suggested tag, such as its usage count. The selected value stays the tag. */
  readonly formatOption?: (tag: string) => string;
}

export function TagsInput({
  label,
  hint,
  required,
  options,
  match,
  maxResults = 50,
  maxSelections,
  disabled,
  placeholder = "Type to add a tag",
  onQueryChange,
  formatOption = tag => tag,
}: TagsInputProps) {
  const id = `paper-tags-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string[]>();
  const values = (field.state.value ?? []) as string[];
  const [query, setQueryState] = useState("");
  const setQuery = (next: string) => {
    setQueryState(next);
    if (next !== query) onQueryChange?.(next);
  };
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
    <FieldFrame id={id} label={label} hint={hint} required={required} error={error}>
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
            field.handleChange(value);
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
        <Combobox.Chips ref={chipsRef} className={combobox({ invalid: Boolean(error) }).chips({ className: "paper-tags-input" })}>
          {values.map(value => (
            <Combobox.Chip key={value} className={slots.chip()}>
              {value}
              <Combobox.ChipRemove className={slots.chipRemove()} aria-label={`Remove ${value}`}>
                <XMarkIcon className={iconClassName("compact")} aria-hidden="true" />
              </Combobox.ChipRemove>
            </Combobox.Chip>
          ))}
          <Combobox.Input
            id={id}
            name={field.name}
            onBlur={field.handleBlur}
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
                field.handleChange([...values, existingTag ?? candidate]);
                setQuery("");
                setOpen(false);
              }
            }}
            aria-required={required || undefined}
            placeholder={atLimit ? "Maximum tags reached" : placeholder}
            className={slots.chipInput()}
            disabled={disabled}
            readOnly={atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={node => {
              if (node && node.ownerDocument.body !== portalContainer) setPortalContainer(node.ownerDocument.body);
            }}
          />
          <Combobox.Trigger disabled={disabled || atLimit} className={cn(slots.trigger(), slots.chipTrigger())} aria-label={`Show ${label.toLocaleLowerCase()} options`}>
            <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
          </Combobox.Trigger>
        </Combobox.Chips>
        <Combobox.Portal container={portalContainer}>
          <Combobox.Positioner anchor={chipsRef} align="start" className={slots.positioner()} sideOffset={4}>
            <Combobox.Popup className={popupClassName}>
              <Combobox.Empty className={slots.empty()}>{alreadySelected ? "This tag is already added." : "Type to add a tag."}</Combobox.Empty>
              <Combobox.List>
                {items.map((value, index) => (
                  <Combobox.Item key={value} value={value} index={index} className={cn(slots.item(), slots.option())}>
                    {canCreate && value === candidate ? `Add “${value}”` : formatOption(value)}
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
