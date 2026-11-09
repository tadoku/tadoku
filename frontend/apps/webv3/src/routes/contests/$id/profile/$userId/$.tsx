import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link, notFound } from '@tanstack/react-router'
import { EmptyState, buttonClassName } from 'paper-ui'
import { z } from 'zod'
import { ContestProfile } from '../../../../../components/ContestProfile'
import { PendingView } from '../../../../../components/StatusViews'
import { contestLogsQuery, contestProfileQuery } from '../../../../../lib/contest-activity'
import { contestsHref } from '../../../../../lib/contests'
import { contestEntryQuery, pageNumber } from '../../../../../lib/leaderboard'
import { viewerQueryOptions } from '../../../../../lib/session'

export const Route = createFileRoute('/contests/$id/profile/$userId/$')({
  loader: async ({ context: { queryClient }, params }) => {
    const input = profileInput(params)
    const [, profile] = await Promise.all([
      queryClient.ensureQueryData(viewerQueryOptions()),
      queryClient.ensureQueryData(contestProfileQuery(input)),
      queryClient.ensureQueryData(contestLogsQuery({ ...input, page: pageNumber(params._splat) })),
      queryClient.ensureQueryData(contestEntryQuery(input.contestId, input.userId)),
    ])
    if (profile.status === 'not-found') throw notFound()
    return profile.status === 'found'
      ? { title: `${profile.registration.user_display_name} · ${profile.registration.contest.title} · Tadoku` }
      : { title: undefined }
  },
  head: ({ loaderData }) => ({ meta: [{ title: loaderData?.title ?? 'Contest profile · Tadoku' }] }),
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
  const input = profileInput(params)
  const profile = useQuery(contestProfileQuery(input))

  if (!profile.data) return <PendingView />
  if (profile.data.status !== 'found') {
    return <EmptyState headingLevel={1} title="Contest profile unavailable" description="Try reloading this page." />
  }
  return <ContestProfile profile={profile.data} page={pageNumber(params._splat)} />
}
