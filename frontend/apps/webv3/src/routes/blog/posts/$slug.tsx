import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound } from '@tanstack/react-router'
import { BlogPostArticle, ContentUnavailable } from '../../../components/ContentPages'
import { PendingView } from '../../../components/StatusViews'
import { postQuery } from '../../../lib/content'

export const Route = createFileRoute('/blog/posts/$slug')({
  loader: async ({ context, params }) => {
    const post = await context.queryClient.ensureQueryData(postQuery(params.slug))
    if (post.kind === 'not-found') throw notFound()
    return { title: post.kind === 'found' ? post.data.title : undefined }
  },
  head: ({ loaderData }) => ({ meta: [{ title: loaderData?.title ? `${loaderData.title} · Blog · Tadoku` : 'Blog · Tadoku' }] }),
  component: function BlogPostRoute() {
    const query = useQuery(postQuery(Route.useParams().slug))
    const post = query.data
    if (!post) return <PendingView />
    return post.kind === 'found' ? <BlogPostArticle post={post.data} /> : <ContentUnavailable query={query} />
  },
})
