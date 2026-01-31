import { RadioGroup as HeadlessRadioGroup, Radio, Label } from '@headlessui/react'
import { CheckIcon } from '@heroicons/react/20/solid'
import React, { ComponentType } from 'react'
import { useController, useFormContext } from 'react-hook-form'

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
      <Label className="label-text">
        {label}
      </Label>
      <div className="col-span-full grid gap-2 grid-cols-fill-48 w-full">
        {options.map(opt => (
          <Radio
            value={opt.value}
            key={opt.value.toString()}
            className="input-frame px-4 py-2 data-[checked]:border-primary cursor-pointer select-none"
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
              <span className="flex items-center justify-center border border-black/10 rounded-xl w-4 h-4 ml-auto text-transparent group-data-[checked]:bg-primary group-data-[checked]:border-primary group-data-[checked]:text-white">
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
          </Radio>
        ))}
      </div>
      {hint ? <span className="label-hint">{hint}</span> : undefined}
      <span className="error">{errorMessage}</span>
    </HeadlessRadioGroup>
  )
}
