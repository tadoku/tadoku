import {
  authzRoleUpdate,
  profileUsersList,
  type AuthzRoleUpdateRequest,
  type ProfileUserList,
  type ProfileUserListEntry,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useMutation, useQuery } from '@tanstack/react-query'

// Admin API

export type UserListEntry = ProfileUserListEntry
export type UserList = ProfileUserList

export const useUserList = (
  opts: {
    pageSize: number
    page: number
    query?: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', 'list', opts],
    async (): Promise<UserList> =>
      apiData(
        profileUsersList({
          ...browserApi(),
          query: {
            page_size: opts.pageSize,
            page: opts.page,
            ...(opts.query ? { query: opts.query } : {}),
          },
        }),
      ),
    { ...options, retry: false },
  )

export const useUpdateUserRole = (
  onSuccess: () => void,
  onError: () => void,
) =>
  useMutation({
    mutationFn: async ({
      userId,
      role,
      reason,
    }: AuthzRoleUpdateRequest & { userId: string }) => {
      await apiData(
        authzRoleUpdate({
          ...browserApi(),
          path: { id: userId },
          body: { role, reason },
        }),
      )
    },
    onSuccess,
    onError,
  })
