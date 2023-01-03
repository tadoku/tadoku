import { AutocompleteMultiInput, Input, Checkbox } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { date } from '@app/common/regex'
import { ContestConfigurationOptions, useCreateContest } from './api'
import { useRouter } from 'next/router'
import { useDebounce } from 'use-debounce'
import { routes } from '@app/common/routes'

export const ContestFormSchema = z
  .object({
    contestStart: z.string().regex(date),
    contestEnd: z.string().regex(date),
    registrationEnd: z.string().regex(date),
    description: z
      .string()
      .min(3, 'Contest name should be at least 3 characters'),
    private: z.boolean(),
    official: z.boolean().optional().default(false),
    languageCodeAllowList: z.array(
      z.object({
        code: z.string(),
        name: z.string(),
      }),
    ),
    activityTypeIdAllowList: z
      .array(
        z.object({
          id: z.number(),
          name: z.string(),
        }),
      )
      .min(1, 'Need to select at least one activity'),
  })
  .refine(
    ({ official, languageCodeAllowList: list }) =>
      !(official && list.length > 0),
    {
      path: ['languageCodeAllowList'],
      message: 'Official contests cannot limit language selection',
    },
  )
  .transform(contest => {
    const {
      contestStart: contest_start,
      contestEnd: contest_end,
      registrationEnd: registration_end,
      ...rest
    } = contest
    return {
      ...rest,
      contest_start,
      contest_end,
      registration_end,
      language_code_allow_list: contest.languageCodeAllowList.map(l => l.code),
      activity_type_id_allow_list: contest.activityTypeIdAllowList.map(
        a => a.id,
      ),
    }
  })

export type ContestFormSchema = z.infer<typeof ContestFormSchema>

interface Props {
  configurationOptions: ContestConfigurationOptions
}

export const ContestForm = ({
  configurationOptions: { languages, activities, canCreateOfficialRound },
}: Props) => {
  const defaultValues: Partial<ContestFormSchema> = {
    activityTypeIdAllowList: activities.filter(a => a.default),
  }

  const methods = useForm({
    resolver: zodResolver(ContestFormSchema),
    defaultValues,
  })
  const isOfficial = methods.watch('official')
  const isPrivate = methods.watch('private')

  const router = useRouter()
  const createContestMutation = useCreateContest(id =>
    router.replace(routes.contestLeaderboard(id)),
  )

  const [createContest] = useDebounce(createContestMutation.mutate, 2500)

  const onSubmit = (data: any) => {
    createContest(data)
  }

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
        className="v-stack spaced max-w-screen-sm"
      >
        <div className="card">
          <Input name="description" label="Contest name" type="text" />
        </div>
        <div className="card">
          <div className="v-stack spaced">
            <h2 className="subtitle">Schedule</h2>
            <div className="v-stack md:h-stack fill spaced">
              <div>
                <Input name="contestStart" label="Start date" type="date" />
              </div>
              <div>
                <Input name="contestEnd" label="End date" type="date" />
              </div>
            </div>
            <Input
              name="registrationEnd"
              label="Sign up deadline"
              type="date"
            />
          </div>
        </div>
        <div className="card v-stack spaced">
          <h2 className="subtitle">Configuration</h2>
          <AutocompleteMultiInput
            name="languageCodeAllowList"
            label="Allowed languages"
            hint="Leave empty to accept all languages"
            options={languages}
            match={(option, query) =>
              option.name
                .toLowerCase()
                .replace(/[^a-zA-Z0-9]/g, '')
                .includes(query.toLowerCase())
            }
            getIdForOption={option => option.code}
            format={option => option.name}
          />
          <AutocompleteMultiInput
            name="activityTypeIdAllowList"
            label="Allowed activities"
            options={activities}
            match={(option, query) =>
              option.name
                .toLowerCase()
                .replace(/[^a-zA-Z0-9]/g, '')
                .includes(query.toLowerCase())
            }
            getIdForOption={option => option.id}
            format={option => option.name}
          />
          {canCreateOfficialRound && !isPrivate && (
            <Checkbox name="official" label="Official round" />
          )}
          {!isOfficial && (
            <Checkbox name="private" label="Only those with link can sign up" />
          )}
        </div>
        <div className="h-stack spaced justify-end">
          <a href={routes.contestListOfficial()} className="btn ghost">
            Cancel
          </a>
          <button
            type="submit"
            className="btn primary"
            disabled={
              methods.formState.isSubmitting || !methods.formState.isValid
            }
          >
            Create
          </button>
        </div>
      </form>
    </FormProvider>
  )
}
