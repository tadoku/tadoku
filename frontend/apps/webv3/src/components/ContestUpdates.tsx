import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { useHydrated } from '@tanstack/react-router'
import type { ImmersionContestView, ImmersionLog } from 'api-client'
import { Button, EmptyState, Flash, Modal, usePaperForm } from 'paper-ui'
import { useState } from 'react'
import { contestLogsQuery, detachSchema, useDetachLog } from '../lib/contest-activity'
import type { LogFailure } from '../lib/logs'
import { useLinks } from '../lib/links'
import { useLoginUrl, viewerQueryOptions } from '../lib/session'
import { LogEntries } from './LogEntries'

type Contest = ImmersionContestView & { id: string }

/** A contest's updates tab: every participant's logs, newest first. */
export function ContestUpdates({ contest, page }: { contest: Contest; page: number }) {
  const links = useLinks()
  return (
    <section className="contest-updates" aria-label="Updates">
      <ContestLogs contest={contest} page={page} pageHref={next => links.contestUpdates(contest.id, next)} showParticipant />
    </section>
  )
}

/**
 * A page of a contest's logs, optionally one participant's. As in Webv2, the contest's owner and administrators can
 * remove a log from the contest; the API refuses anyone else.
 */
export function ContestLogs({ contest, userId, page, pageHref, showParticipant = false }: {
  contest: Contest
  userId?: string
  page: number
  pageHref: (page: number) => string
  showParticipant?: boolean
}) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const logs = useQuery(contestLogsQuery({ contestId: contest.id, userId, page }))
  const [removed, setRemoved] = useState(false)
  const moderatorId = viewer?.status === 'signed-in' && (viewer.role === 'admin' || viewer.userId === contest.owner_user_id)
    ? viewer.userId
    : undefined

  if (!logs.data) return <p className="profile-muted" role="status">Loading updates…</p>
  if (!logs.data.ok) {
    return (
      <EmptyState
        variant="plain"
        headingLevel={3}
        title="Could not load updates"
        description="Try loading the updates again in a moment."
        actions={<Button variant="outline" loading={logs.isFetching} onClick={() => void logs.refetch()}>Try again</Button>}
      />
    )
  }

  return (
    <>
      {removed ? <Flash variant="success" className="contest-updates-notice" role="status">Log removed from contest.</Flash> : null}
      <LogEntries
        logs={logs.data.data.logs}
        total={logs.data.data.total_size}
        page={page}
        pageHref={pageHref}
        showParticipant={showParticipant}
        actions={moderatorId
          ? log => <RemoveFromContest key={moderatorId} log={log} contestId={contest.id} viewerId={moderatorId} onRemoved={() => setRemoved(true)} />
          : undefined}
        empty={<EmptyState variant="plain" headingLevel={3} title="No updates submitted yet" description="Logs submitted to this contest appear here." />}
      />
    </>
  )
}

/** Webv2's confirmation: a moderator gives a reason before a log leaves the contest. */
function RemoveFromContest({ log, contestId, viewerId, onRemoved }: {
  log: ImmersionLog
  contestId: string
  viewerId: string
  onRemoved: () => void
}) {
  const hydrated = useHydrated()
  const detach = useDetachLog()
  const [open, setOpen] = useState(false)
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const form = usePaperForm({
    defaultValues: { reason: '' },
    onSubmit: async ({ value }) => {
      setFailure(null)
      const result = await detach({ viewerId, contestId, logId: log.id, reason: value.reason })
      if (!result.ok) {
        setFailure(result.reason)
        return
      }
      setOpen(false)
      onRemoved()
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)

  return (
    <Modal
      trigger={<Button variant="ghost" disabled={!hydrated}>Remove from contest</Button>}
      title="Remove log from contest"
      open={open}
      onOpenChange={next => {
        if (pending) return
        setOpen(next)
        setFailure(null)
        if (!next) form.reset()
      }}
      footer={null}
    >
      <form.AppForm>
        <form.Form className="goal-form" aria-busy={pending}>
          <p>Why are you removing this log from the contest? Please provide a reason.</p>
          <form.AppField name="reason" validators={{ onSubmit: ({ value }) => reasonError(value) }}>
            {field => <field.TextArea label="Reason" placeholder="e.g. Duplicate entry, incorrect data…" rows={4} required disabled={pending} />}
          </form.AppField>
          {failure ? <DetachFailure failure={failure} /> : null}
          <div className="registration-actions">
            <Button type="submit" variant="destructive" disabled={pending}>{pending ? 'Removing…' : 'Yes, remove it'}</Button>
            <Button type="button" variant="ghost" disabled={pending} onClick={() => setOpen(false)}>Cancel</Button>
          </div>
        </form.Form>
      </form.AppForm>
    </Modal>
  )
}

/** The reason the API accepts: 1 to 1,000 characters once trimmed. */
function reasonError(value: string) {
  if (detachSchema.shape.reason.safeParse(value).success) return undefined
  return value.trim() ? 'Keep the reason under 1,000 characters.' : 'Give a reason for removing this log.'
}

function DetachFailure({ failure }: { failure: LogFailure }) {
  const loginUrl = useLoginUrl()
  return (
    <Flash variant="danger" action={failure === 'signed-out' ? <a className="text-link" href={loginUrl}>Log in</a> : undefined}>
      {detachFailures[failure]}
    </Flash>
  )
}

const detachFailures: Record<LogFailure, string> = {
  'signed-out': 'Your session ended, so the log was not removed. Log in and try again.',
  switched: 'You are now signed in as a different account, so the log was not removed. Reload the page.',
  'account-unavailable': 'We could not check your session, so the log was not removed. Try again in a moment.',
  forbidden: 'Your account cannot remove logs from this contest.',
  invalid: 'Give a reason under 1,000 characters.',
  'not-found': 'This log is no longer part of the contest. Reload the page.',
  deleting: 'Your account is being deleted.',
  unavailable: 'Failed to remove log. Reload the page to check whether it was removed.',
}
