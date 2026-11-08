import { useQuery } from '@tanstack/react-query'
import { PencilSquareIcon, TrashIcon, iconClassName } from 'paper-ui/icons'
import { Link, createFileRoute, notFound, useHydrated } from '@tanstack/react-router'
import { Button, buttonClassName, DescriptionList, EmptyState, entryLink, Modal, PageHeader, SectionHeading, StatGrid, Surface } from 'paper-ui'
import { useEffect, useRef, useState } from 'react'
import { z } from 'zod'
import { formatDate, pointsUnit, scoreFormat } from '../../lib/format'
import { useLinks } from '../../lib/links'
import { logQuery, useDeleteLog, type ImmersionLog, type LogFailure } from '../../lib/logs'
import { viewerKey, viewerQueryOptions } from '../../lib/session'
import { ActivityName } from '../../components/ActivityName'
import { renderLink } from '../../components/router-link'

export const Route = createFileRoute('/logs/$logId')({
  head: () => ({ meta: [{ title: 'Activity · Tadoku' }] }),
  loader: async ({ context, params }) => {
    if (!z.guid().safeParse(params.logId).success) throw notFound()
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    const result = await context.queryClient.ensureQueryData(logQuery(params.logId, viewerKey(viewer)))
    if (!result.ok && result.reason === 'not-found') throw notFound()
  },
  component: LogDetail,
})

function LogDetail() {
  const { logId } = Route.useParams()
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const { data: result } = useQuery(logQuery(logId, viewerKey(viewer)))
  useFocusPageHeading()

  if (!result?.ok) {
    const forbidden = result?.ok === false && result.reason === 'forbidden'
    return (
      <EmptyState
        headingLevel={1}
        title={forbidden ? 'Activity not available to your account' : 'Activity unavailable'}
        description="Try reloading this page."
      />
    )
  }

  const log = result.data
  const owner = viewer?.status === 'signed-in' && viewer.userId === log.user_id
  const learner = log.user_display_name ?? (owner && viewer?.status === 'signed-in' ? viewer.displayName : undefined)
  return (
    <div>
      <PageHeader
        backLink={{ href: links.profileUpdates(log.user_id), label: owner ? 'Your activity' : 'Activity' }}
        renderLink={renderLink}
        className="max-w-record"
        title={`${log.language.name} ${log.activity.name.toLocaleLowerCase()}`}
        lead={(
          <>
            {log.language.name} · <ActivityName activity={log.activity} />
            {learner && !owner ? <> · Logged by <a className="text-link hover:text-link-hover" href={links.profile(log.user_id)}>{learner}</a></> : null}
          </>
        )}
        actions={owner ? <OwnerActions log={log} /> : null}
      />
      <Surface className="max-w-record p-0 wrap-anywhere">
        <div className="grid gap-6 p-record-inset">
          <StatGrid
            stats={[
              ...measure(log),
              { label: 'Score', value: scoreFormat.format(log.score), unit: pointsUnit(log.score) },
              { label: 'Logged', value: formatDate(log.created_at.slice(0, 10)), detail: loggedTime(log.created_at) },
            ]}
          />
          {log.description || log.tags.length ? (
            <DescriptionList
              className="border-t border-rule pt-6"
              items={[
                ...(log.description ? [{ label: 'Description', value: log.description }] : []),
                ...(log.tags.length ? [{ label: 'Tags', value: log.tags.join(', ') }] : []),
              ]}
            />
          ) : null}
        </div>
        <LogContests log={log} owner={owner} />
      </Surface>
    </div>
  )
}

/** Editing and deleting, offered only to the log's owner. The API refuses both to anyone else. */
function OwnerActions({ log }: { log: ImmersionLog }) {
  const links = useLinks()
  const deleteLog = useDeleteLog()
  const hydrated = useHydrated()
  const [open, setOpen] = useState(false)
  const [pending, setPending] = useState(false)
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const keepRef = useRef<HTMLButtonElement>(null)

  const confirm = async () => {
    setPending(true)
    setFailure(null)
    const result = await deleteLog(log.user_id, log.id)
    if (result.ok) {
      window.location.assign(links.profile(log.user_id))
      return
    }
    setPending(false)
    setFailure(result.reason)
  }

  return (
    <>
      <Modal
        trigger={(
          <Button variant="destructive-ghost" disabled={!hydrated} leadingIcon={<TrashIcon className={iconClassName('compact')} />}>
            Delete log
          </Button>
        )}
        title="Delete this log?"
        initialFocus={keepRef}
        open={open}
        onOpenChange={next => {
          if (pending) return
          setOpen(next)
          setFailure(null)
        }}
        footer={(
          <>
            <Button ref={keepRef} variant="outline" disabled={pending} onClick={() => setOpen(false)}>Keep log</Button>
            <Button variant="destructive" disabled={pending} onClick={() => void confirm()}>
              {pending ? 'Deleting…' : 'Delete log'}
            </Button>
          </>
        )}
      >
        <p>
          Deletion cannot be undone. The log will be permanently removed from all contests and your personal tracking
          history.
        </p>
        {failure ? <p role="alert">{deleteFailures[failure] ?? deleteFailures.unavailable}</p> : null}
      </Modal>
      <Link className={buttonClassName({ variant: 'outline' })} to="/logs/$logId/edit" params={{ logId: log.id }}>
        <PencilSquareIcon className={iconClassName('compact')} aria-hidden="true" />Edit log
      </Link>
    </>
  )
}

const deleteFailures: Partial<Record<LogFailure, string>> = {
  switched: 'You are now signed in as a different account. Nothing was deleted.',
  'signed-out': 'Your session ended. Log in again, then delete the log.',
  'account-unavailable': 'Account service unavailable. Nothing was deleted. Try again later.',
  forbidden: 'This log cannot be deleted. A log submitted to a finished contest stays in its results.',
  deleting: 'Your account is being deleted.',
  unavailable: 'We could not confirm the deletion. Reload this page to check whether the log still exists.',
}

/** Contest submissions are private: only the log's owner sees them and the way to manage them. */
function LogContests({ log, owner }: { log: ImmersionLog; owner: boolean }) {
  return (
    <section className="border-t border-rule px-record-inset pb-record-inset pt-6" aria-labelledby="log-contests">
      <SectionHeading
        id="log-contests"
        level={2}
        title="Contests"
        actions={owner ? <Link className="text-link hover:text-link-hover" to="/logs/$logId/contests" params={{ logId: log.id }}>Manage contest submissions</Link> : null}
      />
      {owner ? <OwnContests log={log} /> : <p className="m-0 text-muted">Contest submissions are visible to the log&apos;s owner.</p>}
    </section>
  )
}

function OwnContests({ log }: { log: ImmersionLog }) {
  const links = useLinks()
  const registrations = log.registrations ?? []

  return registrations.length > 0 ? (
    <ul className="m-0 list-none p-0">
      {registrations.map(registration => (
        <li key={registration.contest_id} className="flex items-baseline justify-between gap-4 py-3 [&+&]:border-t [&+&]:border-rule">
          <a className={entryLink({ className: 'profile-entry-title' })} href={links.contest(registration.contest_id)}>{registration.title}</a>
          {registration.score !== undefined ? (
            <span className="whitespace-nowrap font-bold tabular-nums">
              {scoreFormat.format(registration.score)}{' '}
              <small className="text-metadata font-regular text-muted">{pointsUnit(registration.score)}</small>
            </span>
          ) : null}
        </li>
      ))}
    </ul>
  ) : <p className="m-0 text-muted">Saved to your personal record only.</p>
}

/** Moves focus to the page title once the page is interactive, so keyboard users land on the record they opened. */
function useFocusPageHeading() {
  const hydrated = useHydrated()
  useEffect(() => {
    if (!hydrated) return
    const heading = document.querySelector<HTMLElement>('h1')
    heading?.setAttribute('tabindex', '-1')
    heading?.focus()
  }, [hydrated])
}

const amountFormat = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 2 })

/** A timed log's duration; an amount log's amount and any time spent, for example "18 pages". */
function measure(log: ImmersionLog) {
  const unit = log.unit_name ? ` ${log.unit_name.toLowerCase()}${log.amount === 1 ? '' : 's'}` : ''
  const amount = { label: 'Amount', value: `${amountFormat.format(log.amount)}${unit}` }
  if (!log.duration_seconds) return [amount]

  const duration = { label: 'Duration', value: `${log.duration_seconds / 60} min` }
  return log.activity.input_type === 'time_primary' ? [duration] : [amount, duration]
}

/** For example "06:19 UTC". */
function loggedTime(iso: string) {
  return `${new Date(iso).toISOString().slice(11, 16)} UTC`
}
