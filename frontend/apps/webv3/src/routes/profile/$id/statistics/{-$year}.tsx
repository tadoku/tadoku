import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { EmptyState, linkStyles } from 'paper-ui'
import { z } from 'zod'
import { ProfileStatistics } from '../../../../components/ProfileStatistics'
import { PendingView } from '../../../../components/StatusViews'
import { profileStatisticsQuery } from '../../../../lib/profile'
import { viewerKey, viewerQueryOptions } from '../../../../lib/session'

export const Route = createFileRoute('/profile/$id/statistics/{-$year}')({
  validateSearch: z.object({ contests: z.number().int().min(1).optional().catch(undefined) }),
  loaderDeps: ({ search }) => ({ page: search.contests ?? 1 }),
  loader: async ({ context, params, deps }) => {
    const input = statisticsInput(params, deps.page)
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    const result = await context.queryClient.ensureQueryData(profileStatisticsQuery(input, viewerKey(viewer)))
    if (result.status === 'not-found') throw notFound()
    return { name: result.status === 'found' ? result.profile.display_name : undefined }
  },
  head: ({ loaderData }) => ({
    meta: [{ title: loaderData?.name ? `Profile statistics · ${loaderData.name} · Tadoku` : 'Profile · Tadoku' }],
  }),
  notFoundComponent: () => (
    <EmptyState
      headingLevel={1}
      title="Profile not found"
      description="No Tadoku account has this address."
      actions={<Link to="/" {...stylex.props(linkStyles.base)}>Return home</Link>}
    />
  ),
  component: ProfileStatisticsPage,
})

/** The requested profile and year; the year defaults to the current UTC year, as the API counts years in UTC. */
function statisticsInput(params: { id: string; year?: string }, page: number) {
  if (!z.guid().safeParse(params.id).success) throw notFound()
  if (params.year !== undefined && !/^\d{4}$/.test(params.year)) throw notFound()
  return { userId: params.id, year: Number(params.year ?? new Date().getUTCFullYear()), page }
}

function ProfileStatisticsPage() {
  const input = statisticsInput(Route.useParams(), Route.useLoaderDeps().page)
  const { data: viewer } = useQuery(viewerQueryOptions())
  const { data: result } = useQuery(profileStatisticsQuery(input, viewerKey(viewer)))

  if (!result) return <PendingView />
  if (result.status !== 'found') {
    return <EmptyState headingLevel={1} title="Profile unavailable" description="Try reloading this page." />
  }
  const owner = viewer?.status === 'signed-in' && viewer.userId === result.profile.id
  return <ProfileStatistics statistics={result} year={input.year} page={input.page} owner={owner} />
}
