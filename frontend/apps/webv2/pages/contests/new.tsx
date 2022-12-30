import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { ContestForm } from '@app/contests/ContestForm'
import { useContestConfigurationOptions } from '@app/contests/api'

interface Props {}

const Contests: NextPage<Props> = () => {
  const options = useContestConfigurationOptions()

  if (options.isLoading || options.isIdle) {
    return <p>Loading...</p>
  }

  if (options.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Contests', href: '/contests' },
            { label: 'Create new contest', href: '/contests/new' },
          ]}
        />
      </div>
      <h1 className="title mb-4">Create new contest</h1>
      <ContestForm configurationOptions={options.data} />
    </>
  )
}

export default Contests
