import { routes } from '@app/common/routes'
import { HomeIcon, PencilSquareIcon } from '@heroicons/react/20/solid'
import { MinusCircleIcon, PlusCircleIcon } from '@heroicons/react/24/solid'
import type { NextPage } from 'next'
import Head from 'next/head'
import { useState } from 'react'
import { Breadcrumb, ButtonGroup } from 'ui'

interface Props {}

const Page: NextPage<Props> = () => {
  const [pages, setPages] = useState(0)

  return (
    <>
      <Head>
        <title>Page counter - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: `Page counter`,
              href: routes.pageCounter(),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Page counter</h1>
        </div>
        <div>
          <ButtonGroup
            actions={[
              {
                href: routes.logCreateWithAmount(pages),
                label: 'Log pages',
                IconComponent: PencilSquareIcon,
                style: 'secondary',
              },
            ]}
            orientation="right"
          />
        </div>
      </div>
      <div className="mt-8 h-stack spaced">
        <button
          onClick={() => setPages(Math.max(0, pages - 1))}
          className="btn ghost !h-32"
        >
          <MinusCircleIcon className="md:!w-28 md:!h-28" />
        </button>
        <input
          type="number"
          value={pages}
          onChange={e => setPages(parseInt(e.target.value, 0))}
          className="text-6xl px-8 py-4 !h-32 text-center"
        />
        <button onClick={() => setPages(pages + 1)} className="btn ghost !h-32">
          <PlusCircleIcon className="md:!w-28 md:!h-28" />
        </button>
      </div>
    </>
  )
}

export default Page
