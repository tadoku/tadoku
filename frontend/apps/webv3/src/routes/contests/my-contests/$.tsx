import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'

export const Route = createFileRoute('/contests/my-contests/$')({
  head: () => ({ meta: [{ title: 'My contests · Tadoku' }] }),
  loader: ({ context, params }) => loadContests(context.queryClient, 'mine', pageNumber(params._splat)),
  component: function MyContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage collection="mine" page={pageNumber(page)} />
  },
})
