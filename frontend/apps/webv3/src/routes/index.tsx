import { createFileRoute } from '@tanstack/react-router'
import { HomePage } from '../components/HomePage'
import { homeQueryOptions, memberDashboardQueries } from '../lib/home'
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
  loader: async ({ context }) => {
    const viewer = await context.queryClient.ensureQueryData(viewerQueryOptions())
    const year = new Date().getUTCFullYear()
    const [statistics, activity] = viewer.status === 'signed-in'
      ? memberDashboardQueries(viewer, year)
      : [null, null]
    await Promise.all([
      context.queryClient.prefetchQuery(homeQueryOptions()),
      statistics ? context.queryClient.prefetchQuery(statistics) : null,
      activity ? context.queryClient.prefetchQuery(activity) : null,
    ])
    // The hero offers Join only to members who haven't registered, so their registration renders on the server too.
    const contest = context.queryClient.getQueryData(homeQueryOptions().queryKey)?.latestContest
    if (viewer.status === 'signed-in' && contest) {
      await context.queryClient.prefetchQuery(registrationQuery(contest.id, viewerKey(viewer)))
    }
    return { year }
  },
  component: HomePage,
})
