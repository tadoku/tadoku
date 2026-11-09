import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'
import { noindex } from '../../../lib/page-meta'

export const Route = createFileRoute('/contests/joined/$')({
  head: () => ({ meta: [{ title: 'Contests you joined · Tadoku' }, noindex] }),
  loader: ({ context, params }) => loadContests(context.queryClient, { collection: 'joined' }, pageNumber(params._splat)),
  component: function JoinedContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage view={{ collection: 'joined' }} page={pageNumber(page)} />
  },
})
