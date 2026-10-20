import {
  immersionLanguageCreate,
  immersionLanguageList,
  immersionLanguageUpdate,
  type ImmersionLanguage,
  type ImmersionLanguages,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useMutation, useQuery } from '@tanstack/react-query'

export type Language = ImmersionLanguage
export type LanguageList = ImmersionLanguages

export const useLanguageList = (options?: { enabled?: boolean }) =>
  useQuery(
    ['languages', 'list'],
    async (): Promise<LanguageList> =>
      apiData(immersionLanguageList(browserApi())),
    { ...options, retry: false },
  )

export const useLanguageCreate = (
  onSuccess: () => void,
  onError: (error: Error) => void,
) =>
  useMutation({
    mutationFn: async ({ code, name }: { code: string; name: string }) => {
      const result = await immersionLanguageCreate({
        ...browserApi(),
        body: { code, name },
      })
      if (result.response?.status === 409) {
        throw new Error('Language with this code already exists')
      }
      await apiData(result)
    },
    onSuccess,
    onError,
  })

export const useLanguageUpdate = (
  onSuccess: () => void,
  onError: (error: Error) => void,
) =>
  useMutation({
    mutationFn: async ({ code, name }: { code: string; name: string }) => {
      await apiData(
        immersionLanguageUpdate({
          ...browserApi(),
          path: { code },
          body: { name },
        }),
      )
    },
    onSuccess,
    onError,
  })
