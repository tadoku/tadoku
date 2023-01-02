import { ArrowLongRightIcon } from '@heroicons/react/20/solid'
import { BookOpenIcon } from '@heroicons/react/24/solid'
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

const Index: NextPage<Props> = () => {
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
              <h2 className="text-xl">2022 Leaderboard Top 10</h2>
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
    </>
  )
}

export default Index
