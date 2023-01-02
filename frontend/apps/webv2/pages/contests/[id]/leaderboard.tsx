import { useContest } from '@app/contests/api'
import { useRouter } from 'next/router'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''

  const contest = useContest(id)

  if (contest.isLoading || contest.isIdle) {
    return <p>Loading...</p>
  }

  if (contest.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  return (
    <>
      <h1 className="title">{contest.data.description}</h1>
    </>
  )
}

export default Page
