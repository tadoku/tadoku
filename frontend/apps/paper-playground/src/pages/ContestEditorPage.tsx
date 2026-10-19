import { useStore } from '@tanstack/react-form'
import { usePaperForm, Button, EmptyState, PageHeader, buttonClassName } from 'paper-ui'

import { Link, useNavigate, useParams } from 'react-router-dom'
import { supportedLanguages, type Activity, type SampleContest } from '../data'
import { contestEnd, contestStart, formatDateTime } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import { ContestNotFound } from './contestParts'
import { contestPath, useContestDate } from './contestRound'
import './contests.css'

interface ContestFields { title: string; description: string; start: string; end: string; registrationDeadline: string; allLanguages: boolean; languages: string[]; reading: boolean; listening: boolean; unlisted: boolean }
export function ContestEditorPage() {
  const { contestId } = useParams()
  const { contests, viewer, userId, saveContest } = usePlayground()
  const navigate = useNavigate()
  const asOf = useContestDate()
  const contest = contests.find(item => item.id === contestId)
  const methods = usePaperForm({ defaultValues: ({ title: contest?.title ?? '', description: contest?.description ?? '', start: contest?.start ?? '2026-10-01', end: contest?.end ?? '2026-10-31', registrationDeadline: contest?.registrationDeadline ?? '2026-10-05', allLanguages: !contest || contest.languages.includes('All languages'), languages: contest?.languages.filter(language => language !== 'All languages') ?? [], reading: contest?.activities.includes('Reading') ?? true, listening: contest?.activities.includes('Listening') ?? true, unlisted: contest?.unlisted ?? false } as ContestFields), onSubmit: ({ value: values }) => {
    const activities: Activity[] = []
    if (values.reading) activities.push('Reading')
    if (values.listening) activities.push('Listening')
    const next: SampleContest = { id: contest?.id ?? `contest-${crypto.randomUUID()}`, title: values.title.trim(), description: values.description.trim(), start: values.start, end: values.end, registrationDeadline: values.registrationDeadline, languages: values.allLanguages ? ['All languages'] : values.languages, activities, unlisted: values.unlisted, scope: contest?.scope ?? 'community', ownerId: userId, moderatorIds: contest ? contest.moderatorIds : [userId] }
    saveContest(next)
    navigate(`${contestPath(next, asOf)}&saved=1`)
  } })
  const allLanguages = useStore(methods.store, state => state.values['allLanguages'])
  const [start, end, registrationDeadline] = useStore(methods.store, state => [state.values['start'],state.values['end'],state.values['registrationDeadline']])
  const canManage = ['organizer', 'admin'].includes(viewer) && (!contestId || contest?.ownerId === userId)
  if (contestId && !contest) return <ContestNotFound />
  if (!canManage) return <EmptyState headingLevel={1} title="Organizer access required" description="Creating contests requires organizer permission. You can manage only the contests you organize." actions={<Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link>} />

  return <>
    <PageHeader backLink={contest ? { href: contestPath(contest, asOf), label: `Back to ${contest.title}` } : { href: '/contests/managed', label: 'Back to contests' }} renderLink={renderLink} title={contest ? 'Manage contest' : 'Create a community contest'} lead="Set a shared window for immersion, with clear dates and eligibility." />
    <methods.AppForm><methods.Form className="app-form contest-editor-form"><div className="app-form__fields">
      <methods.AppField name="title" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => Boolean(value?.trim()) || 'Enter a contest title.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Contest title" required maxLength={100} />}</methods.AppField>
      <methods.AppField name="description" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => Boolean(value?.trim()) || 'Describe what participants will do.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.TextArea label="What is this contest about?" rows={3} required />}</methods.AppField>
      <div className="app-form__row"><methods.AppField name="start" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; } }}>{field => <field.Input label="Begins (UTC)" type="date" required hint={start ? `Your time: ${formatDateTime(contestStart(start))}` : undefined} />}</methods.AppField><methods.AppField name="end" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => value >= methods.getFieldValue('start') || 'The end date must be on or after the start date.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Ends (UTC)" type="date" required hint={end ? `Your time: ${formatDateTime(contestEnd(end))}` : undefined} />}</methods.AppField></div>
      <methods.AppField name="registrationDeadline" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => value <= methods.getFieldValue('end') || 'Registration must close on or before the round ends.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Registration closes (UTC)" type="date" required hint={registrationDeadline ? `Your time: ${formatDateTime(contestEnd(registrationDeadline))}. Must close on or before the round ends.` : 'Must close on or before the round ends.'} />}</methods.AppField>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Languages</legend><methods.AppField name="allLanguages">{field => <field.Checkbox label="Allow all languages" />}</methods.AppField>{!allLanguages ? <methods.AppField name="languages" validators={{ onChange: ({ value }) => { if (!value.length) return "Choose at least one option."; const result = (value => methods.getFieldValue('allLanguages') || Array.isArray(value) && value.length > 0 || 'Choose at least one allowed language.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.AutocompleteMultiInput label="Allowed languages" options={supportedLanguages} format={value => value} getId={value => value} required />}</methods.AppField> : <p className="muted">Participants choose up to three languages when registering.</p>}</fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Activities</legend><methods.AppField name="reading" validators={{ onChange: () => { const result = (() => methods.getFieldValue('reading') || methods.getFieldValue('listening') || 'Choose at least one activity.')(); return result === true ? undefined : result || undefined; } }}>{field => <field.Checkbox label="Reading" />}</methods.AppField><methods.AppField name="listening">{field => <field.Checkbox label="Listening" />}</methods.AppField></fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Discovery</legend><methods.AppField name="unlisted">{field => <field.Checkbox label="Unlisted — accessible with the link" hint="Unlisted contests are accessible to anyone with the link. They appear in Managed by me and in each participant’s joined contests, but not in public discovery." />}</methods.AppField></fieldset>
    </div><div className="app-form__actions"><Button type="submit">{contest ? 'Save changes' : 'Create contest'}</Button><Link className={buttonClassName({ variant: 'outline' })} to={contest ? contestPath(contest, asOf) : '/contests/managed'}>Cancel</Link></div></methods.Form></methods.AppForm>
  </>
}
