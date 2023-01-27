import { AutocompleteMultiInput, Flash } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  ContestRegistrationView,
  ContestView,
  useContestConfigurationOptions,
  useContestRegistrationUpdate,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { useDebounce } from 'use-debounce'
import { routes } from '@app/common/routes'
import { ExclamationCircleIcon } from '@heroicons/react/20/solid'

export const ContestRegistrationFormSchema = z
  .object({
    contest_id: z.string(),
    new_languages: z
      .array(
        z.object({
          code: z.string(),
          name: z.string(),
        }),
      )
      .max(3, 'Cannot select more than 3 languages')
      .min(1, 'Must select at least 1 languages'),
    languages: z
      .array(
        z.object({
          code: z.string(),
          name: z.string(),
        }),
      )
      .optional(),
  })
  .refine(
    registration => {
      if (!registration.languages) {
        return true
      }

      // Check if new languages include all the old ones, fail if missing
      const previousLangs = registration.languages.map(it => it.code)
      const newLangs = new Set(registration.new_languages.map(it => it.code))

      for (const lang of previousLangs) {
        if (!newLangs.has(lang)) {
          return false
        }
      }

      return true
    },
    {
      path: ['new_languages'],
      message: "Cannot remove languages you've registered previously.",
    },
  )

export type ContestRegistrationFormSchema = z.infer<
  typeof ContestRegistrationFormSchema
>

interface Props {
  contest: ContestView
  isClosed: boolean
  data?: ContestRegistrationView
}

export const ContestRegistrationForm = ({ contest, data, isClosed }: Props) => {
  const allLanguages = contest.allowed_languages === null
  const options = useContestConfigurationOptions({ enabled: allLanguages })

  const defaultValues = {
    contest_id: contest.id ?? '',
    languages: data?.languages,
    new_languages: data?.languages ?? [],
  }

  const methods = useForm<ContestRegistrationFormSchema>({
    resolver: zodResolver(ContestRegistrationFormSchema),
    defaultValues,
  })

  const router = useRouter()

  const createRegistrationMutation = useContestRegistrationUpdate(() =>
    router.replace(routes.contestLeaderboard(contest.id!)),
  )

  const [createContest] = useDebounce(createRegistrationMutation.mutate, 2500)

  const onSubmit = (data: ContestRegistrationFormSchema) => {
    createContest(data)
  }

  if (!contest.id) {
    return (
      <Flash style="error" IconComponent={ExclamationCircleIcon}>
        Could not load contest.
      </Flash>
    )
  }

  if (allLanguages && options.isError) {
    return (
      <Flash style="error" IconComponent={ExclamationCircleIcon}>
        Could not load contest configuration, please try again later.
      </Flash>
    )
  }

  if (allLanguages && !options.data) {
    return <>Loading...</>
  }

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
      >
        <fieldset
          disabled={isClosed}
          className="v-stack spaced max-w-screen-sm"
        >
          <div className="card v-stack spaced">
            <h2 className="subtitle">Contest registration: {contest.title}</h2>
            <AutocompleteMultiInput
              name="new_languages"
              label="Languages"
              hint="Select up to 3 languages"
              options={
                (contest.allowed_languages || options.data?.languages) ?? []
              }
              match={(option, query) =>
                option.name
                  .toLowerCase()
                  .replace(/[^a-zA-Z0-9]/g, '')
                  .includes(query.toLowerCase())
              }
              getIdForOption={option => option.code}
              format={option => option.name}
            />
          </div>
          <div className="h-stack spaced justify-end">
            <a
              href={routes.contestLeaderboard(contest.id!)}
              className="btn ghost"
            >
              Cancel
            </a>
            <button
              type="submit"
              className="btn primary"
              disabled={methods.formState.isSubmitting}
            >
              {defaultValues ? 'Register' : 'Update registration'}
            </button>
          </div>
        </fieldset>
      </form>
    </FormProvider>
  )
}
