import { createFileRoute } from '@tanstack/react-router'
import { HomePage } from '../components/HomePage'
import { homeQueryOptions, memberDashboardQueries } from '../lib/home'
import { viewerQueryOptions } from '../lib/session'

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
    const [statistics, activity] = viewer.status === 'signed-in'
      ? memberDashboardQueries(viewer, new Date().getUTCFullYear())
      : [null, null]
    await Promise.all([
      context.queryClient.prefetchQuery(homeQueryOptions()),
      statistics ? context.queryClient.prefetchQuery(statistics) : null,
      activity ? context.queryClient.prefetchQuery(activity) : null,
    ])
  },
  component: HomePage,
})
