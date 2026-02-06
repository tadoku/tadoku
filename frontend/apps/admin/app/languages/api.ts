import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion`

const Language = z.object({
  code: z.string(),
  name: z.string(),
})

export type Language = z.infer<typeof Language>

const LanguageList = z.object({
  languages: z.array(Language),
})

export type LanguageList = z.infer<typeof LanguageList>

export const useLanguageList = (options?: { enabled?: boolean }) =>
  useQuery(
    ['languages', 'list'],
    async (): Promise<LanguageList> => {
      const response = await fetch(`${root}/languages`, {
        credentials: 'include',
      })

      if (response.status === 401) throw new Error('401')
      if (response.status === 403) throw new Error('403')
      if (response.status !== 200) throw new Error(response.status.toString())

      return LanguageList.parse(await response.json())
    },
    { ...options, retry: false },
  )

export const useLanguageCreate = (
  onSuccess: () => void,
  onError: (error: Error) => void,
) =>
  useMutation({
    mutationFn: async ({ code, name }: { code: string; name: string }) => {
      const response = await fetch(`${root}/languages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code, name }),
        credentials: 'include',
      })
      if (response.status === 409) throw new Error('Language with this code already exists')
      if (response.status !== 200) throw new Error(response.status.toString())
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
      const response = await fetch(`${root}/languages/${code}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
        credentials: 'include',
      })
      if (response.status !== 200) throw new Error(response.status.toString())
    },
    onSuccess,
    onError,
  })
