import { colorForActivity, formatUnit } from '@app/common/format'
import { routes } from '@app/common/routes'
import { useLog } from '@app/logs/api'
import { HomeIcon } from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { CalculatorIcon } from '@heroicons/react/24/solid'
import { DateTime } from 'luxon'
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
      <h1 className="title">
        Log by {log.data.user_display_name ?? 'anonymous'}
      </h1>
      <h2 className="subtitle">
        {DateTime.fromISO(log.data.created_at).toLocaleString(
          DateTime.DATETIME_MED,
        )}
      </h2>
      <div className="w-fit">
        <div className="card mt-4 w-full relative">
          <div
            className={`bg-${logColor}-300 absolute top-0 left-0 right-0 h-2`}
          ></div>
          <div className="h-stack spaced">
            <span
              className={`py-1 px-3 text-xs items-center flex bg-${logColor}-300 text-${logColor}-900`}
            >
              {log.data.activity.name}
            </span>
            <span className="py-1 px-3 text-xs items-center flex text-white bg-slate-500">
              {log.data.language.name}
            </span>
            {log.data.tags.map(it => (
              <span
                className={`py-1 px-3 text-xs items-center flex text-white bg-secondary`}
              >
                {it}
              </span>
            ))}
          </div>
          {log.data.description ? (
            <p className="">{log.data.description}</p>
          ) : null}
        </div>
        <div className="grid card mt-4 w-full ">
          <h3 className="subtitle mb-2">Score calculation</h3>
          <div className="grid grid-cols-3 gap-x-4 gap-y-2 w-full">
            <div className="col-span-2 text-right font-bold text-3xl">
              {log.data.amount}
            </div>
            <div className="flex items-center lowercase font-bold px-4 bg-slate-500 text-white">
              {formatUnit(log.data.amount, log.data.unit_name)}
            </div>
            <div className="flex items-center">
              <XMarkIcon className="w-6 h-6" />
            </div>
            <div className="text-right font-bold text-3xl">
              {log.data.modifier}
            </div>
            <div className="flex items-center lowercase font-bold px-4 bg-slate-500 text-white">
              Modifier
            </div>
            <div className="col-span-2 text-right border-t-2 border-slate-500/20 font-bold text-3xl mt-2 pt-2">
              {log.data.score}
            </div>
            <div className="flex items-center lowercase font-bold px-4 bg-secondary text-white mt-4">
              Score
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default Page
