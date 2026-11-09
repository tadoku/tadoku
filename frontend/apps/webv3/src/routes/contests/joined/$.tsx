import { createFileRoute, redirect } from '@tanstack/react-router'
import { ContestsPage } from '../../../components/ContestsPage'
import { contestsHref, loadContests } from '../../../lib/contests'
import { pageNumber } from '../../../lib/leaderboard'

export const Route = createFileRoute('/contests/joined/$')({
  // The list used to go year by year, at /contests/joined/{year}/{page}; those addresses land on its first page.
  beforeLoad: ({ params }) => {
    if (params._splat?.includes('/')) throw redirect({ href: contestsHref.joined(), statusCode: 308 })
  },
  head: () => ({ meta: [{ title: 'Contests you joined · Tadoku' }] }),
  loader: ({ context, params }) => loadContests(context.queryClient, { collection: 'joined' }, pageNumber(params._splat)),
  component: function JoinedContestsRoute() {
    const { _splat: page } = Route.useParams()
    return <ContestsPage view={{ collection: 'joined' }} page={pageNumber(page)} />
  },
})
