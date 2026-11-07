import { createFileRoute } from '@tanstack/react-router'
import { ContestsPage } from '../../../../components/ContestsPage'
import { loadContests } from '../../../../lib/contests'
import { pageNumber, pathNumber } from '../../../../lib/leaderboard'

/** A year that isn't a number shows the current year, as the yearly leaderboard does. */
const yearOf = (segment: string) => pathNumber(segment, new Date().getUTCFullYear())

export const Route = createFileRoute('/contests/joined/$year/$')({
  head: ({ params }) => ({ meta: [{ title: `Contests you joined in ${yearOf(params.year)} · Tadoku` }] }),
  loader: ({ context, params }) =>
    loadContests(context.queryClient, { collection: 'joined', year: yearOf(params.year) }, pageNumber(params._splat)),
  component: function JoinedContestsRoute() {
    const params = Route.useParams()
    return <ContestsPage view={{ collection: 'joined', year: yearOf(params.year) }} page={pageNumber(params._splat)} />
  },
})
