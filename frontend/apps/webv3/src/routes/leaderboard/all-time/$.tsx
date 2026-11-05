import { createFileRoute } from '@tanstack/react-router'
import { GlobalLeaderboard } from '../../../components/LeaderboardPage'
import { pageNumber, standingsQuery } from '../../../lib/leaderboard'

export const Route = createFileRoute('/leaderboard/all-time/$')({
  head: () => ({ meta: [{ title: 'All time leaderboard · Tadoku' }] }),
  loader: ({ context, params }) =>
    context.queryClient.ensureQueryData(standingsQuery({ kind: 'all-time' }, pageNumber(params._splat))),
  component: function AllTimeRoute() {
    const { _splat: page } = Route.useParams()
    return <GlobalLeaderboard scope={{ kind: 'all-time' }} page={pageNumber(page)} />
  },
})
