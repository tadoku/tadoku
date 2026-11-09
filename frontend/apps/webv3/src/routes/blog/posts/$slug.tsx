import { useQuery } from '@tanstack/react-query'
import { createFileRoute, notFound } from '@tanstack/react-router'
import { BlogPostArticle, ContentUnavailable } from '../../../components/ContentPages'
import { PendingView } from '../../../components/StatusViews'
import { postQuery } from '../../../lib/content'
import { pageUnavailable } from '../../../lib/page-status'
import { markdownDescription, sharingMeta } from '../../../lib/page-meta'

export const Route = createFileRoute('/blog/posts/$slug')({
  loader: async ({ context, params }) => {
    const post = await context.queryClient.ensureQueryData(postQuery(params.slug))
    if (post.kind === 'not-found') throw notFound()
    if (post.kind === 'unavailable') pageUnavailable(context.queryClient, postQuery(params.slug))
    return { title: post.data.title, description: markdownDescription(post.data.content, `${post.data.title}, on the Tadoku blog.`) }
  },
  head: ({ loaderData }) => ({
    meta: loaderData
      ? sharingMeta(loaderData.title, loaderData.description, `${loaderData.title} · Blog · Tadoku`)
      : [{ title: 'Blog · Tadoku' }],
  }),
  component: function BlogPostRoute() {
    const query = useQuery(postQuery(Route.useParams().slug))
    const post = query.data
    if (!post) return <PendingView />
    return post.kind === 'found' ? <BlogPostArticle post={post.data} /> : <ContentUnavailable query={query} />
  },
})
