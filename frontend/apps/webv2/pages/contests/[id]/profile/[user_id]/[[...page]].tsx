import { useRouter } from 'next/router'
import { Breadcrumb, ActionMenu } from 'ui'
import {
  HomeIcon,
  PencilIcon,
  PencilSquareIcon,
  TrashIcon,
} from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ReadingActivityChart } from '@app/contests/ReadingActivityChart'
import { DateTime } from 'luxon'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const userId = router.query['user_id']?.toString() ?? ''
  const contest = {
    data: {
      official: true,
      title: 'dummy',
    },
  }

  const activities = ['Reading', 'Listening', 'Writing', 'Speaking', 'Study']
  const langs = ['Chinese (Mandarin)', 'Japanese', 'Korean']
  const descriptions = [
    undefined,
    'One piece',
    '乙女ゲームの破滅フラグしかない悪役令嬢に転生してしまった…２',
    '今夜、世界からこの涙が消えても',
  ]

  const data = Array.from(Array(14).keys())
    .reverse()
    .map(day => '2022-12-' + (day + 1).toString().padStart(2, '0'))
    .map(date => ({
      date: DateTime.fromISO(date),
      language: langs[Math.floor(Math.random() * langs.length)],
      activity: activities[Math.floor(Math.random() * activities.length)],
      description:
        descriptions[Math.floor(Math.random() * descriptions.length)],
      amount: Math.floor(Math.random() * 100),
      score: Math.floor(Math.random() * 100),
      unit: 'page',
    }))

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.data.official
                ? 'Official contests'
                : 'User contests',
              href: contest.data.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.data.title,
              href: routes.contestLeaderboard(id),
            },
            {
              label: 'User',
              href: routes.contestUserProfile(id, userId),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">antonve</h1>
          <h2 className="subtitle">{contest.data.title}</h2>
        </div>
        <div></div>
      </div>
      <div className="my-4 space-x-4 flex w-full">
        <div className="w-1/5 space-y-4 flex-shrink-0">
          <div className="card narrow">
            <h3 className="subtitle mb-2">Overall score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Japanese score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Chinese (Mandarin) score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Korean score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
        </div>
        <div className="flex-grow flex flex-col card narrow">
          <h3 className="subtitle mb-2">Reading activity</h3>
          <div className="flex-1">
            <ReadingActivityChart />
          </div>
        </div>
      </div>

      <div className="card p-0">
        <div className="h-stack w-full items-center justify-between">
          <h2 className="subtitle p-4">Updates</h2>
        </div>
        <div className="table-container shadow-transparent w-auto">
          <table className="default shadow-transparent">
            <thead>
              <tr>
                <th className="default w-36">Date</th>
                <th className="default w-32">Language</th>
                <th className="default w-28">Activity</th>
                <th className="default">Description</th>
                <th className="default w-36">Amount</th>
                <th className="default w-36 !text-right">Score</th>
                <th className="default !text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {data.map(it => (
                <tr key={it.date.toString()}>
                  <td className="default">
                    {it.date.toLocaleString(DateTime.DATE_MED)}
                  </td>
                  <td className="default">{it.language}</td>
                  <td className="default">{it.activity}</td>
                  <td
                    className={`default text-sm ${
                      !it.description ? 'opacity-50' : ''
                    }`}
                  >
                    {it.description ?? 'N/A'}
                  </td>
                  <td className="default font-bold">
                    {it.amount} {it.unit}
                    {it.amount !== 1 ? 's' : ''}
                  </td>
                  <td className="text-right default font-bold">{it.score}</td>
                  <td className="text-right default">
                    <div className="flex justify-end items-center">
                      <a href="#" className="btn ghost small">
                        <PencilSquareIcon />
                      </a>
                      <a
                        href="#"
                        className="btn ghost text-red-700 hover:!text-red-700/80 small"
                      >
                        <TrashIcon />
                      </a>
                    </div>
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
    </>
  )
}

export default Page
