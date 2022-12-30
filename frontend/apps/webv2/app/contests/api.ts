import { z } from 'zod'
import getConfig from 'next/config'
import { useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion/contests`

const ContestConfigurationOptions = z.object({
  languages: z.array(
    z.object({
      code: z.string(),
      name: z.string(),
    }),
  ),
  activities: z.array(
    z.object({
      id: z.number(),
      name: z.string(),
      default: z.boolean(),
    }),
  ),
})

export type ContestConfigurationOptions = z.infer<
  typeof ContestConfigurationOptions
>

export const useContestConfigurationOptions = () =>
  useQuery(
    ['contest', 'configuration-options'],
    async (): Promise<ContestConfigurationOptions> => {
      const response = await fetch(`${root}/configuration-options`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestConfigurationOptions.parse(await response.json())
    },
  )
