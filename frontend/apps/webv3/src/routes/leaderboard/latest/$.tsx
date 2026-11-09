import { createFileRoute } from '@tanstack/react-router'
import { LatestLeaderboard } from '../../../components/ContestLeaderboard'
import { latestBoardQuery, pageNumber, standingsQuery } from '../../../lib/leaderboard'
import { registrationQuery } from '../../../lib/registration'
import { viewerKey, viewerQueryOptions } from '../../../lib/session'

export const Route = createFileRoute('/leaderboard/latest/$')({
  head: () => ({ meta: [{ title: 'Latest leaderboard · Tadoku' }] }),
  loader: async ({ context: { queryClient }, params }) => {
    const viewerRead = queryClient.ensureQueryData(viewerQueryOptions())
    const contestRead = queryClient.ensureQueryData(latestBoardQuery())
      .then(latest => (latest.kind === 'board' ? latest.board.contest.id : null))
    await Promise.all([
      contestRead.then(contestId => (contestId
        ? queryClient.ensureQueryData(standingsQuery({ kind: 'contest', contestId }, pageNumber(params._splat)))
        : null)),
      Promise.all([viewerRead, contestRead]).then(([viewer, contestId]) => (viewer.status === 'signed-in' && contestId
        ? queryClient.ensureQueryData(registrationQuery(contestId, viewerKey(viewer)))
        : null)),
    ])
  },
  component: function LatestRoute() {
    const { _splat: page } = Route.useParams()
    return <LatestLeaderboard page={pageNumber(page)} />
  },
})
