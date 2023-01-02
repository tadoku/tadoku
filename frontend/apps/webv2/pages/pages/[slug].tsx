import type { NextPage } from 'next'
import { Page } from '@app/content/Page'
import { useRouter } from 'next/router'

interface Props {}

const Manual: NextPage<Props> = () => {
  const router = useRouter()
  const slug = router.query.slug as string

  return <Page slug={slug} />
}

export default Manual
