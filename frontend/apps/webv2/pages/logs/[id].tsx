import { routes } from '@app/common/routes'
import { useLog } from '@app/logs/api'
import { HomeIcon } from '@heroicons/react/20/solid'
import { useRouter } from 'next/router'
import { Breadcrumb } from 'ui'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const log = useLog(id)

  if (log.isLoading || log.isIdle) {
    return <p>Loading...</p>
  }

  if (log.isError) {
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
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            { label: 'Logs', href: '#' },
          ]}
        />
      </div>
    </>
  )
}

export default Page
