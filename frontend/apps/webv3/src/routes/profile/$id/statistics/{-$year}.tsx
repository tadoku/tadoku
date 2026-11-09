import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { EmptyState } from 'paper-ui'
import { z } from 'zod'
import { ProfileStatistics } from '../../../../components/ProfileStatistics'
import { PendingView } from '../../../../components/StatusViews'
import { profileStatisticsQuery } from '../../../../lib/profile'
import { ensureViewerData, viewerKey, viewerQueryOptions } from '../../../../lib/session'
import { pageUnavailable } from '../../../../lib/page-status'

export const Route = createFileRoute('/profile/$id/statistics/{-$year}')({
  validateSearch: z.object({ contests: z.number().int().min(1).optional().catch(undefined) }),
  loaderDeps: ({ search }) => ({ page: search.contests ?? 1 }),
  loader: async ({ context, params, deps }) => {
    const input = statisticsInput(params, deps.page, new Date().getUTCFullYear())
    const [viewer, result] = await ensureViewerData(context.queryClient, key => profileStatisticsQuery(input, key))
    if (result.status === 'not-found') throw notFound()
    if (result.status === 'unavailable') pageUnavailable(context.queryClient, profileStatisticsQuery(input, viewerKey(viewer)))
    return { name: result.status === 'found' ? result.profile.display_name : undefined, year: input.year }
  },
  head: ({ loaderData }) => ({
    meta: [{ title: loaderData?.name ? `Profile statistics · ${loaderData.name} · Tadoku` : 'Profile · Tadoku' }],
  }),
  notFoundComponent: () => (
    <EmptyState
      headingLevel={1}
      title="Profile not found"
      description="No Tadoku account has this address."
      actions={<Link to="/" className="text-link hover:text-link-hover">Return home</Link>}
    />
  ),
  component: ProfileStatisticsPage,
})

/**
 * The requested profile and year; the year defaults to `currentYear`, the UTC year the loader read, as the API counts
 * years in UTC.
 */
function statisticsInput(params: { id: string; year?: string }, page: number, currentYear: number) {
  if (!z.guid().safeParse(params.id).success) throw notFound()
  if (params.year !== undefined && !/^\d{4}$/.test(params.year)) throw notFound()
  return { userId: params.id, year: Number(params.year ?? currentYear), page }
}

function ProfileStatisticsPage() {
  const input = statisticsInput(Route.useParams(), Route.useLoaderDeps().page, Route.useLoaderData().year)
  const { data: viewer } = useQuery(viewerQueryOptions())
  const { data: result } = useQuery(profileStatisticsQuery(input, viewerKey(viewer)))

  if (!result) return <PendingView />
  if (result.status !== 'found') {
    return <EmptyState headingLevel={1} title="Profile unavailable" description="Try reloading this page." />
  }
  const owner = viewer?.status === 'signed-in' && viewer.userId === result.profile.id
  return <ProfileStatistics statistics={result} year={input.year} page={input.page} owner={owner} />
}
