import { colorForActivity, formatArray, formatUnit } from '@app/common/format'
import { routes } from '@app/common/routes'
import { useLog } from '@app/logs/api'
import { HomeIcon } from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { CalculatorIcon } from '@heroicons/react/24/solid'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Breadcrumb, chartColors } from 'ui'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const log = useLog(id)

  if (log.isLoading || log.isIdle) {
    return <p>Loading...</p>
  }

  if (log.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  const logColor = colorForActivity(log.data.activity.id)

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Logs', href: '#' },
          ]}
        />
      </div>
      <h1 className="title">Log details</h1>
      <h2 className="subtitle">
        By {log.data.user_display_name ?? 'anonymous'},{' '}
        {DateTime.fromISO(log.data.created_at).toLocaleString(
          DateTime.DATETIME_MED,
        )}
      </h2>
      <div className="h-stack space-x-4 mt-4">
        <div className="w-fit">
          <div className="card w-full relative">
            <div
              className={`bg-${logColor}-300 absolute top-0 left-0 right-0 h-2`}
            ></div>
            <div className="h-stack spaced">
              <span
                className={`py-1 px-3 text-xs items-center flex bg-${logColor}-300 text-${logColor}-900`}
              >
                {log.data.activity.name}
              </span>
              <span className="py-1 px-3 text-xs items-center flex text-slate-900 bg-slate-200">
                {log.data.language.name}
              </span>
              {log.data.tags.map(it => (
                <span
                  className={`py-1 px-3 text-xs items-center flex text-slate-900 bg-slate-200`}
                >
                  {it}
                </span>
              ))}
            </div>
            {log.data.registrations && log.data.registrations.length >= 1 ? (
              <p>
                Submitted to{' '}
                {formatArray(log.data.registrations, it => (
                  <Link
                    href={routes.contestLeaderboard(it.contest_id)}
                    className="font-bold"
                  >
                    {it.title}
                  </Link>
                ))}
              </p>
            ) : null}
            {log.data.description ? (
              <>
                <h3 className="subtitle my-2">Description</h3>
                <p className="">{log.data.description}</p>
              </>
            ) : null}
            <div className="h-stack w-full mt-4">
              <div className="w-1/2">
                <h3 className="subtitle mb-2">Score</h3>
                <div className="font-bold text-5xl">{log.data.score}</div>
              </div>
              <div className="w-1/2 flex flex-col items-end justify-end opacity-80">
                <h4 className="subtitle text-sm">Breakdown</h4>
                <div className="lowercase flex items-center space-x-1 text-sm">
                  <strong className="text-lg">{log.data.amount}</strong>
                  <span className="text-slate-500">
                    {formatUnit(log.data.amount, log.data.unit_name)}
                  </span>
                  <XMarkIcon className="w-3 h-3 mx-2 text-secondary" />
                  <strong className="text-lg">{log.data.modifier}</strong>
                  <span className="text-slate-500">modifier</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default Page
