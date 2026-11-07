import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import { z } from 'zod'
import type { FlashVariant, ProseNode } from 'paper-ui'
import type { Read } from './leaderboard'

/** One content read: the parsed item, `not-found` when the API has no published item there, or `unavailable`. */
export type Content<T> = { kind: 'found'; data: T } | { kind: 'not-found' } | { kind: 'unavailable' }

/** A CMS page's sanitized HTML, parsed on the server into elements the page renders with Paper's prose styles. */
export interface CmsPage { title: string; body: ProseNode[] }

export interface BlogPost { slug: string; title: string; content: string; publishedAt: string }

export interface BlogPage { posts: BlogPost[]; total: number }

/** A site-wide notice from the CMS. `content` is Markdown; `href` is where to learn more, if anywhere. */
export interface Announcement { id: string; content: string; tone: FlashVariant; href?: string }

export const blogPageSize = 10

export const blogHref = {
  list: (page = 1) => `/blog/${page}`,
  post: (slug: string) => `/blog/posts/${encodeURIComponent(slug)}`,
}

const slug = z.string().min(1)

const getPage = createServerFn({ method: 'GET' })
  .inputValidator(slug)
  .handler(async ({ data }) => {
    const { readPage } = await import('./content.server')
    return readPage(data)
  })

const getPost = createServerFn({ method: 'GET' })
  .inputValidator(slug)
  .handler(async ({ data }) => {
    const { readPost } = await import('./content.server')
    return readPost(data)
  })

const getPosts = createServerFn({ method: 'GET' })
  .inputValidator(z.int().min(1))
  .handler(async ({ data }) => {
    const { readPosts } = await import('./content.server')
    return readPosts(data)
  })

export const pageQuery = (slug: string) => queryOptions({
  queryKey: ['content-page', slug],
  queryFn: (): Promise<Content<CmsPage>> => getPage({ data: slug }),
  staleTime: 30_000,
  retry: false,
})

export const postQuery = (slug: string) => queryOptions({
  queryKey: ['content-post', slug],
  queryFn: (): Promise<Content<BlogPost>> => getPost({ data: slug }),
  staleTime: 30_000,
  retry: false,
})

/** One page of published posts, newest first, counted from 1. */
export const postsQuery = (page: number) => queryOptions({
  queryKey: ['content-posts', page],
  queryFn: (): Promise<Read<BlogPage>> => getPosts({ data: page }),
  staleTime: 30_000,
  retry: false,
})

const getAnnouncements = createServerFn({ method: 'GET' }).handler(async () => {
  const { readAnnouncements } = await import('./content.server')
  return readAnnouncements()
})

/** Read in the browser after hydration, like Webv2, so a slow content API never delays a page. */
export const announcementsQuery = () => queryOptions({
  queryKey: ['content-announcements'],
  queryFn: (): Promise<Announcement[]> => getAnnouncements(),
  staleTime: 60_000,
  retry: false,
})
