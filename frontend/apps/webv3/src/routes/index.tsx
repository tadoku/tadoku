import { createFileRoute } from '@tanstack/react-router'
import { HomePage } from '../components/HomePage'
import { homeBoardQuery, lastYearStatisticsQuery, memberDashboardQueries, nextContestQuery } from '../lib/home'
import { utcToday } from '../lib/format'
import { readsLastYear } from '../lib/recent-activity'
import { contestEntryQuery, latestBoardQuery } from '../lib/leaderboard'
import { registrationQuery } from '../lib/registration'
import { viewerKey, viewerQueryOptions } from '../lib/session'

export const Route = createFileRoute('/')({
  head: () => ({
    meta: [
      { title: 'Tadoku' },
      {
        name: 'description',
        content: 'Join friendly immersion contests, log what you read and listen to, and follow the community’s progress in the languages you’re learning.',
      },
    ],
  }),
  loader: async ({ context: { queryClient } }) => {
    const year = new Date().getUTCFullYear()
    const viewerRead = queryClient.ensureQueryData(viewerQueryOptions())
    const contestRead = queryClient.ensureQueryData(latestBoardQuery())
      .then(latest => (latest.kind === 'board' ? latest.board.contest.id : null))

    await Promise.all([
      queryClient.prefetchQuery(nextContestQuery()),
      contestRead.then(contestId => (contestId ? queryClient.prefetchQuery(homeBoardQuery(contestId)) : null)),
      viewerRead.then(viewer => {
        if (viewer.status !== 'signed-in') return null
        const [statistics, logs] = memberDashboardQueries(viewer, year)
        return Promise.all([
          queryClient.prefetchQuery(statistics),
          queryClient.prefetchQuery(logs),
          readsLastYear(utcToday()) ? queryClient.prefetchQuery(lastYearStatisticsQuery(viewer, year)) : null,
        ])
      }),
      // The board offers Join only to members who haven't registered, and marks their row, so both render on the server.
      Promise.all([viewerRead, contestRead]).then(([viewer, contestId]) => (viewer.status === 'signed-in' && contestId
        ? Promise.all([
          queryClient.prefetchQuery(registrationQuery(contestId, viewerKey(viewer))),
          queryClient.prefetchQuery(contestEntryQuery(contestId, viewer.userId)),
        ])
        : null)),
    ])
    return { year }
  },
  component: HomePage,
})
