import { Combobox } from "@base-ui/react/combobox";
import { useId, useMemo, useRef, useState } from "react";
import { CheckIcon, ChevronDownIcon, XMarkIcon, iconClassName } from "../../icons";
import { FieldFrame, describedBy, useField, type FieldContract } from "./field";

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
        <div className="paper-combobox__input-group">
          <Combobox.Input
            id={id}
            name={field.name}
            onBlur={field.handleBlur}
            aria-required={required || undefined}
            placeholder={placeholder}
            className="paper-input"
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
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
            name={field.name}
            onBlur={field.handleBlur}
            aria-required={required || undefined}
            placeholder={atLimit ? "Maximum selections reached" : placeholder}
            className="paper-combobox__chip-input"
            disabled={disabled || atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={(node) => {
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
  label,
  hint,
  required,
  options,
  match,
  maxResults = 50,
  maxSelections,
  disabled,
  placeholder = "Type to add a tag",
}: TagsInputProps) {
  const id = `paper-tags-${useId().replace(/:/gu, "")}`;
  const { field, error } = useField<string[]>();
  const values = (field.state.value ?? []) as string[];
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
            className="paper-combobox__chip-input"
            disabled={disabled}
            readOnly={atLimit}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy(id, hint, error)}
            ref={node => {
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
