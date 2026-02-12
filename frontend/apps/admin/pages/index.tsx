import Head from 'next/head'
import { NextPageWithLayout } from './_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { useUserList } from '@app/common/api'
import {
  useLatestOfficialContest,
  useContestSummary,
  useContestLeaderboard,
  useYearlyLeaderboard,
  LeaderboardEntry,
} from '@app/immersion/api'
import { Loading } from 'ui'

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

const formatScore = (score: number) => {
  return score.toLocaleString()
}

const LeaderboardList = ({ entries }: { entries: LeaderboardEntry[] }) => (
  <div className="mt-4 space-y-2">
    {entries.map(entry => (
      <div
        key={entry.user_id}
        className="flex items-center justify-between text-sm"
      >
        <span className="text-slate-700">
          {entry.rank}. {entry.user_display_name}
        </span>
        <span className="text-slate-500 font-mono">
          {formatScore(entry.score)}
        </span>
      </div>
    ))}
  </div>
)

const CurrentContestCard = () => {
  const contest = useLatestOfficialContest()
  const summary = useContestSummary(contest.data?.id ?? '', {
    enabled: !!contest.data?.id,
  })
  const leaderboard = useContestLeaderboard(
    { contestId: contest.data?.id ?? '', pageSize: 5, page: 1 },
    { enabled: !!contest.data?.id },
  )

  if (contest.isLoading) {
    return (
      <div className="card">
        <Loading />
      </div>
    )
  }

  if (contest.isError || !contest.data) {
    return (
      <div className="card">
        <h2 className="subtitle">Official Contest</h2>
        <p className="mt-4 text-slate-500">No active contest at the moment.</p>
      </div>
    )
  }

  return (
    <div className="card">
      <div className="flex items-start justify-between">
        <h2 className="subtitle">{contest.data.title}</h2>
        <span className="text-sm text-slate-500">
          {formatDate(contest.data.contest_start)} -{' '}
          {formatDate(contest.data.contest_end)}
        </span>
      </div>
      <div className="mt-4">
        {summary.isSuccess && summary.data && (
          <p className="text-sm text-slate-600">
            {summary.data.participant_count} participants &middot;{' '}
            {summary.data.language_count} languages &middot;{' '}
            {formatScore(summary.data.total_score)} pts
          </p>
        )}

        {leaderboard.isSuccess &&
          leaderboard.data.entries.length > 0 && (
            <LeaderboardList entries={leaderboard.data.entries} />
          )}

        {leaderboard.isLoading && (
          <div className="mt-4">
            <Loading />
          </div>
        )}
      </div>
    </div>
  )
}

const YearlyLeaderboardCard = () => {
  const currentYear = new Date().getFullYear()
  const leaderboard = useYearlyLeaderboard({
    year: currentYear,
    pageSize: 5,
    page: 1,
  })

  return (
    <div className="card">
      <h2 className="subtitle">Yearly Leaderboard ({currentYear})</h2>

      {leaderboard.isLoading && (
        <div className="mt-4">
          <Loading />
        </div>
      )}

      {leaderboard.isError && (
        <p className="mt-4 text-slate-500">Unable to load leaderboard.</p>
      )}

      {leaderboard.isSuccess && leaderboard.data.entries.length === 0 && (
        <p className="mt-4 text-slate-500">No data available yet.</p>
      )}

      {leaderboard.isSuccess && leaderboard.data.entries.length > 0 && (
        <LeaderboardList entries={leaderboard.data.entries} />
      )}
    </div>
  )
}

const PlatformStatsCard = () => {
  const users = useUserList({ pageSize: 1, page: 0 })

  return (
    <div className="card">
      <h2 className="subtitle">Total Users</h2>

      <div className="mt-4">
        {users.isLoading ? (
          <Loading />
        ) : users.isSuccess ? (
          <span className="text-3xl font-bold text-primary">
            {formatScore(users.data.total_size)}
          </span>
        ) : (
          <span className="text-slate-500">Unable to load</span>
        )}
      </div>
    </div>
  )
}

const Page: NextPageWithLayout = () => {
  return (
    <>
      <Head>
        <title>Dashboard - Admin - Tadoku</title>
      </Head>
      <h1 className="title">Dashboard</h1>
      <p className="mt-2 text-slate-600">Welcome to the Tadoku Admin Panel.</p>

      <div className="mt-8 v-stack md:h-stack spaced fill">
        <CurrentContestCard />
        <YearlyLeaderboardCard />
        <PlatformStatsCard />
      </div>
    </>
  )
}

Page.getLayout = getDashboardLayout()

export default Page
