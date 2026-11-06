import { useQuery } from '@tanstack/react-query'
import { PageHeader, Tabbar } from 'paper-ui'
import type { ReactNode } from 'react'
import { firstYear, leaderboardHref, standingsQuery, type StandingsScope } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { renderLink } from './router-link'
import { Standings } from './Standings'

/** The page title and period tabs every leaderboard shares, around the board for one period. */
export function LeaderboardFrame({ period, currentYear, actions, children }: {
  period: 'latest' | 'yearly' | 'all-time'
  currentYear: number
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="standings-page">
      <PageHeader className="standings-header" title="Leaderboard" actions={actions} />
      <Tabbar
        label="Leaderboard period"
        renderLink={renderLink}
        links={[
          { id: 'latest', label: 'Latest official', href: leaderboardHref.latest(), current: period === 'latest' },
          { id: 'yearly', label: 'Yearly', href: leaderboardHref.yearly(currentYear), current: period === 'yearly' },
          { id: 'all-time', label: 'All time', href: leaderboardHref.allTime(), current: period === 'all-time' },
        ]}
      />
      {children}
    </div>
  )
}

/** The official standings for one year or all time, with the year picker for yearly boards. */
export function GlobalLeaderboard({ scope, page }: { scope: Exclude<StandingsScope, { kind: 'contest' }>; page: number }) {
  const links = useLinks()
  const standings = useQuery(standingsQuery(scope, page))
  // The route loader fetched the standings, and the fetch time travels with the dehydrated data, so server and client
  // list the same years.
  const currentYear = new Date(standings.dataUpdatedAt).getUTCFullYear()
  const yearly = scope.kind === 'year'
  const period = yearly ? `in ${scope.year}` : 'of all time'
  const years = Array.from({ length: currentYear - firstYear + 1 }, (_, index) => currentYear - index)

  return (
    <LeaderboardFrame period={yearly ? 'yearly' : 'all-time'} currentYear={currentYear}>
      <div className="standings-workspace">
        <div className="standings-main">
          <section className="standings-sheet" aria-label={yearly ? `${scope.year} official standings` : 'All-time official standings'}>
            {yearly ? (
              <Tabbar
                label="Year"
                renderLink={renderLink}
                links={years.map(year => ({
                  id: String(year),
                  label: String(year),
                  href: leaderboardHref.yearly(year),
                  current: year === scope.year,
                }))}
              />
            ) : null}
            <Standings
              result={standings.data}
              page={page}
              caption={`Official contest standings ${period}`}
              pageHref={next => (yearly ? leaderboardHref.yearly(scope.year, next) : leaderboardHref.allTime(next))}
              profileHref={links.profile}
              empty={{
                title: 'No official scores for this period',
                description: 'Scores will appear here when eligible activity is logged in an official contest.',
              }}
              retrying={standings.isFetching}
              onRetry={() => void standings.refetch()}
            />
          </section>
        </div>
        <aside className="standings-aside">
          <p>Only official contests count; community contest scores are not included.</p>
        </aside>
      </div>
    </LeaderboardFrame>
  )
}
