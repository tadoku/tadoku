import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { ContentConfig, ContentItem, ContentListResponse } from './types'

const { publicRuntimeConfig } = getConfig()
const root = `${publicRuntimeConfig.apiEndpoint}/content`

// Zod schemas for validation
const ContentItemSchema = z.object({
  id: z.string(),
  namespace: z.string().optional().default(''),
  slug: z.string(),
  title: z.string(),
  content: z.string().optional(),
  html: z.string().optional(),
  published_at: z.string().nullable().optional().default(null),
  created_at: z.string().optional().default(''),
  updated_at: z.string().optional().default(''),
})

const ContentListSchema = z.object({
  total_size: z.number(),
  next_page_token: z.string(),
})

async function handleResponse(response: Response) {
  if (response.status === 401) throw new Error('401')
  if (response.status === 403) throw new Error('403')
  if (response.status === 404) throw new Error('404')
  if (!response.ok) throw new Error(response.status.toString())
  return response.json()
}

function parseItem(data: z.infer<typeof ContentItemSchema>, bodyField: 'content' | 'html'): ContentItem {
  return {
    id: data.id,
    namespace: data.namespace ?? '',
    slug: data.slug,
    title: data.title,
    body: (bodyField === 'content' ? data.content : data.html) ?? '',
    published_at: data.published_at ?? null,
    created_at: data.created_at ?? '',
    updated_at: data.updated_at ?? '',
  }
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
      const params = new URLSearchParams({
        page_size: opts.pageSize.toString(),
        page: opts.page.toString(),
        include_drafts: 'true',
      })
      const response = await fetch(
        `${root}/${config.type}/${namespace}?${params}`,
        { credentials: 'include' },
      )
      const data = await handleResponse(response)
      const list = ContentListSchema.parse(data)
      const items = z.array(ContentItemSchema).parse(data[config.type])

      return {
        items: items.map(item => parseItem(item, config.bodyField)),
        total_size: list.total_size,
        next_page_token: list.next_page_token,
      }
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
    async (): Promise<ContentItem> => {
      const response = await fetch(
        `${root}/${config.type}/${namespace}/${slug}`,
        { credentials: 'include' },
      )
      const data = await handleResponse(response)
      const item = ContentItemSchema.parse(data)
      return parseItem(item, config.bodyField)
    },
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
    mutationFn: async (input: {
      id: string
      slug: string
      title: string
      body: string
      published_at?: string | null
    }) => {
      const payload: Record<string, unknown> = {
        id: input.id,
        slug: input.slug,
        title: input.title,
        [config.bodyField]: input.body,
      }
      if (input.published_at) {
        payload.published_at = input.published_at
      }

      const response = await fetch(
        `${root}/${config.type}/${namespace}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          credentials: 'include',
        },
      )
      const data = await handleResponse(response)
      const item = ContentItemSchema.parse(data)
      return parseItem(item, config.bodyField)
    },
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
    mutationFn: async (input: {
      id: string
      slug: string
      title: string
      body: string
      published_at?: string | null
    }) => {
      const payload: Record<string, unknown> = {
        slug: input.slug,
        title: input.title,
        [config.bodyField]: input.body,
      }
      if (input.published_at) {
        payload.published_at = input.published_at
      }

      const response = await fetch(
        `${root}/${config.type}/${namespace}/${input.id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          credentials: 'include',
        },
      )
      const data = await handleResponse(response)
      const item = ContentItemSchema.parse(data)
      return parseItem(item, config.bodyField)
    },
    onSuccess,
    onError,
  })
}

// Find by ID: checks cached list data first, falls back to fetching the list.
// The backend only has a find-by-slug GET endpoint, but the list endpoint returns
// full item data so we can resolve by ID from list results.
export function useContentFindById(
  config: ContentConfig,
  namespace: string,
  id: string,
  options?: { enabled?: boolean },
) {
  const queryClient = useQueryClient()

  return useQuery(
    [config.type, 'findById', namespace, id],
    async (): Promise<ContentItem> => {
      // Check cached list queries for this item
      const cachedQueries = queryClient.getQueriesData<ContentListResponse>(
        [config.type, 'list'],
      )
      for (const [, data] of cachedQueries) {
        if (data) {
          const item = data.items.find(i => i.id === id)
          if (item) return item
        }
      }

      // Not in cache - fetch the list to find the item
      const params = new URLSearchParams({
        page_size: '100',
        page: '0',
        include_drafts: 'true',
      })
      const response = await fetch(
        `${root}/${config.type}/${namespace}?${params}`,
        { credentials: 'include' },
      )
      const data = await handleResponse(response)
      const items = z.array(ContentItemSchema).parse(data[config.type])
      const parsed = items.map(item => parseItem(item, config.bodyField))
      const found = parsed.find(i => i.id === id)
      if (!found) throw new Error('404')
      return found
    },
    { ...options, retry: false },
  )
}
