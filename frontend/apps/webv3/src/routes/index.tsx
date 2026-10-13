import { createFileRoute } from '@tanstack/react-router'
import { HomePage } from '../components/HomePage'
import { homeQueryOptions } from '../lib/home'

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
  loader: ({ context }) => context.queryClient.prefetchQuery(homeQueryOptions()),
  component: HomePage,
})
