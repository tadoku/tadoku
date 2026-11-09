import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'
import { noindex } from '../../../lib/page-meta'

export const Route = createFileRoute('/contests/my-contests/$')({
  head: () => ({ meta: [{ title: 'Contests you organize · Tadoku' }, noindex] }),
  loader: ({ context, params }) => loadContests(context.queryClient, { collection: 'organizing' }, pageNumber(params._splat)),
  component: function MyContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage view={{ collection: 'organizing' }} page={pageNumber(page)} />
  },
})
