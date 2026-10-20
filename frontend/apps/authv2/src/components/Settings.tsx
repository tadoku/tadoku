import { useState } from 'react'
import { useSearch } from '@tanstack/react-router'
import { Button, PageHeader, SectionHeading, Surface, usePaperForm } from 'paper-ui'
import { accountSearch, errorMessage, returnDestination, submitAccount } from '../lib/account'
import { Route } from '../routes/__root'
import { AccountFeedback, AccountUnavailable, CredentialsForm } from './AccountForms'

export function Settings() {
  const { identity } = Route.useRouteContext()
  const search = accountSearch(useSearch({ strict: false }))
  const [profileCode, setProfileCode] = useState(search.error)
  const [passwordCode, setPasswordCode] = useState<string>()
  const [saved, setSaved] = useState(search.saved)
  const account = identity.status === 'ready' ? identity.session.account : null
  const continuation = returnDestination(search.return_to)
  const profile = usePaperForm({
    defaultValues: { email: account?.email || '', display_name: account?.display_name || '' },
    onSubmit: async ({ value, formApi }) => {
      setProfileCode(undefined)
      setSaved(undefined)
      const result = await submitAccount('profile', { ...value, return_to: search.return_to || '' })
      if (result.ok) { setSaved('profile'); return }
      setProfileCode(result.code)
      const field = result.code === 'invalid_display_name' ? 'display_name' : result.code === 'invalid_email' || result.code === 'email_taken' ? 'email' : undefined
      if (field) {
        setProfileCode(undefined)
        formApi.setFieldMeta(field, meta => ({ ...meta, errorMap: { ...meta.errorMap, onSubmit: errorMessage(result.code) } }))
        document.getElementById(`profile-${field}`)?.focus()
      }
    },
  })
  const password = usePaperForm({
    defaultValues: { password: '' },
    onSubmit: async ({ value, formApi }) => {
      setPasswordCode(undefined)
      setSaved(undefined)
      const result = await submitAccount('password', { ...value, return_to: search.return_to || '' })
      if (result.ok) {
        if (continuation) { window.location.assign(result.redirect); return }
        formApi.reset()
        setSaved('password')
        return
      }
      setPasswordCode(result.code)
      if (result.code === 'invalid_password') {
        setPasswordCode(undefined)
        formApi.setFieldMeta('password', meta => ({ ...meta, errorMap: { ...meta.errorMap, onSubmit: errorMessage(result.code) } }))
        document.getElementById('new-password')?.focus()
      }
    },
  })
  if (identity.status === 'unavailable') return <AccountUnavailable />
  if (!account) return <CredentialsForm mode="login" />
  return <>
    <PageHeader title="Settings" lead={`Your account, ${account.display_name}.`} />
    {search.return_to ? <p>{continuation ? 'Set your new password to finish and continue to Tadoku.' : errorMessage('invalid_return_to')}</p> : null}
    <Surface as="section" className="account-status" aria-labelledby="profile-heading">
      <SectionHeading id="profile-heading" level={2} title="Profile" description="Your display name is public. Changing your email requires a login within the last hour." />
      <AccountFeedback code={profileCode} />
      {saved === 'profile' ? <p role="status">Profile saved.</p> : null}
      <profile.AppForm><profile.Form className="account-form" method="post" action="/actions/profile" onSubmitCapture={event => {
        const values = new FormData(event.currentTarget)
        for (const field of ['email', 'display_name'] as const) profile.setFieldValue(field, String(values.get(field) || ''))
      }}>
        <input type="hidden" name="return_to" value={search.return_to || ''} />
        <profile.AppField name="display_name">{field => <field.Input id="profile-display_name" label="Display name" autoComplete="nickname" required hint="2–32 characters, including a visible character." />}</profile.AppField>
        <profile.AppField name="email">{field => <field.Input id="profile-email" label="Email" type="email" autoComplete="email" required maxLength={320} />}</profile.AppField>
        <profile.Subscribe selector={state => state.isSubmitting}>{pending => <Button type="submit" disabled={pending} aria-busy={pending}>{pending ? 'Saving…' : 'Save profile'}</Button>}</profile.Subscribe>
      </profile.Form></profile.AppForm>
    </Surface>
    <Surface as="section" className="account-status" aria-labelledby="password-heading">
      <SectionHeading id="password-heading" level={2} title="Password" description="Log in within the last hour, or follow a recovery link. Saving signs out your other sessions." />
      <AccountFeedback code={passwordCode} />
      {saved === 'password' ? <p role="status">Password saved. Your other sessions have been signed out.</p> : null}
      <password.AppForm><password.Form className="account-form" method="post" action="/actions/password" onSubmitCapture={event => password.setFieldValue('password', String(new FormData(event.currentTarget).get('password') || ''))}>
        <input type="hidden" name="return_to" value={search.return_to || ''} />
        <password.AppField name="password">{field => <field.Input id="new-password" label="New password" type="password" autoComplete="new-password" required hint="At least 8 characters, at most 72 UTF-8 bytes. Avoid words from your email." />}</password.AppField>
        <password.Subscribe selector={state => state.isSubmitting}>{pending => <Button type="submit" disabled={pending} aria-busy={pending}>{pending ? 'Saving…' : continuation ? 'Save password and continue' : 'Save password'}</Button>}</password.Subscribe>
      </password.Form></password.AppForm>
    </Surface>
  </>
}
