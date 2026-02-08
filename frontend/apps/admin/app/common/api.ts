import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const immersionRoot = `${publicRuntimeConfig.apiEndpoint}/immersion`
const authzRoot = `${publicRuntimeConfig.apiEndpoint}/authz`

// Admin API

const UserListEntry = z.object({
  id: z.string(),
  display_name: z.string(),
  email: z.string(),
  created_at: z.string(),
  role: z.string().optional(),
})

export type UserListEntry = z.infer<typeof UserListEntry>

const UserList = z.object({
  users: z.array(UserListEntry),
  total_size: z.number(),
})

export type UserList = z.infer<typeof UserList>

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
    async (): Promise<UserList> => {
      const params: Record<string, string> = {
        page_size: opts.pageSize.toString(),
        page: opts.page.toString(),
      }
      if (opts.query) {
        params.query = opts.query
      }
      const response = await fetch(
        `${immersionRoot}/users?${new URLSearchParams(params)}`,
        { credentials: 'include' },
      )

      if (response.status === 401) {
        throw new Error('401')
      }
      if (response.status === 403) {
        throw new Error('403')
      }
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserList.parse(await response.json())
    },
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
    }: {
      userId: string
      role: 'user' | 'banned'
      reason: string
    }) => {
      const response = await fetch(`${authzRoot}/users/${userId}/role`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role, reason }),
        credentials: 'include',
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
    },
    onSuccess,
    onError,
  })
