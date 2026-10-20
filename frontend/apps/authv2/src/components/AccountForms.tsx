import { useMemo, useState } from 'react'
import { Button, PageHeader, Surface, usePaperForm } from 'paper-ui'
import { useSearch } from '@tanstack/react-router'
import { accountSearch, errorMessage, returnDestination, submitAccount } from '../lib/account'
import { Route } from '../routes/__root'

export function AccountFeedback({ code }: { code?: string }) {
  const { return_to } = accountSearch(useSearch({ strict: false }))
  const reauthenticate = code === 'stale_session' || code === 'unauthorized'
  return code ? <div className="account-feedback" role="alert">
    <p>{errorMessage(code)}</p>
    {reauthenticate ? <a href={`/login${return_to ? `?${new URLSearchParams({ return_to })}` : ''}`}>Log in again</a> : null}
  </div> : null
}

export function AccountUnavailable() {
  const { identity } = Route.useRouteContext()
  return <><PageHeader title="Your Tadoku account" /><Surface className="account-status">
    <p role="status">Account service is unavailable. Try again later.</p>
    <a className="text-link" href={new URL('/login', identity.accountUrl).href}>Open existing account app</a>
  </Surface></>
}

export function CredentialsForm({ mode }: { mode: 'login' | 'registration' }) {
  const { identity } = Route.useRouteContext()
  const search = accountSearch(useSearch({ strict: false }))
  const [code, setCode] = useState(search.error)
  const registration = mode === 'registration'
  const continuation = returnDestination(search.return_to)
  const initialValues = useMemo(() => {
    const element = typeof document === 'undefined' ? null : document.getElementById(`credentials-${mode}-form`)
    const values = typeof HTMLFormElement !== 'undefined' && element instanceof HTMLFormElement ? new FormData(element) : null
    return {
      email: String(values?.get('email') || ''),
      display_name: String(values?.get('display_name') || ''),
      password: String(values?.get('password') || ''),
    }
  }, [mode])
  const form = usePaperForm({
    defaultValues: initialValues,
    onSubmit: async ({ value, formApi }) => {
      setCode(undefined)
      const result = await submitAccount(mode, { ...value, return_to: search.return_to || '' })
      if (result.ok) { window.location.assign(result.redirect); return }
      setCode(result.code)
      const field = result.code === 'invalid_display_name' ? 'display_name'
        : result.code === 'email_taken' || result.code === 'invalid_email' ? 'email'
        : result.code === 'wrong_password' || result.code === 'invalid_password' ? 'password' : undefined
      if (field) {
        setCode(undefined)
        formApi.setFieldMeta(field, meta => ({ ...meta, errorMap: { ...meta.errorMap, onSubmit: errorMessage(result.code) } }))
        document.getElementById(`credentials-${field}`)?.focus()
      }
    },
  })
  if (identity.status === 'unavailable') return <AccountUnavailable />
  return <>
    <PageHeader title={registration ? 'Create an account' : 'Log in'} lead={registration ? 'Join Tadoku and make time for your languages.' : 'Welcome back. Pick up where you left off.'} />
    {identity.session.account ? <p>You are signed in as {identity.session.account.display_name}. <a href="/">Open Settings</a>, or log in again below.</p> : null}
    <Surface className="account-status">
      <AccountFeedback code={search.return_to && !continuation ? 'invalid_return_to' : code} />
      <form.AppForm><form.Form id={`credentials-${mode}-form`} className="account-form" method="post" action={`/actions/${mode}`} onSubmitCapture={event => {
        const values = new FormData(event.currentTarget)
        for (const field of ['email', 'display_name', 'password'] as const) form.setFieldValue(field, String(values.get(field) || ''))
      }}>
        <input type="hidden" name="return_to" value={search.return_to || ''} />
        <form.AppField name="email">{field => <field.Input id="credentials-email" label="Email" type="email" autoComplete="email" required maxLength={320} />}</form.AppField>
        {registration ? <form.AppField name="display_name">{field => <field.Input id="credentials-display_name" label="Display name" autoComplete="nickname" required hint="2–32 characters. This is your public name." />}</form.AppField> : null}
        <form.AppField name="password">{field => <field.Input id="credentials-password" label="Password" type="password" autoComplete={registration ? 'new-password' : 'current-password'} required hint={registration ? 'At least 8 characters, at most 72 UTF-8 bytes. Avoid words from your email.' : undefined} />}</form.AppField>
        <form.Subscribe selector={state => state.isSubmitting}>{pending => <Button type="submit" disabled={pending} aria-busy={pending}>{pending ? 'Please wait…' : registration ? 'Create account' : 'Log in'}</Button>}</form.Subscribe>
      </form.Form></form.AppForm>
      <nav className="account-links" aria-label="Account help">
        <a href={`${registration ? '/login' : '/register'}${search.return_to ? `?${new URLSearchParams({ return_to: search.return_to })}` : ''}`}>{registration ? 'Already have an account? Log in' : 'Create an account'}</a>
        <a href={`/recovery${search.return_to ? `?${new URLSearchParams({ return_to: search.return_to })}` : ''}`}>Forgot your password?</a>
      </nav>
    </Surface>
  </>
}

export function RecoveryForm() {
  const { identity } = Route.useRouteContext()
  const search = accountSearch(useSearch({ strict: false }))
  const [code, setCode] = useState(search.error)
  const [sent, setSent] = useState(search.saved === 'recovery')
  const form = usePaperForm({
    defaultValues: { email: '' },
    onSubmit: async ({ value }) => {
      setCode(undefined)
      const result = await submitAccount('recovery', { ...value, return_to: search.return_to || '' })
      if (result.ok) setSent(true)
      else setCode(result.code)
    },
  })
  if (identity.status === 'unavailable') return <AccountUnavailable />
  return <>
    <PageHeader title="Reset your password" lead="We’ll send a link to help you get back into your account." />
    <Surface className="account-status">
      <AccountFeedback code={code} />
      {sent ? <p role="status">If an account exists for that email, a recovery link is on its way.</p> : null}
      <form.AppForm><form.Form className="account-form" method="post" action="/actions/recovery" onSubmitCapture={event => form.setFieldValue('email', String(new FormData(event.currentTarget).get('email') || ''))}>
        <input type="hidden" name="return_to" value={search.return_to || ''} />
        <form.AppField name="email">{field => <field.Input label="Email" type="email" autoComplete="email" required maxLength={320} />}</form.AppField>
        <form.Subscribe selector={state => state.isSubmitting}>{pending => <Button type="submit" disabled={pending} aria-busy={pending}>{pending ? 'Sending…' : 'Send recovery email'}</Button>}</form.Subscribe>
      </form.Form></form.AppForm>
      <p className="account-note">Recovery links expire after one hour and work once. If a link has expired, request another here.</p>
      <a href="/login">Back to log in</a>
    </Surface>
  </>
}
