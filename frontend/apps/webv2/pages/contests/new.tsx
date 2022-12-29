import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { ContestForm } from '@app/contests/ContestForm'

interface Props {}

const Contests: NextPage<Props> = () => {
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
      <ContestForm />
    </>
  )
}

export default Contests
