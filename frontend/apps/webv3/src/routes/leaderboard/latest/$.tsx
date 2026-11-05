import { createFileRoute } from '@tanstack/react-router'
import { LatestLeaderboard } from '../../../components/ContestLeaderboard'
import { latestBoardQuery, pageNumber, registeredQuery, standingsQuery } from '../../../lib/leaderboard'
import { viewerKey, viewerQueryOptions } from '../../../lib/session'

export const Route = createFileRoute('/leaderboard/latest/$')({
  head: () => ({ meta: [{ title: 'Latest leaderboard · Tadoku' }] }),
  loader: async ({ context: { queryClient }, params }) => {
    const [latest, viewer] = await Promise.all([
      queryClient.ensureQueryData(latestBoardQuery()),
      queryClient.ensureQueryData(viewerQueryOptions()),
    ])
    if (latest.kind !== 'board') return
    const contestId = latest.board.contest.id
    await Promise.all([
      queryClient.ensureQueryData(standingsQuery({ kind: 'contest', contestId }, pageNumber(params._splat))),
      viewer.status === 'signed-in' ? queryClient.ensureQueryData(registeredQuery(contestId, viewerKey(viewer))) : null,
    ])
  },
  component: function LatestRoute() {
    const { _splat: page } = Route.useParams()
    return <LatestLeaderboard page={pageNumber(page)} />
  },
})
