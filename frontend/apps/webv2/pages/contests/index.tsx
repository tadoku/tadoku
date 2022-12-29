import type { NextPage } from 'next'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'

interface Props {}

const Contests: NextPage<Props> = () => {
  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Contests', href: '/contests' },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center">
        <h1 className="title">Contests</h1>
        <div className="h-stack justify-end">
          <Link href="/contests/new" className="btn secondary">
            Create contest
          </Link>
        </div>
      </div>
    </>
  )
}

export default Contests
