import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'

export const Route = createFileRoute('/contests/user-contests/$')({
  head: () => ({ meta: [{ title: 'Community contests · Tadoku' }] }),
  loader: ({ context, params }) => loadContests(context.queryClient, 'community', pageNumber(params._splat)),
  component: function CommunityContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage collection="community" page={pageNumber(page)} />
  },
})
