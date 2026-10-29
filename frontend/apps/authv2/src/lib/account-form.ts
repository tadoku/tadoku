import { useStore } from '@tanstack/react-form'
import { usePaperForm } from 'paper-ui'
import { useState, type FormEvent } from 'react'
import { problemFor, sendIdentity, type Code, type Operation, type Problem } from './identity'
import { useIdentity } from './session'

type Values = Record<string, string>

/** Values typed into the server-rendered controls before hydration, which the form adopts as its own. */
function typedValues<V extends Values>(id: string, defaults: V): V {
  if (typeof document === 'undefined') return defaults
  const element = document.getElementById(id)
  if (!(element instanceof HTMLFormElement)) return defaults
  const data = new FormData(element)
  const typed = Object.entries(defaults).map(([name, fallback]) => {
    const value = data.get(name)
    return [name, typeof value === 'string' ? value : fallback]
  })
  return Object.fromEntries(typed) as V
}

/** Navigation never settles, so the submit button stays pending until the next page replaces this one. */
export function navigate(url: string) {
  window.location.assign(url)
  return new Promise<never>(() => {})
}

/**
 * A progressively enhanced account form. Without JavaScript it posts to the `/actions` route; once hydrated it adopts
 * any values already typed and sends JSON from the browser to the identity API, mapping rejections onto the field
 * they concern or onto the form.
 */
export function useAccountForm<V extends Values>({
  operation, id, defaultValues, nativeError, reauthenticate, body, onSuccess,
}: {
  readonly operation: Operation
  readonly id: string
  readonly defaultValues: V
  /** An error code returned to this page by a native submission. */
  readonly nativeError?: Code
  /** Where to go, instead of showing the problem, when only a fresh login resolves it. */
  readonly reauthenticate?: string
  readonly body?: (values: V) => Values
  readonly onSuccess: (values: V) => unknown
}) {
  const { api } = useIdentity()
  const [initialValues] = useState(() => typedValues(id, defaultValues))
  const names = Object.keys(defaultValues)
  const form = usePaperForm({
    defaultValues: initialValues,
    validators: {
      onSubmitAsync: async ({ value }) => {
        // Sent from the browser, so login throttling sees the visitor's own address.
        const browser = { baseUrl: api, credentials: 'include' as const, headers: { Accept: 'application/json' } }
        const result = await sendIdentity(operation, browser, body ? body(value) : value)
        if (result.ok) return
        const problem = problemFor(result.code, operation, value)
        if (problem.reauthenticate && reauthenticate) return navigate(reauthenticate)
        return problem.field ? { fields: { [problem.field]: problem.message } } : { form: problem, fields: {} }
      },
    },
    onSubmit: ({ value }) => onSuccess(value),
    onSubmitInvalid: ({ formApi }) => {
      const problem = formApi.state.errorMap.onSubmit as Problem | undefined
      if (problem?.code === 'wrong_password') formApi.resetField('password' as never)
      const name = problem?.code === 'wrong_password' ? 'password'
        : names.find(field => formApi.getFieldMeta(field as never)?.errors.length)
      if (name) document.getElementById(`${id}-${name}`)?.focus()
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const submitted = useStore(form.store, state => state.submissionAttempts > 0)
  const succeeded = useStore(form.store, state => state.isSubmitSuccessful)
  const formError = useStore(form.store, state => state.errorMap.onSubmit) as Problem | undefined
  const problem = formError ?? (!submitted && nativeError ? problemFor(nativeError, operation) : undefined)

  const adoptTypedValues = (event: FormEvent<HTMLFormElement>) => {
    const data = new FormData(event.currentTarget)
    for (const name of names) {
      const value = data.get(name)
      if (typeof value === 'string' && value !== (form.getFieldValue(name as never) as unknown)) {
        form.setFieldValue(name as never, value as never)
      }
    }
  }

  return {
    form,
    pending,
    submitted,
    succeeded,
    problem,
    fieldId: (name: keyof V & string) => `${id}-${name}`,
    formProps: { id, method: 'post', action: `/actions/${operation}`, onSubmitCapture: adoptTypedValues },
  }
}
