import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound } from '@tanstack/react-router'
import { CmsPageArticle, ContentUnavailable } from '../../components/ContentPages'
import { PendingView } from '../../components/StatusViews'
import { pageQuery } from '../../lib/content'
import { pageUnavailable } from '../../lib/page-status'
import { htmlDescription, sharingMeta } from '../../lib/page-meta'

export const Route = createFileRoute('/pages/$slug')({
  loader: async ({ context, params }) => {
    const page = await context.queryClient.ensureQueryData(pageQuery(params.slug))
    if (page.kind === 'not-found') throw notFound()
    if (page.kind === 'unavailable') pageUnavailable(context.queryClient, pageQuery(params.slug))
    return { title: page.data.title, description: htmlDescription(page.data.html, `${page.data.title} on Tadoku.`) }
  },
  head: ({ loaderData }) => ({ meta: loaderData ? sharingMeta(loaderData.title, loaderData.description) : [{ title: 'Tadoku' }] }),
  component: function CmsPageRoute() {
    const query = useQuery(pageQuery(Route.useParams().slug))
    const page = query.data
    if (!page) return <PendingView />
    return page.kind === 'found' ? <CmsPageArticle page={page.data} /> : <ContentUnavailable query={query} />
  },
})
