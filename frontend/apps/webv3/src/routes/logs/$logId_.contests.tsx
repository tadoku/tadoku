import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { Link, createFileRoute, notFound, useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, buttonClassName, EmptyState, formFooter, PageHeader, Surface, usePaperForm } from 'paper-ui'
import { useState } from 'react'
import { z } from 'zod'
import { DraftOwnerGate } from '../../components/DraftOwnerGate'
import { LogDescription } from '../../components/LogDescription'
import { useLinks } from '../../lib/links'
import { formatTracking, logName, pointsUnit, scoreFormat } from '../../lib/format'
import { contestChoices } from '../../lib/log-draft'
import {
  logQuery,
  ownerRegistrationsQuery,
  useUpdateLogContests,
  type ImmersionLog,
  type LogFailure,
  type Registration,
} from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'
import { recordSurface } from '../../components/standings-styles'
import { formControls } from '../../components/form-styles'

export const Route = createFileRoute('/logs/$logId_/contests')({
  head: () => ({ meta: [{ title: 'Submit to contests · Tadoku' }] }),
  // Webv2's `?preselect=1` selects every eligible contest for a log not yet submitted to any.
  // The router reads `1` as a number; a string would be written back to the URL as `"1"`.
  validateSearch: z.object({ preselect: z.coerce.number().optional().catch(undefined) }),
  loader: async ({ context, params }) => {
    if (!z.guid().safeParse(params.logId).success) throw notFound()
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status !== 'signed-in') return

    const log = await context.queryClient.ensureQueryData(logQuery(params.logId, viewer.userId))
    if (!log.ok && log.reason === 'not-found') throw notFound()
    if (log.ok && (log.data.user_id === viewer.userId || viewer.role === 'admin')) {
      await context.queryClient.ensureQueryData(ownerRegistrationsQuery(viewer.userId, log.data.user_id))
    }
  },
  component: LogContests,
})

const signIn = {
  title: 'Sign in to submit to contests',
  description: 'Log in to choose the contests your activity counts in.',
}

function LogContests() {
  return <DraftOwnerGate signIn={signIn}>{(viewerId, admin) => <ContestsDraft viewerId={viewerId} admin={admin} />}</DraftOwnerGate>
}

/**
 * The log's owner or an administrator changes its submissions, choosing among the owner's registrations; the API decides.
 * The choice starts from the log as first read.
 */
function ContestsDraft({ viewerId, admin }: { viewerId: string; admin: boolean }) {
  const { logId } = Route.useParams()
  const logResult = useQuery({ ...logQuery(logId, viewerId), staleTime: Infinity })
  const log = logResult.data
  const ownerId = log?.ok ? log.data.user_id : viewerId
  const allowed = ownerId === viewerId || admin
  const registrationsResult = useQuery({ ...ownerRegistrationsQuery(viewerId, ownerId), enabled: Boolean(log?.ok) && allowed })
  const loginUrl = useLoginUrl()
  const registrations = registrationsResult.data

  if (logResult.isPending) return null
  if (log?.ok && !allowed) {
    return (
      <EmptyState
        headingLevel={1}
        title="This is someone else's record"
        description="Only the learner who logged this activity can submit it to contests."
        actions={<Link className="text-link hover:text-link-hover" to="/logs/$logId" params={{ logId }}>View the log</Link>}
      />
    )
  }
  if (log?.ok && registrationsResult.isPending) return null
  if (log?.ok && registrations?.ok) {
    return <ContestsForm log={log.data} registrations={registrations.data} viewerId={viewerId} />
  }

  if (log?.ok === false && log.reason === 'not-found') {
    return <EmptyState headingLevel={1} title="Activity not found" description="This log may have been deleted." />
  }
  if ((log?.ok === false && log.reason === 'signed-out') || (registrations?.ok === false && registrations.reason === 'signed-out')) {
    return (
      <EmptyState
        headingLevel={1}
        title="Your session ended"
        description="Log in again to submit this activity to contests."
        actions={<a className={buttonClassName()} href={loginUrl}>Log in</a>}
      />
    )
  }
  return (
    <EmptyState
      headingLevel={1}
      title="Contests unavailable"
      description="We could not load your contests for this activity. Try again."
      actions={<Button onClick={() => void Promise.all([logResult.refetch(), registrationsResult.refetch()])}>Try again</Button>}
    />
  )
}

function ContestsForm({ log, registrations, viewerId }: {
  log: ImmersionLog
  registrations: Registration[]
  viewerId: string
}) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const updateLogContests = useUpdateLogContests()
  const links = useLinks()
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const [today] = useState(() => new Date().toISOString().slice(0, 10))
  const choices = contestChoices(log, registrations, today)
  const eligible = choices.filter(choice => choice.kind === 'eligible').map(choice => choice.registration.id)
  const submitted = log.registrations ?? []
  const preselect = Route.useSearch().preselect === 1 && submitted.length === 0
  const [defaultValues] = useState(() => ({
    registrationIds: preselect
      ? eligible
      : eligible.filter(id => submitted.some(submission => submission.registration_id === id)),
  }))
  const form = usePaperForm({
    defaultValues,
    onSubmit: async ({ value }) => {
      setFailure(null)
      const registrationIds = value.registrationIds.filter(id => eligible.includes(id))
      const result = await updateLogContests({ viewerId, logId: log.id, registrationIds })
      if (result.ok) {
        await navigate({ to: '/logs/$logId', params: { logId: log.id } })
      } else {
        setFailure(result.reason)
      }
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const disabled = !hydrated || pending
  // Submissions to finished contests are not offered: the API keeps them and their saved scores.
  const finished = submitted.filter(submission => submission.contest_end < today)

  return (
    <div>
      <PageHeader
        title="Submit to contests"
        backLink={{ href: `/logs/${log.id}`, label: logName(log) }}
      />
      <Surface className={recordSurface}>
        <LogSummary log={log} />
        <form.AppForm>
          <form.Form className="min-w-0" aria-busy={pending}>
            <fieldset className={formControls} disabled={disabled}>
              <fieldset className="m-0 grid gap-2 border-0 p-0">
                <legend className="sr-only">Contests</legend>
                {/* As in Webv2: open contests first, then unavailable ones, each naming why on its own row. */}
                <div className="grid [&>*:not(:first-child)]:-mt-px">
                {[...choices].sort((a, b) => Number(a.kind === 'ineligible') - Number(b.kind === 'ineligible')).map(choice => (
                  <form.AppField key={choice.registration.id} name="registrationIds">
                    {field => (
                      <field.Checkbox
                        variant="card"
                        value={choice.registration.id}
                        label={choice.registration.contest.title}
                        description={organizer(choice.registration.contest)}
                        aside={choice.kind === 'ineligible' ? choice.reason : `${scoreFormat.format(log.score)} ${pointsUnit(log.score)}`}
                        disabled={choice.kind === 'ineligible'}
                      />
                    )}
                  </form.AppField>
                ))}
                </div>
                {eligible.length ? <p className="mb-0 mt-2 text-metadata text-muted">Unchecking removes this log’s score from that contest.</p> : null}
                {choices.some(choice => choice.kind === 'ineligible')
                  ? <p className="mb-0 mt-1 text-metadata text-muted">Registered for the wrong language? <a className="text-link hover:text-link-hover" href={links.contests}>Update your registration</a> from its contest page.</p>
                  : null}
                {finished.map(submission => (
                  <p key={submission.registration_id}>
                    {submission.title} has finished. Its saved score
                    {submission.score === undefined ? '' : ` of ${submission.score} points`} stays.
                  </p>
                ))}
              </fieldset>
              {eligible.length ? null : (
                <EmptyState
                  variant="plain"
                  headingLevel={2}
                  title="No eligible contests"
                  description={choices.length
                    ? 'This activity already counts in your personal history.'
                    : 'No contest registrations found. You can submit to contests after joining one. This activity already counts in your personal history.'}
                  actions={<a className="text-link hover:text-link-hover" href={links.contests}>Find a contest to join</a>}
                />
              )}
            </fieldset>

            {failure ? <SaveFailure failure={failure} /> : null}
            <div className={formFooter().root()}>
              <div className={formFooter().actions()}>
                <Link className={buttonClassName({ variant: 'ghost' })} to="/logs/$logId" params={{ logId: log.id }}>Cancel</Link>
                <Button type="submit" disabled={disabled || !eligible.length}>
                  {pending ? 'Saving…' : 'Save submissions'}
                </Button>
              </div>
            </div>
          </form.Form>
        </form.AppForm>
      </Surface>
    </div>
  )
}

/** A registration read from the owner's contest history, as an administrator does, does not name the organizer. */
const organizer = (contest: Registration['contest']) => {
  const name = contest.official ? 'Tadoku' : contest.owner_user_display_name
  return name ? `Organized by ${name}` : undefined
}

/** The log being submitted, as Webv2 shows it above the contests: what, how much and its score. */
function LogSummary({ log }: { log: ImmersionLog }) {
  const what = [log.language.name, log.activity.name, formatTracking(log)].filter(Boolean).join(' · ')

  return (
    <section
      className={[
        'grid grid-cols-content-end items-baseline gap-x-4 gap-y-0.5 border-b border-rule bg-canvas',
        '-mx-record-inset -mt-record-inset mb-6 px-record-inset py-4 [&_p]:m-0',
      ].join(' ')}
      aria-label="Your log"
    >
      <p className="col-span-full text-metadata text-muted">Your log</p>
      <p className="font-semibold">{what}</p>
      {log.description ? <p className="col-start-1 text-sm text-muted wrap-anywhere"><LogDescription log={{ ...log, description: log.description }} /></p> : null}
      <p className="col-start-2 row-start-2 whitespace-nowrap font-bold tabular-nums">
        {scoreFormat.format(log.score)} <small className="text-metadata font-regular text-muted">{pointsUnit(log.score)}</small>
      </p>
    </section>
  )
}

const failureMessages: Record<LogFailure, string> = {
  invalid: 'Your submissions were not saved. A contest may have ended or your registration changed; reload and try again.',
  forbidden: 'Your account cannot change this activity\'s contests.',
  'not-found': 'This activity no longer exists. Nothing was saved.',
  deleting: 'Your account is being deleted. Contest submissions cannot change.',
  'signed-out': 'Your session ended. Log in in a new tab, return here and save again.',
  switched: 'You are now signed in as a different account. Nothing was saved.',
  'account-unavailable': 'Account service unavailable. Nothing was saved. Try again when the service recovers.',
  unavailable: 'We could not confirm whether your submissions were saved. Saving again is safe.',
}

function SaveFailure({ failure }: { failure: LogFailure }) {
  const loginUrl = useLoginUrl()
  return (
    <div role="alert">
      <p>{failureMessages[failure]}</p>
      {failure === 'signed-out'
        ? <a className="text-link hover:text-link-hover" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
        : null}
    </div>
  )
}
