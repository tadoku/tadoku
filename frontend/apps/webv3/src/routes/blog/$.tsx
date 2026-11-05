import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { BlogList, ContentUnavailable } from '../../components/ContentPages'
import { PendingView } from '../../components/StatusViews'
import { postsQuery } from '../../lib/content'
import { pageNumber } from '../../lib/leaderboard'

export const Route = createFileRoute('/blog/$')({
  loader: ({ context, params }) => context.queryClient.ensureQueryData(postsQuery(pageNumber(params._splat))),
  head: ({ params }) => ({ meta: [{ title: `Blog · page ${pageNumber(params._splat)} · Tadoku` }] }),
  component: function BlogRoute() {
    const page = pageNumber(Route.useParams()._splat)
    const query = useQuery(postsQuery(page))
    const blog = query.data
    if (!blog) return <PendingView />
    return blog.ok ? <BlogList blog={blog.data} page={page} /> : <ContentUnavailable query={query} />
  },
})
