import { Combobox, Transition } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import React, { Fragment, useState } from 'react'
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
        multiple
        by={(a, b): boolean => getIdForOption(a) === getIdForOption(b)}
      >
        <div className="input relative">
          <div className="z-0">
            <Combobox.Input
              id={id}
              onChange={event => setQuery(event.target.value)}
              displayValue={(selected: T[]) =>
                selected?.map(option => format(option)).join(', ') ?? ''
              }
              className="!pr-7"
            />
            <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-2">
              <ChevronUpDownIcon
                className="h-5 w-5 text-gray-400"
                aria-hidden="true"
              />
            </Combobox.Button>
          </div>
          <Transition
            as={Fragment}
            leave="transition ease-in duration-100"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
            afterLeave={() => setQuery('')}
          >
            <Combobox.Options
              className={`absolute mt-2 z-50 max-h-60 w-full overflow-auto bg-white py-1 shadow-md shadow-slate-500/20 ring-1 ring-secondary ring-opacity-5 focus:outline-none`}
            >
              {filtered.length === 0 && query !== '' ? (
                <div className="relative cursor-default select-none py-2 px-4 text-gray-700">
                  No matches
                </div>
              ) : (
                filtered.map(option => (
                  <Combobox.Option
                    key={getIdForOption(option)}
                    value={option}
                    className={({ active }) =>
                      `relative cursor-default select-none py-2 pl-10 pr-4 ${
                        active ? 'bg-secondary text-white' : ''
                      }`
                    }
                  >
                    {({ selected, active }) => (
                      <>
                        <span
                          className={`block truncate ${
                            selected ? 'font-medium' : 'font-normal'
                          }`}
                        >
                          {format(option)}
                        </span>
                        {selected ? (
                          <span
                            className={`absolute inset-y-0 left-0 flex items-center pl-3 ${
                              active ? 'text-white' : 'text-secondary'
                            }`}
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        ) : null}
                      </>
                    )}
                  </Combobox.Option>
                ))
              )}
            </Combobox.Options>
          </Transition>
        </div>
      </Combobox>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
