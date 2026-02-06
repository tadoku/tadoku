import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditPage } from '@app/content/ContentEditPage'
import { postsConfig } from '@app/content/posts'

const Page: NextPageWithLayout = () => {
  return <ContentEditPage config={postsConfig} />
}

Page.getLayout = getDashboardLayout('posts')

export default Page
