import {
  contentPageCreate,
  contentPageDelete,
  contentPageFindBySlug,
  contentPageList,
  contentPageUpdate,
  contentPageVersionGet,
  contentPageVersionList,
  contentPostCreate,
  contentPostDelete,
  contentPostFindBySlug,
  contentPostList,
  contentPostUpdate,
  contentPostVersionGet,
  contentPostVersionList,
  type ContentPage,
  type ContentPageVersion,
  type ContentPost,
  type ContentPostVersion,
} from 'api-client'
import { apiData, browserApi, requireFields } from 'ui/api'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ContentConfig, ContentItem, ContentListResponse } from './types'

type ContentInput = {
  id: string
  slug: string
  title: string
  body: string
  published_at?: string | null
}

// The API reads a missing and a null publish time alike, as a draft.
const written = (input: ContentInput) => ({
  slug: input.slug,
  title: input.title,
  ...(input.published_at ? { published_at: input.published_at } : {}),
})

function item(record: ContentPage | ContentPost, body: string | undefined): ContentItem {
  return {
    id: requireFields(record, 'id').id,
    namespace: ('namespace' in record && record.namespace) || '',
    slug: record.slug,
    title: record.title,
    body: body ?? '',
    published_at: record.published_at ?? null,
    created_at: record.created_at ?? '',
    updated_at: record.updated_at ?? '',
  }
}

function versionDetail(record: ContentPageVersion | ContentPostVersion, body: string | undefined): ContentVersionDetail {
  const { id, version, title, created_at } = record
  return { id, version, title, created_at, body: body ?? '' }
}

type Page = { page_size: number; page: number; include_drafts: boolean }

/** The generated operations behind each kind of content, mapped onto the editor's content items. */
const kinds = {
  pages: {
    list: async (namespace: string, query: Page) => {
      const data = await apiData(contentPageList({ ...browserApi(), path: { namespace }, query }))
      return { ...data, items: data.pages.map(page => item(page, page.html)) }
    },
    find: async (namespace: string, slug: string) => {
      const page = await apiData(contentPageFindBySlug({ ...browserApi(), path: { namespace, slug } }))
      return item(page, page.html)
    },
    create: async (namespace: string, input: ContentInput) => {
      const body = { ...written(input), id: input.id, html: input.body }
      const page = await apiData(contentPageCreate({ ...browserApi(), path: { namespace }, body }))
      return item(page, page.html)
    },
    update: async (namespace: string, input: ContentInput) => {
      const body = { ...written(input), html: input.body }
      const page = await apiData(contentPageUpdate({ ...browserApi(), path: { namespace, slug: input.id }, body }))
      return item(page, page.html)
    },
    remove: (namespace: string, id: string) =>
      apiData(contentPageDelete({ ...browserApi(), path: { namespace, slug: id } })),
    versions: async (namespace: string, id: string) =>
      (await apiData(contentPageVersionList({ ...browserApi(), path: { namespace, id } }))).versions,
    version: async (namespace: string, id: string, contentId: string) => {
      const record = await apiData(contentPageVersionGet({ ...browserApi(), path: { namespace, id, contentId } }))
      return versionDetail(record, record.html)
    },
  },
  posts: {
    list: async (namespace: string, query: Page) => {
      const data = await apiData(contentPostList({ ...browserApi(), path: { namespace }, query }))
      return { ...data, items: data.posts.map(post => item(post, post.content)) }
    },
    find: async (namespace: string, slug: string) => {
      const post = await apiData(contentPostFindBySlug({ ...browserApi(), path: { namespace, slug } }))
      return item(post, post.content)
    },
    create: async (namespace: string, input: ContentInput) => {
      const body = { ...written(input), id: input.id, content: input.body }
      const post = await apiData(contentPostCreate({ ...browserApi(), path: { namespace }, body }))
      return item(post, post.content)
    },
    update: async (namespace: string, input: ContentInput) => {
      const body = { ...written(input), content: input.body }
      const post = await apiData(contentPostUpdate({ ...browserApi(), path: { namespace, slug: input.id }, body }))
      return item(post, post.content)
    },
    remove: (namespace: string, id: string) =>
      apiData(contentPostDelete({ ...browserApi(), path: { namespace, slug: id } })),
    versions: async (namespace: string, id: string) =>
      (await apiData(contentPostVersionList({ ...browserApi(), path: { namespace, id } }))).versions,
    version: async (namespace: string, id: string, contentId: string) => {
      const record = await apiData(contentPostVersionGet({ ...browserApi(), path: { namespace, id, contentId } }))
      return versionDetail(record, record.content)
    },
  },
}

export function useContentList(
  config: ContentConfig,
  namespace: string,
  opts: { pageSize: number; page: number },
  options?: { enabled?: boolean },
) {
  return useQuery(
    [config.type, 'list', namespace, opts],
    async (): Promise<ContentListResponse> => {
      const { items, total_size, next_page_token } = await kinds[config.type].list(namespace, {
        page_size: opts.pageSize,
        page: opts.page,
        include_drafts: true,
      })
      return { items, total_size, next_page_token }
    },
    { ...options, retry: false },
  )
}

export function useContentFind(
  config: ContentConfig,
  namespace: string,
  slug: string,
  options?: { enabled?: boolean },
) {
  return useQuery(
    [config.type, 'find', namespace, slug],
    async (): Promise<ContentItem> => kinds[config.type].find(namespace, slug),
    { ...options, retry: false },
  )
}

export function useContentCreate(
  config: ContentConfig,
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (input: ContentInput) => kinds[config.type].create(namespace, input),
    onSuccess,
    onError,
  })
}

export function useContentUpdate(
  config: ContentConfig,
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (input: ContentInput) => kinds[config.type].update(namespace, input),
    onSuccess,
    onError,
  })
}

export function useContentDelete(
  config: ContentConfig,
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (id: string) => {
      await kinds[config.type].remove(namespace, id)
    },
    onSuccess,
    onError,
  })
}

// Version types
export interface ContentVersion {
  id: string
  version: number
  title: string
  created_at: string
}

export interface ContentVersionDetail extends ContentVersion {
  body: string
}

export function useContentVersionList(
  config: ContentConfig,
  namespace: string,
  id: string,
  options?: { enabled?: boolean },
) {
  return useQuery(
    [config.type, 'versions', namespace, id],
    async (): Promise<ContentVersion[]> => kinds[config.type].versions(namespace, id),
    { ...options, retry: false },
  )
}

export function useContentVersionGet(
  config: ContentConfig,
  namespace: string,
  id: string,
  contentId: string | null,
  options?: { enabled?: boolean },
) {
  return useQuery(
    [config.type, 'version', namespace, id, contentId],
    async (): Promise<ContentVersionDetail> => kinds[config.type].version(namespace, id, contentId ?? ''),
    { ...options, enabled: !!contentId && (options?.enabled ?? true), retry: false },
  )
}

// Find by ID: the backend slug endpoint falls back to ID lookup for admins.
export function useContentFindById(
  config: ContentConfig,
  namespace: string,
  id: string,
  options?: { enabled?: boolean },
) {
  return useQuery(
    [config.type, 'findById', namespace, id],
    async (): Promise<ContentItem> => kinds[config.type].find(namespace, id),
    { ...options, retry: false },
  )
}
