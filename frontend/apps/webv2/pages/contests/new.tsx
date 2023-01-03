import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { ContestForm } from '@app/contests/ContestForm'
import { useContestConfigurationOptions } from '@app/contests/api'
import { routes } from '@app/common/routes'

interface Props {}

const Contests: NextPage<Props> = () => {
  const options = useContestConfigurationOptions()

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Contests', href: routes.contestListOfficial() },
            { label: 'Create new contest', href: routes.contestNew() },
          ]}
        />
      </div>
      <h1 className="title mb-4">Create new contest</h1>
      {options.isLoading ? <p>Loading...</p> : null}
      {options.isError ? (
        <span className="flash error">
          Could not load page, please try again later.
        </span>
      ) : null}
      {options.isSuccess ? (
        <ContestForm configurationOptions={options.data} />
      ) : null}
    </>
  )
}

export default Contests
