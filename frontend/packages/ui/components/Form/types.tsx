import { FieldPath, FieldValues, RegisterOptions } from 'react-hook-form'

export interface FormElementProps<T extends FieldValues> {
  name: FieldPath<T>
  options?: RegisterOptions
}

export interface Option {
  value: string
  label: string
}

export interface OptionGroup {
  label: string
  options: Option[]
}
