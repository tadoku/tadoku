import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditPage } from '@app/content/ContentEditPage'
import { pagesConfig } from '@app/content/pages'

const Page: NextPageWithLayout = () => {
  return <ContentEditPage config={pagesConfig} />
}

Page.getLayout = getDashboardLayout('pages')

export default Page
