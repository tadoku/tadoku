import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()
const root = `${publicRuntimeConfig.apiEndpoint}/content`

const AnnouncementSchema = z.object({
  id: z.string(),
  namespace: z.string().optional().default(''),
  title: z.string(),
  content: z.string(),
  style: z.enum(['success', 'warning', 'error', 'info']),
  href: z.string().nullable().optional().default(null),
  starts_at: z.string(),
  ends_at: z.string(),
  created_at: z.string().optional().default(''),
  updated_at: z.string().optional().default(''),
})

export type Announcement = z.infer<typeof AnnouncementSchema>

const AnnouncementListSchema = z.object({
  announcements: z.array(AnnouncementSchema),
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

export function useAnnouncementList(
  namespace: string,
  opts: { pageSize: number; page: number },
  options?: { enabled?: boolean },
) {
  return useQuery(
    ['announcements', 'list', namespace, opts],
    async () => {
      const params = new URLSearchParams({
        page_size: opts.pageSize.toString(),
        page: opts.page.toString(),
      })
      const response = await fetch(
        `${root}/announcements/${namespace}?${params}`,
        { credentials: 'include' },
      )
      const data = await handleResponse(response)
      return AnnouncementListSchema.parse(data)
    },
    { ...options, retry: false },
  )
}

export function useAnnouncementFind(
  namespace: string,
  id: string,
  options?: { enabled?: boolean },
) {
  return useQuery(
    ['announcements', 'find', namespace, id],
    async (): Promise<Announcement> => {
      const response = await fetch(
        `${root}/announcements/${namespace}/${id}`,
        { credentials: 'include' },
      )
      const data = await handleResponse(response)
      return AnnouncementSchema.parse(data)
    },
    { ...options, retry: false },
  )
}

export function useAnnouncementCreate(
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (input: {
      id: string
      title: string
      content: string
      style: string
      href?: string | null
      starts_at: string
      ends_at: string
    }) => {
      const response = await fetch(
        `${root}/announcements/${namespace}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(input),
          credentials: 'include',
        },
      )
      const data = await handleResponse(response)
      return AnnouncementSchema.parse(data)
    },
    onSuccess,
    onError,
  })
}

export function useAnnouncementUpdate(
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (input: {
      id: string
      title: string
      content: string
      style: string
      href?: string | null
      starts_at: string
      ends_at: string
    }) => {
      const response = await fetch(
        `${root}/announcements/${namespace}/${input.id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(input),
          credentials: 'include',
        },
      )
      const data = await handleResponse(response)
      return AnnouncementSchema.parse(data)
    },
    onSuccess,
    onError,
  })
}

export function useAnnouncementDelete(
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(
        `${root}/announcements/${namespace}/${id}`,
        {
          method: 'DELETE',
          credentials: 'include',
        },
      )
      if (response.status === 204) return
      if (response.status === 403) throw new Error('403')
      if (response.status === 404) throw new Error('404')
      if (!response.ok) throw new Error(response.status.toString())
    },
    onSuccess,
    onError,
  })
}
