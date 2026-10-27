import {
  authzRoleUpdate,
  profileAccountDeletionCreate,
  profileUsersList,
  zProfileErrorResponse,
  type AuthzRoleUpdateRequest,
  type ProfileAccountDeletionRequest,
  type ProfileErrorResponse,
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

export type AccountDeletionRefusal =
  | ProfileErrorResponse['error']
  | 'confirmation_mismatch'
  | 'rejected'
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'unavailable'

export type AccountDeletionOutcome =
  | { kind: 'filed' }
  | { kind: 'refused'; reason: AccountDeletionRefusal }

const refusalByStatus: Record<number, AccountDeletionRefusal> = {
  400: 'rejected',
  401: 'unauthorized',
  403: 'forbidden',
  404: 'not_found',
}

function accountDeletionRefusal(
  status: number | undefined,
  error: unknown,
): AccountDeletionRefusal {
  if (
    status === 400 &&
    typeof error === 'object' &&
    error !== null &&
    'error' in error &&
    error.error === 'confirmation_mismatch'
  ) {
    return 'confirmation_mismatch'
  }
  const coded = zProfileErrorResponse.safeParse(error)
  if (status === 409 && coded.success) return coded.data.error
  return refusalByStatus[status ?? 0] ?? 'unavailable'
}

/** Files an account deletion. Every answer, including a transport failure, resolves to an outcome. */
export const useFileAccountDeletion = () =>
  useMutation({
    mutationFn: async ({
      userId,
      email,
      reason,
    }: ProfileAccountDeletionRequest & {
      userId: string
    }): Promise<AccountDeletionOutcome> => {
      try {
        const { error, response } = await profileAccountDeletionCreate({
          ...browserApi(),
          path: { id: userId },
          body: { email, reason },
        })
        if (response?.ok) return { kind: 'filed' }
        return {
          kind: 'refused',
          reason: accountDeletionRefusal(response?.status, error),
        }
      } catch {
        return { kind: 'refused', reason: 'unavailable' }
      }
    },
  })
