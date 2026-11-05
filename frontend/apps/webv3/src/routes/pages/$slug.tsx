import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound } from '@tanstack/react-router'
import { CmsPageArticle, ContentUnavailable } from '../../components/ContentPages'
import { PendingView } from '../../components/StatusViews'
import { pageQuery } from '../../lib/content'

export const Route = createFileRoute('/pages/$slug')({
  loader: async ({ context, params }) => {
    const page = await context.queryClient.ensureQueryData(pageQuery(params.slug))
    if (page.kind === 'not-found') throw notFound()
    return { title: page.kind === 'found' ? page.data.title : undefined }
  },
  head: ({ loaderData }) => ({ meta: [{ title: loaderData?.title ? `${loaderData.title} · Tadoku` : 'Tadoku' }] }),
  component: function CmsPageRoute() {
    const query = useQuery(pageQuery(Route.useParams().slug))
    const page = query.data
    if (!page) return <PendingView />
    return page.kind === 'found' ? <CmsPageArticle page={page.data} /> : <ContentUnavailable query={query} />
  },
})
