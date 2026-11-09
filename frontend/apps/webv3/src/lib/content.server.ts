import {
  contentAnnouncementListActive,
  contentPageFindBySlug,
  contentPostFindBySlug,
  contentPostList,
  type ContentPost,
} from 'api-client'
import createDOMPurify from 'dompurify'
import { JSDOM } from 'jsdom'
import type { Announcement, BlogPage, BlogPost, CmsPage, Content, SanitizedHtml } from './content'
import { blogPageSize } from './content'
import type { Read } from './leaderboard'
import { publicReadRequest, sessionRequest } from './session.server'

const namespace = 'tadoku'

// Webv2 sanitizes page HTML with DOMPurify's defaults in the browser; Webv3 applies the same call before rendering.
const purify = createDOMPurify(new JSDOM('').window)

/**
 * Content is public, so reads carry no session. A missing or unpublished slug is 404; an ID-shaped slug the API
 * reserves for administrators is refused, which is no published item for a visitor either.
 */
async function find<T, R>(request: Promise<{ data?: T; response?: Response }>, parse: (data: T) => R | undefined): Promise<Content<R>> {
  const answer = await request.catch(() => undefined)
  if (answer?.response && [401, 403, 404].includes(answer.response.status)) return { kind: 'not-found' }

  const data = answer?.data === undefined ? undefined : parse(answer.data)
  return data === undefined ? { kind: 'unavailable' } : { kind: 'found', data }
}

/** Posts need their publish time; request bodies share the schema, so it is optional there. */
function post({ slug, title, content, published_at: publishedAt }: ContentPost): BlogPost | undefined {
  return publishedAt ? { slug, title, content, publishedAt } : undefined
}

export function readPage(slug: string): Promise<Content<CmsPage>> {
  return find(
    contentPageFindBySlug({ ...publicReadRequest('contentPageFindBySlug'), path: { namespace, slug } }),
    page => page.html === undefined ? undefined : { title: page.title, html: purify.sanitize(page.html) as SanitizedHtml },
  )
}

export function readPost(slug: string): Promise<Content<BlogPost>> {
  return find(contentPostFindBySlug({ ...publicReadRequest('contentPostFindBySlug'), path: { namespace, slug } }), post)
}

export async function readPosts(page: number): Promise<Read<BlogPage>> {
  const { data } = await contentPostList({
    ...publicReadRequest('contentPostList'),
    path: { namespace },
    query: { page_size: blogPageSize, page: page - 1 },
  }).catch(() => ({ data: undefined }))
  if (!data) return { ok: false }

  const posts = data.posts.map(post)
  return posts.every((item): item is BlogPost => item !== undefined)
    ? { ok: true, data: { posts, total: data.total_size } }
    : { ok: false }
}

/** The announcements running now; an unavailable list shows none, as in Webv2. */
export async function readAnnouncements(): Promise<Announcement[]> {
  const { data } = await contentAnnouncementListActive({ ...sessionRequest('contentAnnouncementListActive'), path: { namespace } })
    .catch(() => ({ data: undefined }))
  return (data?.announcements ?? []).flatMap(({ id, content, style, href }) =>
    id ? [{ id, content, tone: tones[style], href: href || undefined }] : [])
}

const tones = { info: 'information', success: 'success', warning: 'warning', error: 'danger' } as const
