import { useRouter } from 'next/router'
import { Breadcrumb, Tabbar, VerticalTabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import Link from 'next/link'

const Page = () => {
  const router = useRouter()
  const userId = router.query['id']?.toString() ?? ''

  // if (profile.isError || !contest) {
  //   return (
  //     <span className="flash error">
  //       Could not load page, please try again later.
  //     </span>
  //   )
  // }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Profile</h1>
          <h2 className="subtitle">antonve</h2>
        </div>
        <div></div>
      </div>
      <Tabbar
        links={[
          {
            href: routes.userProfileStatistics(userId),
            label: 'Statistics',
            active: true,
          },
          {
            href: routes.userProfileUpdates(userId),
            label: 'Updates',
            active: false,
            disabled: true,
          },
        ]}
      />
      <div className="h-stack mt-4 space-x-8">
        <div className="flex-grow v-stack spaced">
          <div className="card narrow">
            <h3 className="subtitle">525 updates in 2023</h3>
            <div className="bg-cyan-400 w-full h-28 mt-4"></div>
          </div>
          <div className="h-stack spaced flex-grow">
            <div className="card w-full p-0">
              <h3 className="subtitle p-4">Contests</h3>
              <ul className="divide-y-2 divide-slate-500/5 border-t-2 border-slate-500/5">
                {[
                  'Contest 1',
                  'Contest 2',
                  'Contest 3',
                  'Contest 4',
                  'Contest 5',
                  'Contest 6',
                  'Contest 7',
                ].map(u => (
                  <li key={`${u[0]}-${u[1]}`}>
                    <Link
                      href="#"
                      className="reset px-4 py-2 flex justify-between items-center hover:bg-slate-500/5"
                    >
                      <span className="font-bold text-base">{u}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
            <div className="card narrow w-full">
              <h3 className="subtitle">Activities</h3>
              <div className="bg-lime-400 w-full h-64 mt-4"></div>
            </div>
          </div>
        </div>
        <div className="flex-shrink">
          <VerticalTabbar
            links={[
              {
                href: routes.userProfileStatistics(userId, 2023),
                label: '2023',
                active: true,
              },
              {
                href: routes.userProfileStatistics(userId, 2022),
                label: '2022',
                active: false,
              },
              {
                href: routes.userProfileStatistics(userId, 2021),
                label: '2021',
                active: false,
              },
              {
                href: routes.userProfileStatistics(userId, 2020),
                label: '2020',
                active: false,
              },
            ]}
          />
        </div>
      </div>
    </>
  )
}

export default Page
