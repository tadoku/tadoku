import { useRouter } from 'next/router'
import { Breadcrumb, ActionMenu } from 'ui'
import { ChevronRightIcon, HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { ReadingActivityChart } from '@app/contests/ReadingActivityChart'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useContestProfileScores } from '@app/contests/api'
import { formatScore } from '@app/common/format'

function truncate(text: string | undefined, len: number) {
  if (text === undefined) {
    return text
  }

  if (text.length > len - 3) {
    return text.substring(0, len - 3) + '...'
  }

  return text
}

const Page = () => {
  const router = useRouter()
  const contestId = router.query['id']?.toString() ?? ''
  const userId = router.query['user_id']?.toString() ?? ''

  const profile = useContestProfileScores({ userId, contestId })

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
      id: 'abc',
      date: DateTime.fromISO(date),
      language: langs[Math.floor(Math.random() * langs.length)],
      activity: activities[Math.floor(Math.random() * activities.length)],
      description:
        descriptions[Math.floor(Math.random() * descriptions.length)],
      amount: Math.floor(Math.random() * 100000),
      score: Math.floor(Math.random() * 100000),
      unit: 'page',
    }))

  if (profile.isLoading || profile.isIdle) {
    return <p>Loading...</p>
  }

  const contest = profile.data?.registration.contest

  if (profile.isError || !contest) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const scores = new Map<string, number>()
  for (const { language_code, score } of profile.data.scores) {
    scores.set(language_code, score)
  }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.official ? 'Official contests' : 'User contests',
              href: contest.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.title,
              href: routes.contestLeaderboard(contestId),
            },
            {
              label: profile.data.registration.user_display_name,
              href: routes.contestUserProfile(contestId, userId),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">antonve</h1>
          <h2 className="subtitle">{contest.title}</h2>
        </div>
        <div></div>
      </div>
      <div className="my-4 lg:space-x-4 flex flex-col lg:flex-row w-full">
        <div className="lg:w-1/5 grid gap-4">
          <div className="card narrow">
            <h3 className="subtitle mb-2">Overall score</h3>
            <span className="text-4xl font-bold">
              {formatScore(profile.data.overall_score)}
            </span>
          </div>
          {profile.data.registration.languages.map(({ code, name }) => (
            <div className="card narrow" key={code}>
              <h3 className="subtitle mb-2">{name}</h3>
              <span className="text-4xl font-bold">
                {scores.get(code) ?? 0}
              </span>
            </div>
          ))}
        </div>
        <div className="mt-4 lg:mt-0 flex-grow flex flex-col card narrow">
          <h3 className="subtitle mb-2">Reading activity</h3>
          <div className="flex-1 max-h-72 lg:max-h-[28rem]">
            <ReadingActivityChart
              userId={userId}
              registration={profile.data.registration}
            />
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
                <th className="default w-28 hidden md:table-cell">Activity</th>
                <th className="default hidden lg:table-cell">Description</th>
                <th className="default w-36 hidden md:table-cell">Amount</th>
                <th className="default w-24 !text-right">Score</th>
                <th className="default"></th>
              </tr>
            </thead>
            <tbody>
              {data.map(it => (
                <tr key={it.date.toString()} className="link">
                  <td className="default link">
                    <Link className="reset" href={routes.log(it.id)}>
                      {it.date.toLocaleString(DateTime.DATE_MED)}
                    </Link>
                  </td>
                  <td className="default link">
                    <Link className="reset" href={routes.log(it.id)}>
                      {it.language}
                    </Link>
                  </td>
                  <td className="default link hidden md:table-cell">
                    <Link className="reset" href={routes.log(it.id)}>
                      {it.activity}
                    </Link>
                  </td>
                  <td
                    className={`default text-sm link hidden lg:table-cell ${
                      !it.description ? 'opacity-50' : ''
                    }`}
                  >
                    <Link className="reset" href={routes.log(it.id)}>
                      {truncate(it.description, 38) ?? 'N/A'}
                    </Link>
                  </td>
                  <td className="default link font-bold hidden md:table-cell">
                    <Link className="reset" href={routes.log(it.id)}>
                      {it.amount} {it.unit}
                      {it.amount !== 1 ? 's' : ''}
                    </Link>
                  </td>
                  <td className="default link font-bold">
                    <Link
                      className="reset justify-end"
                      href={routes.log(it.id)}
                    >
                      {it.score}
                    </Link>
                  </td>
                  <td className="default link w-12">
                    <Link
                      className="reset flex-shrink"
                      href={routes.log(it.id)}
                    >
                      <ChevronRightIcon className="w-5 h-5" />
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
      pagination
    </>
  )
}

export default Page
