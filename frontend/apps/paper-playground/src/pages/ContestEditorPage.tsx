import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { AutocompleteMultiInput, Button, Checkbox, EmptyState, Input, PageHeader, TextArea, buttonClassName } from 'paper-ui'
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
  const methods = useForm<ContestFields>({ defaultValues: { title: contest?.title ?? '', description: contest?.description ?? '', start: contest?.start ?? '2026-10-01', end: contest?.end ?? '2026-10-31', registrationDeadline: contest?.registrationDeadline ?? '2026-10-05', allLanguages: !contest || contest.languages.includes('All languages'), languages: contest?.languages.filter(language => language !== 'All languages') ?? [], reading: contest?.activities.includes('Reading') ?? true, listening: contest?.activities.includes('Listening') ?? true, unlisted: contest?.unlisted ?? false } })
  const allLanguages = useWatch({ control: methods.control, name: 'allLanguages' })
  const [start, end, registrationDeadline] = useWatch({ control: methods.control, name: ['start', 'end', 'registrationDeadline'] })
  const canManage = ['organizer', 'admin'].includes(viewer) && (!contestId || contest?.ownerId === userId)
  if (contestId && !contest) return <ContestNotFound />
  if (!canManage) return <EmptyState headingLevel={1} title="Organizer access required" description="Creating contests requires organizer permission. You can manage only the contests you organize." actions={<Link className={buttonClassName({ variant: 'outline' })} to="/contests/official">Browse contests</Link>} />
  const save = methods.handleSubmit(values => {
    const activities: Activity[] = []
    if (values.reading) activities.push('Reading')
    if (values.listening) activities.push('Listening')
    const next: SampleContest = { id: contest?.id ?? `contest-${crypto.randomUUID()}`, title: values.title.trim(), description: values.description.trim(), start: values.start, end: values.end, registrationDeadline: values.registrationDeadline, languages: values.allLanguages ? ['All languages'] : values.languages, activities, unlisted: values.unlisted, scope: contest?.scope ?? 'community', ownerId: userId, moderatorIds: contest ? contest.moderatorIds : [userId] }
    saveContest(next)
    navigate(`${contestPath(next, asOf)}&saved=1`)
  })
  return <>
    <PageHeader backLink={contest ? { href: contestPath(contest, asOf), label: `Back to ${contest.title}` } : { href: '/contests/managed', label: 'Back to contests' }} renderLink={renderLink} title={contest ? 'Manage contest' : 'Create a community contest'} lead="Set a shared window for immersion, with clear dates and eligibility." />
    <FormProvider {...methods}><form onSubmit={save} noValidate className="app-form contest-editor-form"><div className="app-form__fields">
      <Input name="title" label="Contest title" required maxLength={100} rules={{ validate: value => Boolean(value?.trim()) || 'Enter a contest title.' }} />
      <TextArea name="description" label="What is this contest about?" rows={3} required rules={{ validate: value => Boolean(value?.trim()) || 'Describe what participants will do.' }} />
      <div className="app-form__row"><Input name="start" label="Begins (UTC)" type="date" required hint={start ? `Your time: ${formatDateTime(contestStart(start))}` : undefined} /><Input name="end" label="Ends (UTC)" type="date" required hint={end ? `Your time: ${formatDateTime(contestEnd(end))}` : undefined} rules={{ validate: value => value >= methods.getValues('start') || 'The end date must be on or after the start date.' }} /></div>
      <Input name="registrationDeadline" label="Registration closes (UTC)" type="date" required hint={registrationDeadline ? `Your time: ${formatDateTime(contestEnd(registrationDeadline))}. Must close on or before the round ends.` : 'Must close on or before the round ends.'} rules={{ validate: value => value <= methods.getValues('end') || 'Registration must close on or before the round ends.' }} />
      <fieldset className="contest-form-group"><legend className="paper-type-component">Languages</legend><Checkbox name="allLanguages" label="Allow all languages" />{!allLanguages ? <AutocompleteMultiInput name="languages" label="Allowed languages" options={supportedLanguages} format={value => value} getId={value => value} required rules={{ validate: value => methods.getValues('allLanguages') || Array.isArray(value) && value.length > 0 || 'Choose at least one allowed language.' }} /> : <p className="muted">Participants choose up to three languages when registering.</p>}</fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Activities</legend><Checkbox name="reading" label="Reading" rules={{ validate: () => methods.getValues('reading') || methods.getValues('listening') || 'Choose at least one activity.' }} /><Checkbox name="listening" label="Listening" /></fieldset>
      <fieldset className="contest-form-group"><legend className="paper-type-component">Discovery</legend><Checkbox name="unlisted" label="Unlisted — accessible with the link" hint="Unlisted contests are accessible to anyone with the link. They appear in Managed by me and in each participant’s joined contests, but not in public discovery." /></fieldset>
    </div><div className="app-form__actions"><Button type="submit">{contest ? 'Save changes' : 'Create contest'}</Button><Link className={buttonClassName({ variant: 'outline' })} to={contest ? contestPath(contest, asOf) : '/contests/managed'}>Cancel</Link></div></form></FormProvider>
  </>
}
