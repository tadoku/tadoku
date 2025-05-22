import {
  Combobox,
  Transition,
  RadioGroup as HeadlessRadioGroup,
} from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import React, { ComponentType, Fragment, HTMLProps, useState } from 'react'
import {
  FieldPath,
  FieldValues,
  RegisterOptions,
  useController,
  useFormContext,
} from 'react-hook-form'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  options?: RegisterOptions
}

interface InputProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
  hint?: string
}

export function Input<T extends FieldValues>(props: InputProps<T>) {
  const { name, label, type, options, hint, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <input
        type={type}
        id={name}
        {...inputProps}
        className={`input ${inputProps?.className ?? ''}`}
        {...register(name, options)}
      />
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface TextaAreaProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLTextAreaElement>, 'name'> {
  label: string
  hint?: string
}

export function TextArea<T extends FieldValues>(props: TextaAreaProps<T>) {
  const { name, label, options, hint, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This input is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
      <textarea
        id={name}
        {...inputProps}
        className={`input ${inputProps?.className ?? ''}`}
        {...register(name, options)}
      />
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface CheckboxProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name' | 'type'> {
  label: string
  hint?: string
}

export function Checkbox<T extends FieldValues>(props: CheckboxProps<T>) {
  const { name, label, options, hint, ...inputProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This input is invalid'

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
      {hint ? <span className="label-hint">{hint}</span> : undefined}
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
  hint?: string
}

export function Select<T extends FieldValues>(props: SelectProps<T>) {
  const { name, label, type, options, values, hint, ...selectProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  const errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <label className={`label ${hasError ? 'error' : ''}`} htmlFor={name}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
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
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}

interface RadioSelectProps<T extends FieldValues>
  extends Props<T>,
    Omit<HTMLProps<HTMLInputElement>, 'name'> {
  label: string
  values: Option[]
  hint?: string
}

export function RadioSelect<T extends FieldValues>(props: RadioSelectProps<T>) {
  const { name, label, type, options, values, hint, ...selectProps } = props

  const {
    register,
    formState: { errors },
  } = useFormContext()

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <div className={`label ${hasError ? 'error' : ''}`}>
      <span className="label-text">
        {label}
        {hint ? (
          <span className="label-hint hidden sm:flex">{hint}</span>
        ) : undefined}
      </span>
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
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </div>
  )
}

export function AutocompleteInput<T>(props: {
  label: string
  name: string
  hint?: string
  options: T[]
  match: (option: T, query: string) => boolean
  format: (option: T) => string
}) {
  const [query, setQuery] = useState('')

  const { name, label, options, match, format, hint } = props
  const { control } = useFormContext()
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({ name, control })

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
      <Combobox value={value || null} onChange={onChange}>
        <div className="input relative">
          <div className="z-0">
            <Combobox.Input
              id={id}
              onChange={event => setQuery(event.target.value)}
              displayValue={selected => {
                if (!selected) {
                  return ''
                }

                return format(selected as T)
              }}
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
        </div>
      </Combobox>
      {hint ? <span className="label-hint sm:hidden">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </label>
  )
}

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

export interface RadioProps {
  name: string
  label: string
  hint?: string
  defaultValue?: any
  options: {
    label: string
    description: string
    IconComponent?: ComponentType<any>
    value: any
    disabled?: boolean
    title?: string | undefined
  }[]
}

export function RadioGroup({ name, label, hint, options }: RadioProps) {
  const { control } = useFormContext()
  const {
    field: { value, onChange },
    formState: { errors },
  } = useController({
    name,
    control,
    rules: { required: true },
  })

  const hasError = errors[name] !== undefined
  let errorMessage =
    errors[name]?.message?.toString() || 'This selection is invalid'

  return (
    <HeadlessRadioGroup
      value={value}
      onChange={onChange}
      className={`label ${hasError ? 'error' : ''}`}
    >
      <HeadlessRadioGroup.Label className="label-text">
        {label}
      </HeadlessRadioGroup.Label>
      <div className="col-span-full grid gap-2 grid-cols-fill-48 w-full">
        {options.map(opt => (
          <HeadlessRadioGroup.Option
            value={opt.value}
            key={opt.value.toString()}
            className="input-frame px-4 py-2 ui-checked:border-primary cursor-pointer select-none"
            disabled={opt.disabled}
            title={opt.title}
          >
            <div
              className={`h-stack  items-center w-full ${
                opt.disabled
                  ? 'pointer-events-none text-secondary/30'
                  : 'text-secondary'
              }`}
            >
              {opt.IconComponent ? (
                <opt.IconComponent className="w-4 h-4 mr-2" />
              ) : null}
              <div className={`font-bold mr-4`}>{opt.label}</div>
              <span className="flex items-center justify-center border border-black/10 rounded-xl w-4 h-4 ml-auto text-transparent ui-checked:bg-primary ui-checked:border-primary ui-checked:text-white">
                <CheckIcon className="w-3 h-3" />
              </span>
            </div>
            <div
              className={`mt-2 text-xs ${
                opt.disabled ? ' text-secondary/30' : 'text-gray-600'
              }`}
            >
              {opt.description}
            </div>
          </HeadlessRadioGroup.Option>
        ))}
      </div>
      {hint ? <span className="label-hint">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </HeadlessRadioGroup>
  )
}

interface AmountWithUnitProps {
  label: string
  name: string
  units: Option[]
}

export const AmountWithUnit = ({ label, name, units }: AmountWithUnitProps) => {
  const {
    register,
    formState: { errors },
  } = useFormContext()

  const fieldError = errors?.[name as keyof typeof errors] as any

  const hasError = fieldError !== undefined
  const errorMessage =
    fieldError?.message?.toString() || 'This input is invalid'

  return (
    <div className="w-full">
      <label
        className="block mb-1 font-medium text-sm text-gray-700"
        htmlFor={`${name}.value`}
      >
        {label}
      </label>
      <div className="flex h-11 overflow-visible input-frame group focus-within:ring-2 focus-within:ring-primary/40 focus-within:border-primary transition ease-in-out">
        <input
          type="number"
          step="any"
          id={`${name}.value`}
          {...register(`${name}.value`, { valueAsNumber: true })}
          className="!border-l-0 !border-t-0 !border-b-0 !border-r border-black/10 focus:!border-black/10 !h-10 !bg-none focus:!ring-0 focus:!outline-none w-full"
          placeholder="Enter amount"
        />
        <select
          {...register(`${name}.unit`)}
          className="w-auto !h-10 px-2 pr-8 border-none focus:!ring-0 focus:!outline-none bg-transparent"
        >
          {units.map(unit => (
            <option key={unit.value} value={unit.value}>
              {unit.label}
            </option>
          ))}
        </select>
      </div>
      {hasError ? (
        <p className="text-sm text-red-600 mt-1">{errorMessage}</p>
      ) : null}
    </div>
  )
}
