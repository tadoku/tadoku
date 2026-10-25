import { useForm, useStore } from '@tanstack/react-form'
import { useQueryClient } from '@tanstack/react-query'
import { useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, PageHeader, Surface } from 'paper-ui'
import { useState } from 'react'
import { z } from 'zod'
import { createLog, eligibleRegistrations, logForm, logQueryOptions, unitsFor } from '../lib/logs'
import type { Configuration, Failure, LogFormValue, Registration } from '../lib/logs'
import { useLinks } from '../lib/links'
import { useLoginUrl, viewerQueryOptions } from '../lib/session'

const messages: Record<Failure, string> = {
  'account-unavailable': 'Account service unavailable. Nothing was saved. Your entries are kept here; try again when the service recovers.',
  switched: 'You are now signed in as a different account. Nothing was saved. This draft is kept for the account that started it.',
  invalid: 'The activity was not saved. Check the amount, unit and contests, then try again.',
  'signed-out': 'Your session ended. Log in in a new tab, return here and save again. Your entries are kept here.',
  forbidden: 'Your account cannot save activity. Your entries are kept here.',
  deleting: 'Your account is being deleted. Activity cannot be saved.',
  unavailable: 'We could not confirm whether your activity was saved. Check your activity before saving again to avoid a duplicate.',
  'not-found': 'The activity could not be saved. Review the contests and try again.',
}

type Draft = { languageCode: string; activityId: number; quantity: number; unitId: string; tagsText: string; description: string; registrationIds: string[] }

export function LogForm({ configuration, registrations, viewerId }: { configuration: Configuration; registrations: Registration[]; viewerId: string }) {
  const hydrated = useHydrated()
  const loginUrl = useLoginUrl()
  const links = useLinks()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [failure, setFailure] = useState<Failure | null>(null)
  const languageCode = configuration.user_language_codes?.find(code => configuration.languages.some(language => language.code === code)) ?? configuration.languages[0].code
  const activityId = configuration.activities[0].id
  const toValue = (draft: Draft): LogFormValue => ({
    languageCode: draft.languageCode, activityId: draft.activityId,
    measure: configuration.activities.find(activity => activity.id === draft.activityId)?.input_type === 'time_primary'
      ? { kind: 'time', minutes: draft.quantity }
      : { kind: 'amount', amount: draft.quantity, unitId: draft.unitId },
    tags: draft.tagsText.split(','), description: draft.description,
    registrationIds: eligibleRegistrations(registrations, draft.languageCode, draft.activityId).filter(registration => draft.registrationIds.includes(registration.id)).map(registration => registration.id),
  })
  const draftSchema = z.object({
    languageCode: z.string(), activityId: z.number(), quantity: z.number(), unitId: z.string(),
    tagsText: z.string(), description: z.string(), registrationIds: z.array(z.string()),
  }).superRefine((draft, context) => {
    const result = logForm.safeParse(toValue(draft))
    if (!result.success) for (const issue of result.error.issues) {
      const field = issue.path[0] === 'measure' ? issue.path[1] === 'unitId' ? 'unitId' : 'quantity'
        : issue.path[0] === 'tags' ? 'tagsText' : issue.path[0]
      context.addIssue({ code: 'custom', path: [String(field)], message: issue.message })
    }
  })
  const form = useForm({
    defaultValues: {
      languageCode, activityId, quantity: Number.NaN, unitId: unitsFor(configuration, activityId, languageCode)[0]?.id ?? '',
      tagsText: '', description: '', registrationIds: eligibleRegistrations(registrations, languageCode, activityId).map(registration => registration.id),
    } satisfies Draft,
    validators: { onChange: draftSchema, onSubmit: draftSchema },
    onSubmitInvalid: () => { requestAnimationFrame(() => document.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus()) },
    onSubmit: async ({ value }) => {
      setFailure(null)
      try {
        const result = await createLog({ data: { ...logForm.parse(toValue(value)), viewerId } })
        if (!result.ok) {
          setFailure(result.reason)
          if (result.reason === 'switched') await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
          return
        }
        queryClient.setQueryData(logQueryOptions(result.data.id, viewerId).queryKey, result)
        await navigate({ to: '/logs/$logId', params: { logId: result.data.id } })
        document.querySelector<HTMLElement>('h1')?.focus()
      } catch {
        setFailure('unavailable')
      }
    },
  })
  const values = useStore(form.store, state => state.values)
  const pending = useStore(form.store, state => state.isSubmitting)
  const submitted = useStore(form.store, state => state.submissionAttempts > 0)
  const time = configuration.activities.find(activity => activity.id === values.activityId)?.input_type === 'time_primary'
  const units = unitsFor(configuration, values.activityId, values.languageCode)
  const eligible = eligibleRegistrations(registrations, values.languageCode, values.activityId)
  const selectionChanged = (language: string, activity: number) => {
    form.setFieldValue('unitId', unitsFor(configuration, activity, language)[0]?.id ?? '')
    form.setFieldValue('registrationIds', eligibleRegistrations(registrations, language, activity).map(registration => registration.id))
  }

  return <div className="goal-page">
    <PageHeader title="Log activity" lead="Keep a record of what you read, listen to or practise." />
    <Surface className="log-surface">
      <form className="goal-form" noValidate aria-busy={pending} onSubmit={event => { event.preventDefault(); if (!form.state.isSubmitting) void form.handleSubmit() }}>
        <fieldset className="goal-form__controls" disabled={!hydrated || pending}>
          <form.Field name="languageCode">{field => <div className="paper-field">
            <label className="paper-field__label" htmlFor="log-language">Language</label>
            <select className="paper-input" id="log-language" value={field.state.value} onBlur={field.handleBlur} onChange={event => { field.handleChange(event.target.value); selectionChanged(event.target.value, values.activityId) }}>
              {configuration.user_language_codes?.length ? <optgroup label="Your languages">{configuration.languages.filter(language => configuration.user_language_codes?.includes(language.code)).map(language => <option key={language.code} value={language.code}>{language.name}</option>)}</optgroup> : null}
              <optgroup label="All languages">{configuration.languages.filter(language => !configuration.user_language_codes?.includes(language.code)).map(language => <option key={language.code} value={language.code}>{language.name}</option>)}</optgroup>
            </select>
          </div>}</form.Field>
          <form.Field name="activityId">{field => <div className="paper-field">
            <label className="paper-field__label" htmlFor="log-activity">Activity</label>
            <select className="paper-input" id="log-activity" value={field.state.value} onBlur={field.handleBlur} onChange={event => { const id = Number(event.target.value); field.handleChange(id); selectionChanged(values.languageCode, id); form.setFieldValue('quantity', Number.NaN) }}>
              {configuration.activities.map(activity => <option key={activity.id} value={activity.id}>{activity.name}</option>)}
            </select>
          </div>}</form.Field>
          <div className="goal-form__row">
            <form.Field name="quantity">{field => <Field id="log-quantity" label={time ? 'Minutes' : 'Amount'} error={submitted ? errorText(field.state.meta.errors) : undefined}>
              <input className="paper-input" id="log-quantity" type="number" inputMode="decimal" step="any" min="0" value={Number.isNaN(field.state.value) ? '' : field.state.value} aria-invalid={submitted && field.state.meta.errors.length > 0} aria-describedby={submitted && field.state.meta.errors.length ? 'log-quantity-error' : undefined} onBlur={field.handleBlur} onChange={event => field.handleChange(event.target.valueAsNumber)} />
            </Field>}</form.Field>
            {!time ? <form.Field name="unitId">{field => <Field id="log-unit" label="Unit" error={submitted ? errorText(field.state.meta.errors) : undefined}>
              <select className="paper-input" id="log-unit" value={field.state.value} aria-invalid={submitted && field.state.meta.errors.length > 0} onBlur={field.handleBlur} onChange={event => field.handleChange(event.target.value)}>
                {!units.length ? <option value="">No units available</option> : null}
                {units.map(unit => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
              </select>
            </Field>}</form.Field> : null}
          </div>
          <form.Field name="description">{field => <Field id="log-description" label="Description" error={submitted ? errorText(field.state.meta.errors) : undefined}>
            <textarea className="paper-input" id="log-description" rows={3} value={field.state.value} aria-invalid={submitted && field.state.meta.errors.length > 0} aria-describedby={submitted && field.state.meta.errors.length ? 'log-description-error' : undefined} onBlur={field.handleBlur} onChange={event => field.handleChange(event.target.value)} />
          </Field>}</form.Field>
          <form.Field name="tagsText">{field => <Field id="log-tags" label="Tags" error={submitted ? errorText(field.state.meta.errors) : undefined}>
            <input className="paper-input" id="log-tags" value={field.state.value} aria-invalid={submitted && field.state.meta.errors.length > 0} aria-describedby={submitted && field.state.meta.errors.length ? 'log-tags-hint log-tags-error' : 'log-tags-hint'} onBlur={field.handleBlur} onChange={event => field.handleChange(event.target.value)} />
            <p className="paper-field__hint" id="log-tags-hint">Separate tags with commas. Up to 10 tags.</p>
          </Field>}</form.Field>
          <form.Field name="registrationIds">{field => <fieldset className="paper-choice-field">
            <legend className="paper-field__label">Contests</legend>
            {eligible.length ? eligible.map(registration => <label className="log-contest" key={registration.id}>
              <input type="checkbox" checked={field.state.value.includes(registration.id)} onChange={event => field.handleChange(event.target.checked ? [...field.state.value, registration.id] : field.state.value.filter(id => id !== registration.id))} /> {registration.contest.title}
            </label>) : <p>Saved to your personal record only.</p>}
          </fieldset>}</form.Field>
        </fieldset>
        {failure ? <div role="alert"><p>{messages[failure]}</p>{failure === 'switched' ? <a className="text-link" href="/logs/new">Start a new activity for this account</a> : failure === 'signed-out' ? <a className="text-link" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a> : failure === 'unavailable' ? <a className="text-link" href={links.legacyUrl(`/profile/${viewerId}/updates`)} target="_blank" rel="noopener noreferrer">Check your activity</a> : null}</div> : null}
        <Button type="submit" disabled={!hydrated || pending}>{pending ? 'Saving…' : 'Save activity'}</Button>
      </form>
    </Surface>
  </div>
}

function errorText(errors: unknown[]) {
  return errors.map(error => typeof error === 'string' ? error : error && typeof error === 'object' && 'message' in error ? String(error.message) : '').filter(Boolean).join(' ')
}

function Field({ id, label, error, children }: { id: string; label: string; error?: string; children: import('react').ReactNode }) {
  return <div className="paper-field"><label className="paper-field__label" htmlFor={id}>{label}</label>{children}{error ? <p id={`${id}-error`} className="paper-field__error" role="alert">{error}</p> : null}</div>
}
