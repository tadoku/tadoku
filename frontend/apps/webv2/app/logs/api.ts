import { z } from 'zod'
import getConfig from 'next/config'
import { useQuery } from 'react-query'
import { Activity, Language } from '@app/contests/api'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion/logs`

export const Unit = z.object({
  id: z.string(),
  log_activity_id: z.number(),
  name: z.string(),
  modifier: z.number(),
  language_code: z.string().nullable().optional(),
})

export type Unit = z.infer<typeof Unit>

export const Tag = z.object({
  id: z.string(),
  log_activity_id: z.number(),
  name: z.string(),
})

export type Tag = z.infer<typeof Tag>

const LogConfigurationOptions = z.object({
  languages: z.array(Language),
  activities: z.array(Activity),
  units: z.array(Unit),
  tags: z.array(Tag),
})

export type LogConfigurationOptions = z.infer<typeof LogConfigurationOptions>

export const useLogConfigurationOptions = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'log', 'configuration-options'],
    async (): Promise<LogConfigurationOptions> => {
      const response = await fetch(`${root}/configuration-options`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return LogConfigurationOptions.parse(await response.json())
    },
    options,
  )
