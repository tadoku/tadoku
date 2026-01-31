import {
  Combobox,
  ComboboxInput,
  ComboboxButton,
  ComboboxOptions,
  ComboboxOption,
} from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import React, { useState } from 'react'
import { useController, useFormContext } from 'react-hook-form'

export function AutocompleteMultiInput<T>(props: {
  label: string
  name: string
  hint?: string
  options: T[]
  match: (option: T, query: string) => boolean
  format: (option: T) => string
  getIdForOption: (option: T) => string | number
}) {
  const [query, setQuery] = useState('')

  const { name, label, hint, options, match, format, getIdForOption } = props
  const { control } = useFormContext()
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({ defaultValue: [], name: props.name, control })

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  const filtered =
    query === '' ? options : options.filter(option => match(option, query))

  // Needs to be suffixed with -search so 1password doesn't try to autocomplete...
  const id = `${name}-search`

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={id}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <Combobox
        value={value || []}
        onChange={onChange}
        onClose={() => setQuery('')}
        multiple
        by={(a: T, b: T): boolean => getIdForOption(a) === getIdForOption(b)}
      >
        <div className="input relative">
          <div className="z-0">
            <ComboboxInput
              id={id}
              onChange={event => setQuery(event.target.value)}
              displayValue={(selected: T[]) =>
                selected?.map(option => format(option)).join(', ') ?? ''
              }
              className="!pr-7"
            />
            <ComboboxButton className="absolute inset-y-0 right-0 flex items-center pr-2">
              <ChevronUpDownIcon
                className="h-5 w-5 text-gray-400"
                aria-hidden="true"
              />
            </ComboboxButton>
          </div>
          <ComboboxOptions
            transition
            className={`absolute mt-2 z-50 max-h-60 w-full overflow-auto bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none transition ease-in duration-100 data-[closed]:opacity-0`}
          >
            {filtered.length === 0 && query !== '' ? (
              <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                No matches
              </div>
            ) : (
              filtered.map(option => (
                <ComboboxOption
                  key={getIdForOption(option)}
                  value={option}
                  className="relative cursor-default select-none py-2 pl-10 pr-4 data-[focus]:bg-secondary data-[focus]:text-white"
                >
                  {({ selected }) => (
                    <>
                      <span
                        className={`block truncate ${
                          selected ? 'font-medium' : 'font-normal'
                        }`}
                      >
                        {format(option)}
                      </span>
                      {selected ? (
                        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-secondary data-[focus]:text-white">
                          <CheckIcon className="h-5 w-5" aria-hidden="true" />
                        </span>
                      ) : null}
                    </>
                  )}
                </ComboboxOption>
              ))
            )}
          </ComboboxOptions>
        </div>
      </Combobox>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
