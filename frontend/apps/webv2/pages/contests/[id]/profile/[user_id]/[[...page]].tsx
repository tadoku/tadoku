import { useRouter } from 'next/router'
import { Breadcrumb, ButtonGroup, Tabbar } from 'ui'
import { HomeIcon } from '@heroicons/react/20/solid'
import { PencilSquareIcon, PlusIcon } from '@heroicons/react/24/solid'
import { routes } from '@app/common/routes'
import { ContestLeaderboard } from '@app/contests/ContestLeaderboard'
import { ReadingActivityChart } from '@app/contests/ReadingActivityChart'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const userId = router.query['user_id']?.toString() ?? ''
  const contest = {
    data: {
      official: true,
      title: 'dummy',
    },
  }

  return (
    <>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: contest.data.official
                ? 'Official contests'
                : 'User contests',
              href: contest.data.official
                ? routes.contestListOfficial()
                : routes.contestListUserContests(),
            },
            {
              label: contest.data.title,
              href: routes.contestLeaderboard(id),
            },
            {
              label: 'User',
              href: routes.contestUserProfile(id, userId),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">antonve</h1>
          <h2 className="subtitle">{contest.data.title}</h2>
        </div>
        <div></div>
      </div>
      <div className="mt-4 space-x-4 flex w-full">
        <div className="w-1/5 space-y-4 flex-shrink-0">
          <div className="card narrow">
            <h3 className="subtitle mb-2">Overall score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Japanese score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Chinese score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
          <div className="card narrow">
            <h3 className="subtitle mb-2">Korean score</h3>
            <span className="text-4xl font-bold">3,243</span>
          </div>
        </div>
        <div className="flex-grow flex flex-col card narrow">
          <h3 className="subtitle mb-2">Reading activity</h3>
          <div className="flex-1">
            <ReadingActivityChart />
          </div>
        </div>
      </div>
    </>
  )
}

export default Page
