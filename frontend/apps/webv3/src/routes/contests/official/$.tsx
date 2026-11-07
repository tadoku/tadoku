import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'

export const Route = createFileRoute('/contests/official/$')({
  head: () => ({ meta: [{ title: 'Official contests · Tadoku' }] }),
  loader: ({ context, params }) => loadContests(context.queryClient, { collection: 'official' }, pageNumber(params._splat)),
  component: function OfficialContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage view={{ collection: 'official' }} page={pageNumber(page)} />
  },
})
