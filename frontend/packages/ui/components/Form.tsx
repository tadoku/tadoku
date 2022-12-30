import { Combobox, Transition } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import React, { Fragment, HTMLProps, useState } from 'react'
import {
  FieldPath,
  FieldValues,
  FormState,
  RegisterOptions,
  useController,
  UseControllerProps,
  UseFormRegister,
} from 'react-hook-form'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  register: UseFormRegister<T>
  formState: FormState<T>
  options?: RegisterOptions
}

interface InputProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
}

export function Input<T extends FieldValues>(props: InputProps<T>) {
  const { name, register, formState, label, type, options, ...inputProps } =
    props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <input
        type={type}
        id={name}
        {...inputProps}
        className={`input ${inputProps?.className ?? ''}`}
        {...register(name, options)}
      />
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface TextaAreaProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLTextAreaElement>, 'name'> {
  label: string
}

export function TextArea<T extends FieldValues>(props: TextaAreaProps<T>) {
  const { name, register, formState, label, options, ...inputProps } = props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <textarea
        id={name}
        {...inputProps}
        className={`input ${inputProps?.className ?? ''}`}
        {...register(name, options)}
      />
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface CheckboxProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type'> {
  label: string
}

export function Checkbox<T extends FieldValues>(props: CheckboxProps<T>) {
  const { name, register, formState, label, options, ...inputProps } = props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <label className="label-inline" htmlFor={name}>
        <input
          type="checkbox"
          id={name}
          {...inputProps}
          className={`input ${inputProps?.className ?? ''}`}
          {...register(name, options)}
        />
        <span>{label}</span>
      </label>
      <span className="error">{errorMessage}</span>
    </div>
  )
}

interface Option {
  value: string
  label: string
}

interface SelectProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLSelectElement>, 'name'> {
  label: string
  values: Option[]
}

export function Select<T extends FieldValues>(props: SelectProps<T>) {
  const {
    name,
    register,
    formState,
    label,
    type,
    options,
    values,
    ...selectProps
  } = props
  const hasError = formState.errors[name] !== undefined
  const errorMessage =
    formState.errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <select
        {...selectProps}
        className={`input ${selectProps?.className ?? ''}`}
        {...register(name, options)}
      >
        {values.map(({ value, label }) => (
          <option value={value} key={value}>
            {label}
          </option>
        ))}
      </select>
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface RadioSelectProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
  values: Option[]
}

export function RadioSelect<T extends FieldValues>(props: RadioSelectProps<T>) {
  const {
    name,
    register,
    formState,
    label,
    type,
    options,
    values,
    ...selectProps
  } = props
  const hasError = formState.errors[name] !== undefined
  let errorMessage =
    formState.errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <span className="label-text">{label}</span>
      {values.map(({ value, label }) => (
        <label
          className="label-inline"
          htmlFor={`${name}-${value}`}
          key={value}
        >
          <input
            type="radio"
            value={value}
            id={`${name}-${value}`}
            {...selectProps}
            {...register(name, options)}
          />
          <span>{label}</span>
        </label>
      ))}
      <span className="error">{errorMessage}</span>
    </div>
  )
}

export function AutocompleteInput<T>(
  props: {
    label: string
    name: string
    options: T[]
    match: (option: T, query: string) => boolean
    format: (option: T) => string
  } & UseControllerProps,
) {
  const [query, setQuery] = useState('')

  const { name, label, options, match, format } = props
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController(props)

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  const filtered =
    query === '' ? options : options.filter(option => match(option, query))

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <Combobox value={value || null} onChange={onChange}>
        <div className="relative z-0 input">
          <Combobox.Input onChange={event => setQuery(event.target.value)} />
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
                  key={format(option)}
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
      </Combobox>
      <span className="error">{errorMessage}</span>
    </label>
  )
}

export function AutocompleteMultiInput<T>(
  props: {
    label: string
    name: string
    hint?: string
    options: T[]
    match: (option: T, query: string) => boolean
    format: (option: T) => string
    getIdForOption: (option: T) => string | number
  } & UseControllerProps,
) {
  const [query, setQuery] = useState('')

  const { name, label, hint, options, match, format, getIdForOption } = props
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({ defaultValue: [], ...props })

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  const filtered =
    query === '' ? options : options.filter(option => match(option, query))

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">{label}</span>
      <Combobox
        value={value || []}
        onChange={onChange}
        multiple
        by={(a, b): boolean => getIdForOption(a) === getIdForOption(b)}
      >
        <div className="input relative z-0">
          <Combobox.Input
            onChange={event => setQuery(event.target.value)}
            displayValue={selected =>
              selected?.map(option => format(option as T)).join(', ') ?? ''
            }
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
                  key={format(option)}
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
      </Combobox>
      {hint ? <span className="label-hint">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}
