import type { NextPage } from 'next'
import { Breadcrumb, Loading } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { ContestForm } from '@app/immersion/ContestForm'
import { useContestConfigurationOptions } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import Head from 'next/head'

interface Props {}

const Contests: NextPage<Props> = () => {
  const options = useContestConfigurationOptions()

  return (
    <>
      <Head>
        <title>New contests - Tadoku</title>
      </Head>
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
      {options.isLoading ? <Loading /> : null}
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
