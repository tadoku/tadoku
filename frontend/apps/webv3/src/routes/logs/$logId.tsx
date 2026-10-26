import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound, useHydrated } from '@tanstack/react-router'
import { DescriptionList, EmptyState, PageHeader, SectionHeading, StatGrid, Surface } from 'paper-ui'
import { useEffect } from 'react'
import { z } from 'zod'
import { useLinks } from '../../lib/links'
import { logQuery, type ImmersionLog } from '../../lib/logs'
import { viewerKey, viewerQueryOptions } from '../../lib/session'

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
  const subject = `${log.language.name} ${log.activity.name}`
  return (
    <div className="goal-page">
      <PageHeader title={log.description || subject} lead={`${log.language.name} · ${log.activity.name}`} />
      <StatGrid stats={[measure(log), { label: 'Score', value: scoreFormat.format(log.score) }]} />
      <Surface className="log-surface">
        <DescriptionList
          items={[
            ...(log.user_display_name ? [{ label: 'Learner', value: log.user_display_name }] : []),
            { label: 'Created', value: utcTimestamp(log.created_at) },
            { label: 'Tags', value: log.tags.join(', ') || 'No tags' },
            { label: 'Record ID', value: log.id },
          ]}
        />
      </Surface>
      <LogContests log={log} owner={owner} />
    </div>
  )
}

/** Contest submissions are private: only the log's owner sees them and the way to manage them. */
function LogContests({ log, owner }: { log: ImmersionLog; owner: boolean }) {
  return (
    <section aria-labelledby="log-contests">
      <SectionHeading id="log-contests" level={2} title="Contests" />
      {owner ? <OwnContests log={log} /> : <p>Contest submissions are visible to the log&apos;s owner.</p>}
    </section>
  )
}

function OwnContests({ log }: { log: ImmersionLog }) {
  const links = useLinks()
  const registrations = log.registrations ?? []

  return (
    <>
      {registrations.length > 0 ? (
        <ul>
          {registrations.map(registration => (
            <li key={registration.contest_id}>
              <a className="text-link" href={links.contest(registration.contest_id)}>{registration.title}</a>
              {registration.score !== undefined ? ` · ${registration.score} points` : ''}
            </li>
          ))}
        </ul>
      ) : <p>Saved to your personal record only.</p>}
      <a className="text-link" href={links.legacyUrl(`/logs/${log.id}/contests`)}>Manage contest submissions</a>
    </>
  )
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

const scoreFormat = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 2 })

function measure(log: ImmersionLog) {
  return log.duration_seconds
    ? { label: 'Duration', value: `${log.duration_seconds / 60} min` }
    : { label: 'Amount', value: `${log.amount} ${log.unit_name}` }
}

/** For example `2026-10-25 06:19:03.796 UTC`. */
function utcTimestamp(iso: string) {
  return new Date(iso).toISOString().replace('T', ' ').replace('Z', ' UTC')
}
