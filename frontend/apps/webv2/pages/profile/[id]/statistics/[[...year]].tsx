import { useRouter } from 'next/router'
import { Breadcrumb } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'

const Page = () => {
  const router = useRouter()
  const contestId = router.query['id']?.toString() ?? ''
  const userId = router.query['user_id']?.toString() ?? ''

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
    </>
  )
}

export default Page
