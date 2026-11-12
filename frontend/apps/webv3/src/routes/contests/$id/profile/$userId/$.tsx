import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { EmptyState, buttonClassName } from 'paper-ui'
import { z } from 'zod'
import { ContestProfile } from '../../../../../components/ContestProfile'
import { PendingView } from '../../../../../components/StatusViews'
import { contestLogsQuery, contestProfileQuery, contestSummaryQuery } from '../../../../../lib/contest-activity'
import { filterCount, noFilters, parseFilters, toLogFilters, type ProfileFilters } from '../../../../../lib/contest-filters'
import { contestsHref } from '../../../../../lib/contests'
import { contestEntryQuery, pageNumber } from '../../../../../lib/leaderboard'
import { viewerQueryOptions } from '../../../../../lib/session'
import { pageUnavailable } from '../../../../../lib/page-status'
import { noindex, sharingMeta } from '../../../../../lib/page-meta'

export const Route = createFileRoute('/contests/$id/profile/$userId/$')({
  // Unknown or malformed filter values are dropped here, so the address and the API only ever see valid ones.
  validateSearch: (search: Record<string, unknown>): ProfileFilters => parseFilters(search),
  loaderDeps: ({ search }) => ({ filters: search }),
  loader: async ({ context: { queryClient }, params, deps, cause }) => {
    const input = profileInput(params)
    const page = pageNumber(params._splat)
    const filters = toLogFilters(deps.filters)
    const filtered = [
      queryClient.ensureQueryData(contestSummaryQuery({ ...input, filters })),
      filterCount(deps.filters) ? queryClient.ensureQueryData(contestSummaryQuery({ ...input, filters: toLogFilters(noFilters) })) : null,
      queryClient.ensureQueryData(contestLogsQuery({ ...input, page, filters })),
    ]
    // A filter or page change keeps showing the current figures while the new ones load, so an open Filters menu stays.
    const [, profile] = await Promise.all([
      queryClient.ensureQueryData(viewerQueryOptions()),
      queryClient.ensureQueryData(contestProfileQuery(input)),
      queryClient.ensureQueryData(contestEntryQuery(input.contestId, input.userId)),
      ...(cause === 'stay' ? filtered.map(read => void read?.catch(() => undefined)) : filtered),
    ])
    if (profile.status === 'not-found') throw notFound()
    if (profile.status === 'unavailable') pageUnavailable(queryClient, contestProfileQuery(input))
    const { user_display_name: name, contest } = profile.registration
    return {
      name: `${name} · ${contest.title}`,
      description: `${name}’s scores and logs in ${contest.title} on Tadoku.`,
      private: contest.private,
    }
  },
  head: ({ loaderData }) => ({
    meta: loaderData
      ? [...sharingMeta(loaderData.name, loaderData.description), ...(loaderData.private ? [noindex] : [])]
      : [{ title: 'Contest profile · Tadoku' }],
  }),
  notFoundComponent: () => (
    <EmptyState
      headingLevel={1}
      title="Participant not found"
      description="This contest has no participant with this address."
      actions={<Link className={buttonClassName()} to={contestsHref.official()}>Browse contests</Link>}
    />
  ),
  component: ContestProfileRoute,
})

function profileInput(params: { id: string; userId: string }) {
  if (!z.guid().safeParse(params.id).success || !z.guid().safeParse(params.userId).success) throw notFound()
  return { contestId: params.id, userId: params.userId }
}

function ContestProfileRoute() {
  const params = Route.useParams()
  const filters = Route.useSearch()
  const input = profileInput(params)
  const profile = useQuery(contestProfileQuery(input))

  if (!profile.data) return <PendingView />
  if (profile.data.status !== 'found') {
    return <EmptyState headingLevel={1} title="Contest profile unavailable" description="Try reloading this page." />
  }
  return <ContestProfile profile={profile.data} page={pageNumber(params._splat)} filters={filters} />
}
