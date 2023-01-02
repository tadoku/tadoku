import { usePostList } from '@app/content/api'
import { PostDetail } from '@app/content/Post'
import { ArrowLongRightIcon } from '@heroicons/react/20/solid'
import { BookOpenIcon } from '@heroicons/react/24/solid'
import { DateTime, Interval } from 'luxon'
import type { NextPage } from 'next'
import Link from 'next/link'

interface Props {}

const data = [
  { rank: '1', user: 'powz', score: 5054.25054 },
  { rank: '2', user: 'Bijak', score: 3605.23605 },
  { rank: '3', user: 'ShockOLatte', score: 2518.72518 },
  { rank: '4', user: 'Ludie', score: 2517.32517 },
  { rank: '5', user: 'Chamsae', score: 2434.42434 },
  { rank: '6', user: 'Salome', score: 2107.12107 },
  { rank: '7', user: 'mmmm', score: 2060.1206 },
  { rank: '8', user: 'Yaku', score: 1667.21667 },
  { rank: '9', user: 'Socks', score: 1635.81635 },
  { rank: '10', user: 'clair', score: 1592.91592 },
]

const scheduledContests = [
  { name: 'Round 1', startMonth: 1, endDay: 31 },
  { name: 'Round 2', startMonth: 3, endDay: 31 },
  { name: 'Round 3', startMonth: 5, endDay: 14 },
  { name: 'Round 4', startMonth: 7, endDay: 31 },
  { name: 'Round 5', startMonth: 9, endDay: 30 },
  { name: 'Round 6', startMonth: 11, endDay: 14 },
]

const Index: NextPage<Props> = () => {
  const posts = usePostList({ pageSize: 1, page: 0 })
  const now = DateTime.utc()
  const year = now.year

  return (
    <>
      <div className="w-full min-h-screen absolute -top-16 left-0 right-0 bg-[url('/img/header.jpg')] bg-no-repeat bg-top z-0"></div>
      <div className="relative h-stack space-x-8 z-10">
        <div className="w-2/6 v-stack space-y-8">
          <div className="card flex flex-col justify-center bg-sky-50">
            <h1 className="title text-xl">Get good at your second language</h1>
            <p>
              Tadoku is a friendly foreign-language immersion contest and
              tracking platform aimed at building a habit of reading and
              listening in your non-native languages.
            </p>
          </div>
          <div className="card">
            <h1 className="title text-xl">Why should I participate?</h1>
            <p>
              Extensive reading and listening of native materials is a great way
              to improve your understanding of the language you&apos;re
              learning. There are many benefits to doing so: it builds
              vocabulary, reinforces grammar patterns, and you learn about the
              culture where your language is spoken.
            </p>
            <p>
              As you track your progress over time you will notice that you can
              understand more and more of the language you're learning.
            </p>
            <Link href="#" className="mt-4 btn primary block !w-full">
              Join Tadoku
              <BookOpenIcon className="ml-2" />
            </Link>
          </div>
        </div>
        <div className="flex-grow">
          <div className="card p-0">
            <div className="h-stack w-full items-center justify-between p-7 pb-4">
              <h2 className="text-xl">{year} Leaderboard Top 10</h2>
              <Link href="#" className="btn">
                See more
                <ArrowLongRightIcon className="ml-2" />
              </Link>
            </div>
            <div className="table-container shadow-transparent w-auto">
              <table className="default shadow-transparent">
                <thead>
                  <tr>
                    <th className="default !pl-7">Rank</th>
                    <th className="default">Nickname</th>
                    <th className="default !text-right !pr-7">Score</th>
                  </tr>
                </thead>
                <tbody>
                  {data.map(u => (
                    <tr key={u.rank} className="link">
                      <td className="link w-10">
                        <Link
                          href={`/profile/${u.user}`}
                          className="reset justify-center"
                        >
                          {u.rank}
                        </Link>
                      </td>
                      <td className="link">
                        <Link
                          href={`/profile/${u.user}`}
                          className="reset text-lg"
                        >
                          {u.user}
                        </Link>
                      </td>
                      <td className="link">
                        <Link
                          href={`/profile/${u.user}`}
                          className="reset justify-end text-lg !pr-7"
                        >
                          {Math.round(u.score * 10) / 10}
                        </Link>
                      </td>
                    </tr>
                  ))}
                  {data.length === 0 ? (
                    <tr>
                      <td
                        colSpan={3}
                        className="default h-32 font-bold text-center text-xl text-slate-400"
                      >
                        No partipants yet, be the first to sign up!
                      </td>
                    </tr>
                  ) : null}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
      <div className="mt-8 flex flex-row w-full space-x-8">
        {posts.data?.posts[0] ? (
          <div className="flex-grow">
            <PostDetail post={posts.data.posts[0]} />
          </div>
        ) : null}
        <div className="w-3/6 card self-start p-0">
          <h2 className="text-xl p-7 pb-4">Contest Schedule for {year}</h2>
          <table className="default w-full shadow-none">
            <thead>
              <tr>
                <th className="default !pl-7">Round</th>
                <th className="default">Start</th>
                <th className="default">End</th>
                <th className="default">Status</th>
              </tr>
            </thead>
            <tbody>
              {scheduledContests.map(c => {
                const start = DateTime.utc(year, c.startMonth, 1)
                const end = DateTime.utc(year, c.startMonth, c.endDay)
                const interval = Interval.fromDateTimes(start, end)
                return (
                  <tr>
                    <td className="default !pl-7">
                      <strong>{c.name}</strong>
                    </td>
                    <td className="default">{start.toFormat('MMM d')}</td>
                    <td className="default">{end.toFormat('MMM d')}</td>
                    <td className="default">
                      {interval.isAfter(now) ? (
                        <span>Scheduled</span>
                      ) : interval.contains(now) ? (
                        <strong className="text-lime-700">Ongoing</strong>
                      ) : (
                        <span className="text-red-700">Ended</span>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </>
  )
}

export default Index
