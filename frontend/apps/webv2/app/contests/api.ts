import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery } from 'react-query'
import { ContestFormSchema } from '@app/contests/ContestForm'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion/contests`

const ContestConfigurationOptions = z
  .object({
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
    can_create_official_round: z.boolean(),
  })
  .transform(data => {
    const { can_create_official_round: canCreateOfficialRound, ...rest } = data

    return {
      ...rest,
      canCreateOfficialRound,
    }
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

export const useCreateContest = (onSuccess: (id: string) => void) =>
  useMutation({
    mutationFn: async (contest: ContestFormSchema) => {
      const res = await fetch(root, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(contest),
      })
      return await res.json()
    },
    onSuccess(data) {
      onSuccess(data.id)
    },
  })
