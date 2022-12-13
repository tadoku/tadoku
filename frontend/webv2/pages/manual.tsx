import type { NextPage } from 'next'
import Breadcrumb from 'tadoku-ui/components/Breadcrumb'
import { HomeIcon } from '@heroicons/react/20/solid'
import Page from '@app/blog/Page'

interface Props {}

const Manual: NextPage<Props> = () => {
  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: '/', IconComponent: HomeIcon },
            { label: 'Manual', href: 'manual' },
          ]}
        />
      </div>
      <div>
        <Page slug="manual" />
      </div>
    </>
  )
}

export default Manual
