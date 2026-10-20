import {
  contentAnnouncementCreate,
  contentAnnouncementDelete,
  contentAnnouncementFindById,
  contentAnnouncementList,
  contentAnnouncementUpdate,
  type ContentAnnouncement,
} from 'api-client'
import { apiData, browserApi, requireFields, type WithRequired } from 'ui/api'
import { useMutation, useQuery } from '@tanstack/react-query'

export type Announcement = WithRequired<ContentAnnouncement, 'id'>

type AnnouncementInput = Announcement

export function useAnnouncementList(
  namespace: string,
  opts: { pageSize: number; page: number },
  options?: { enabled?: boolean },
) {
  return useQuery(
    ['announcements', 'list', namespace, opts],
    async () => {
      const data = await apiData(
        contentAnnouncementList({
          ...browserApi(),
          path: { namespace },
          query: { page_size: opts.pageSize, page: opts.page },
        }),
      )
      return {
        ...data,
        announcements: data.announcements.map(announcement =>
          requireFields(announcement, 'id'),
        ),
      }
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
    async (): Promise<Announcement> =>
      requireFields(
        await apiData(
          contentAnnouncementFindById({
            ...browserApi(),
            path: { namespace, id },
          }),
        ),
        'id',
      ),
    { ...options, retry: false },
  )
}

export function useAnnouncementCreate(
  namespace: string,
  onSuccess: () => void,
  onError: (error: Error) => void,
) {
  return useMutation({
    mutationFn: async (input: AnnouncementInput) =>
      apiData(
        contentAnnouncementCreate({
          ...browserApi(),
          path: { namespace },
          body: input,
        }),
      ),
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
    mutationFn: async (input: AnnouncementInput) =>
      apiData(
        contentAnnouncementUpdate({
          ...browserApi(),
          path: { namespace, id: input.id },
          body: input,
        }),
      ),
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
      await apiData(
        contentAnnouncementDelete({
          ...browserApi(),
          path: { namespace, id },
        }),
      )
    },
    onSuccess,
    onError,
  })
}
