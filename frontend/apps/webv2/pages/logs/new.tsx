import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { useLogConfigurationOptions } from '@app/logs/api'
import { useOngoingContestRegistrations } from '@app/contests/api'
import { routes } from '@app/common/routes'
import { LogForm } from '@app/logs/NewLogForm/Form'

interface Props {}

const Page: NextPage<Props> = () => {
  const registrations = useOngoingContestRegistrations()
  const options = useLogConfigurationOptions()
  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'New log', href: routes.logCreate() },
          ]}
        />
      </div>
      <h1 className="title mb-4">New log</h1>
      {options.isLoading || registrations.isLoading ? <p>Loading...</p> : null}
      {options.isError || registrations.isError ? (
        <span className="flash error">
          Could not load page, please try again later.
        </span>
      ) : null}
      {options.isSuccess && registrations.isSuccess ? (
        <LogForm registrations={registrations.data} options={options.data} />
      ) : null}
    </>
  )
}

export default Page
