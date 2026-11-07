import * as stylex from '@stylexjs/stylex'
import { createFormHook } from '@tanstack/react-form'
import { sx, type StyleXProp } from '../../foundations/props'
import type { FieldsetHTMLAttributes, FormHTMLAttributes, ReactNode } from 'react'
import { border, color, space } from '../../tokens/tokens.stylex'
import { fieldContext, formContext, useFormContext } from './field'
import { Input } from './Input'
import { Checkbox, Select, TextArea } from './controls'
import { RadioGroup, RadioSelect } from './choice-controls'
import { ToggleSelect } from './ToggleSelect'
import { AutocompleteInput, AutocompleteMultiInput, TagsInput } from './autocomplete-controls'
import { AmountWithUnit } from './AmountWithUnit'

function Form({ children, className, xstyle, ...props }: Omit<FormHTMLAttributes<HTMLFormElement>, "onSubmit" | "noValidate"> & { readonly xstyle?: StyleXProp }) {
  const form = useFormContext()
  return <form {...props} {...sx(className, xstyle)} noValidate onSubmit={async event => {
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

const footerStyles = stylex.create({
  footer: {    alignItems: 'center',
    borderBlockStartColor: color.ruleDefault,
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
 boxSizing: "border-box",
    columnGap: space[24],
    display: 'flex',
    flexWrap: 'wrap-reverse',
    justifyContent: 'space-between',
    marginBlockStart: space[24],
    paddingBlockStart: space[24],
    rowGap: space[12],
  },
  actions: { alignItems: 'center', columnGap: space[20], display: 'flex', flexWrap: 'wrap', rowGap: space[12] },
})

/** A form's closing row: actions at the start, an optional summary such as an estimate at the end. */
export function FormFooter({ children, summary, xstyle }: { readonly children: ReactNode; readonly summary?: ReactNode; readonly xstyle?: StyleXProp }) {
  return (
    <div {...stylex.props(footerStyles.footer, xstyle)}>
      <div {...stylex.props(footerStyles.actions)}>{children}</div>
      {summary}
    </div>
  )
}

const fieldSetStyles = stylex.create({
  root: { margin: 0, padding: 0, borderWidth: 0, gap: space[24], display: 'grid', minInlineSize: 0, },
})

/** A form's fields stacked at the form rhythm; `disabled` turns every control in it off at once. */
export function FieldSet({ xstyle, ...props }: FieldsetHTMLAttributes<HTMLFieldSetElement> & { readonly xstyle?: StyleXProp }) {
  return <fieldset {...props} {...stylex.props(fieldSetStyles.root, xstyle)} />
}
