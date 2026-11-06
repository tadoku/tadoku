import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { Link, createFileRoute, notFound, useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, EmptyState, PageHeader, Surface, buttonClassName, usePaperForm } from 'paper-ui'
import { useState } from 'react'
import { z } from 'zod'
import { DraftOwnerGate } from '../../components/DraftOwnerGate'
import { useLinks } from '../../lib/links'
import { contestChoices, type ContestChoice } from '../../lib/log-draft'
import {
  logOptionsQuery,
  logQuery,
  useUpdateLogContests,
  type ImmersionLog,
  type LogFailure,
  type Registration,
} from '../../lib/logs'
import { useLoginUrl, viewerQueryOptions } from '../../lib/session'

export const Route = createFileRoute('/logs/$logId_/contests')({
  head: () => ({ meta: [{ title: 'Submit to contests · Tadoku' }] }),
  // Webv2's `?preselect=1` selects every eligible contest for a log not yet submitted to any.
  // The router reads `1` as a number; a string would be written back to the URL as `"1"`.
  validateSearch: z.object({ preselect: z.coerce.number().optional().catch(undefined) }),
  loader: async ({ context, params }) => {
    if (!z.guid().safeParse(params.logId).success) throw notFound()
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    if (viewer.status !== 'signed-in') return

    const [log] = await Promise.all([
      context.queryClient.ensureQueryData(logQuery(params.logId, viewer.userId)),
      context.queryClient.ensureQueryData(logOptionsQuery(viewer.userId)),
    ])
    if (!log.ok && log.reason === 'not-found') throw notFound()
  },
  component: LogContests,
})

const signIn = {
  title: 'Sign in to submit to contests',
  description: 'Log in to choose the contests your activity counts in.',
}

function LogContests() {
  return <DraftOwnerGate signIn={signIn}>{viewerId => <ContestsDraft viewerId={viewerId} />}</DraftOwnerGate>
}

/** Only the log's owner changes its submissions. The choice starts from the log as first read. */
function ContestsDraft({ viewerId }: { viewerId: string }) {
  const { logId } = Route.useParams()
  const logResult = useQuery({ ...logQuery(logId, viewerId), staleTime: Infinity })
  const optionsResult = useQuery(logOptionsQuery(viewerId))
  const loginUrl = useLoginUrl()
  const log = logResult.data
  const options = optionsResult.data

  if (logResult.isPending || optionsResult.isPending) return null
  if (log?.ok && log.data.user_id !== viewerId) {
    return (
      <EmptyState
        headingLevel={1}
        title="This is someone else's record"
        description="Only the learner who logged this activity can submit it to contests."
        actions={<Link className="text-link" to="/logs/$logId" params={{ logId }}>View the log</Link>}
      />
    )
  }
  if (log?.ok && options?.ok) {
    return <ContestsForm log={log.data} registrations={options.data.registrations} viewerId={viewerId} />
  }

  if (log?.ok === false && log.reason === 'not-found') {
    return <EmptyState headingLevel={1} title="Activity not found" description="This log may have been deleted." />
  }
  if ((log?.ok === false && log.reason === 'signed-out') || (options?.ok === false && options.reason === 'signed-out')) {
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
      actions={<Button onClick={() => void Promise.all([logResult.refetch(), optionsResult.refetch()])}>Try again</Button>}
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
    <div className="goal-page">
      <PageHeader
        title="Submit to contests"
        lead={[log.language.name, log.activity.name, log.description].filter(Boolean).join(' · ')}
        backLink={{ href: `/logs/${log.id}`, label: 'Back to log' }}
      />
      <Surface className="log-surface">
        <form.AppForm>
          <form.Form className="goal-form" aria-busy={pending}>
            <fieldset className="goal-form__controls" disabled={disabled}>
              <fieldset className="paper-choice-field log-contest-choices">
                <legend className="paper-visually-hidden">Contests</legend>
                {choices.filter(choice => choice.kind !== 'ineligible').map(choice => (
                  <form.AppField key={choice.registration.id} name="registrationIds">
                    {field => <field.Checkbox label={choice.registration.contest.title} value={choice.registration.id} />}
                  </form.AppField>
                ))}
                {/* Unavailable contests are grouped under their shared reason instead of repeating it on every row. */}
                {ineligibleByReason(choices).map(([reason, group]) => (
                  <div key={reason} className="log-contest-group" role="group" aria-label={reason}>
                    <p className="log-contest-group__reason">{reason}</p>
                    {group.map(choice => (
                      <form.AppField key={choice.registration.id} name="registrationIds">
                        {field => <field.Checkbox label={choice.registration.contest.title} value={choice.registration.id} disabled />}
                      </form.AppField>
                    ))}
                  </div>
                ))}
                {finished.map(submission => (
                  <p key={submission.registration_id}>
                    {submission.title} has finished. Its saved score
                    {submission.score === undefined ? '' : ` of ${submission.score} points`} stays.
                  </p>
                ))}
              </fieldset>
              {eligible.length ? (
                <p className="log-contests-note">Unchecking removes this log’s score from that contest.</p>
              ) : (
                <EmptyState
                  variant="plain"
                  headingLevel={2}
                  title="No eligible contests"
                  description={choices.length
                    ? 'This activity already counts in your personal history.'
                    : 'No contest registrations found. You can submit to contests after joining one. This activity already counts in your personal history.'}
                  actions={<a className="text-link" href={links.contests}>Find a contest to join</a>}
                />
              )}
            </fieldset>

            {failure ? <SaveFailure failure={failure} /> : null}
            <div className="goal-actions">
              <Button type="submit" disabled={disabled || !eligible.length}>
                {pending ? 'Saving…' : 'Save submissions'}
              </Button>
              <Link className="text-link" to="/logs/$logId" params={{ logId: log.id }}>Cancel</Link>
            </div>
          </form.Form>
        </form.AppForm>
      </Surface>
    </div>
  )
}

/** Ineligible choices grouped by their reason, in first-seen order. */
function ineligibleByReason(choices: ContestChoice[]) {
  const groups = new Map<string, Extract<ContestChoice, { kind: 'ineligible' }>[]>()
  for (const choice of choices) {
    if (choice.kind === 'ineligible') groups.set(choice.reason, [...(groups.get(choice.reason) ?? []), choice])
  }
  return [...groups]
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
        ? <a className="text-link" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
        : null}
    </div>
  )
}
