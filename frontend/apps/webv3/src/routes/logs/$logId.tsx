import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound, useHydrated } from '@tanstack/react-router'
import { DescriptionList, EmptyState, PageHeader, SectionHeading, StatGrid, Surface } from 'paper-ui'
import { useEffect } from 'react'
import { z } from 'zod'
import { logQueryOptions } from '../../lib/logs'
import { viewerQueryOptions } from '../../lib/session'
import { useLinks } from '../../lib/links'

export const Route = createFileRoute('/logs/$logId')({
  head: () => ({ meta: [{ title: 'Activity · Tadoku' }] }),
  loader: async ({ context, params }) => {
    if (!z.uuid().safeParse(params.logId).success) throw notFound()
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    const viewerId = viewer.status === 'signed-in' || viewer.status === 'banned' ? viewer.userId : viewer.status
    const result = await context.queryClient.ensureQueryData(logQueryOptions(params.logId, viewerId))
    if (!result.ok && result.reason === 'not-found') throw notFound()
  },
  component: LogDetail,
})

function LogDetail() {
  const { logId } = Route.useParams()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const viewerId = viewer?.status === 'signed-in' || viewer?.status === 'banned' ? viewer.userId : viewer?.status ?? 'anonymous'
  const { data: result } = useQuery(logQueryOptions(logId, viewerId))
  const links = useLinks()
  const hydrated = useHydrated()
  useEffect(() => { if (hydrated) { const heading = document.querySelector('h1'); heading?.setAttribute('tabindex', '-1'); heading?.focus() } }, [hydrated])
  if (!result?.ok) return <EmptyState headingLevel={1} title={result && !result.ok && result.reason === 'forbidden' ? 'Activity not available to your account' : 'Activity unavailable'} description="Try reloading this page." />
  const log = result.data
  const owner = viewer?.status === 'signed-in' && viewer.userId === log.user_id
  return <div className="goal-page">
    <PageHeader title={log.description || `${log.language.name} ${log.activity.name}`} lead={`${log.language.name} · ${log.activity.name}`} />
    <StatGrid stats={[{ label: log.duration_seconds ? 'Duration' : 'Amount', value: log.duration_seconds ? `${log.duration_seconds / 60} min` : `${log.amount} ${log.unit_name}` }, { label: 'Score', value: new Intl.NumberFormat('en-GB', { maximumFractionDigits: 2 }).format(log.score) }]} />
    <Surface className="log-surface"><DescriptionList items={[
      { label: 'Learner', value: log.user_display_name },
      { label: 'Created', value: new Date(log.created_at).toISOString().replace('T', ' ').replace('Z', ' UTC') },
      { label: 'Tags', value: log.tags.join(', ') || 'No tags' },
      { label: 'Record ID', value: log.id },
    ]} /></Surface>
    <section aria-labelledby="log-contests"><SectionHeading id="log-contests" level={2} title="Contests" />
      {owner ? <>{log.registrations.length ? <ul>{log.registrations.map(registration => <li key={registration.contest_id}><a className="text-link" href={links.contest(registration.contest_id)}>{registration.title}</a>{registration.score !== undefined ? ` · ${registration.score} points` : ''}</li>)}</ul> : <p>Saved to your personal record only.</p>}<a className="text-link" href={links.legacyUrl(`/logs/${log.id}/contests`)}>Manage contest submissions</a></> : <p>Contest submissions are visible to the log&apos;s owner.</p>}
    </section>
  </div>
}
