import { useRouter } from 'next/router'
import { Breadcrumb, Tabbar, VerticalTabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'

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
      <div className="h-stack mt-4">
        <div className="w-[100px]">
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
