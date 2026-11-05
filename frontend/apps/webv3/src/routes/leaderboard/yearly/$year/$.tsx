import { createFileRoute } from '@tanstack/react-router'
import { GlobalLeaderboard } from '../../../../components/LeaderboardPage'
import { pageNumber, pathNumber, standingsQuery } from '../../../../lib/leaderboard'

/** Like Webv2, a year that isn't a number shows the current year. */
const yearOf = (segment: string) => pathNumber(segment, new Date().getUTCFullYear())

export const Route = createFileRoute('/leaderboard/yearly/$year/$')({
  head: ({ params }) => ({ meta: [{ title: `${yearOf(params.year)} leaderboard · Tadoku` }] }),
  loader: ({ context, params }) =>
    context.queryClient.ensureQueryData(standingsQuery({ kind: 'year', year: yearOf(params.year) }, pageNumber(params._splat))),
  component: function YearlyRoute() {
    const params = Route.useParams()
    return <GlobalLeaderboard scope={{ kind: 'year', year: yearOf(params.year) }} page={pageNumber(params._splat)} />
  },
})
