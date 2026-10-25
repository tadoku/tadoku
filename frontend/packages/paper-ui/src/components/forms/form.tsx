import { createFormHook } from '@tanstack/react-form'
import type { FormHTMLAttributes } from 'react'
import { fieldContext, formContext, useFormContext } from './field'
import { Input } from './Input'
import { Checkbox, Select, TextArea } from './controls'
import { RadioGroup, RadioSelect } from './choice-controls'
import { ToggleSelect } from './ToggleSelect'
import { AutocompleteInput, AutocompleteMultiInput, TagsInput } from './autocomplete-controls'
import { AmountWithUnit } from './AmountWithUnit'

function Form({ children, ...props }: Omit<FormHTMLAttributes<HTMLFormElement>, "onSubmit" | "noValidate">) {
  const form = useFormContext()
  return <form {...props} noValidate onSubmit={async event => {
    event.preventDefault()
    // TanStack Form starts another submission while one is running, so a repeated submit would repeat the operation.
    if (form.state.isSubmitting) return
    const element = event.currentTarget
    await form.handleSubmit()
    if (!form.state.isValid) {
      element.querySelector<HTMLElement>('input[aria-invalid="true"], select[aria-invalid="true"], textarea[aria-invalid="true"], [aria-invalid="true"] button:not(:disabled)')?.focus()
    }
  }}>{children}</form>
}

export const { useAppForm: usePaperForm, withForm: withPaperForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: { Input, Checkbox, Select, TextArea, RadioGroup, RadioSelect, ToggleSelect, AutocompleteInput, AutocompleteMultiInput, TagsInput, AmountWithUnit },
  formComponents: { Form },
})
