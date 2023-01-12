import { AutocompleteMultiInput, Input, Checkbox, TextArea } from 'ui'
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
    contest_start: z.string().regex(date),
    contest_end: z.string().regex(date),
    registration_end: z.string().regex(date),
    title: z.string().min(3, 'Title should be at least 3 characters'),
    description: z
      .string()
      .max(500, 'Description should be 500 characters or fewer')
      .optional(),
    private: z.boolean(),
    official: z.boolean().optional().default(false),
    language_code_allow_list: z.array(
      z.object({
        code: z.string(),
        name: z.string(),
      }),
    ),
    activity_type_id_allow_list: z
      .array(
        z.object({
          id: z.number(),
          name: z.string(),
        }),
      )
      .min(1, 'Need to select at least one activity'),
  })
  .refine(
    ({ official, language_code_allow_list: list }) =>
      !(official && list.length > 0),
    {
      path: ['language_code_allow_list'],
      message: 'Official contests cannot limit language selection',
    },
  )

export type ContestFormSchema = z.infer<typeof ContestFormSchema>

interface Props {
  configurationOptions: ContestConfigurationOptions
}

export const ContestForm = ({
  configurationOptions: {
    languages,
    activities,
    can_create_official_round: canCreateOfficialRound,
  },
}: Props) => {
  const defaultValues: Partial<ContestFormSchema> = {
    activity_type_id_allow_list: activities.filter(a => a.default),
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
          <div className="v-stack spaced">
            <h2 className="subtitle">Contest</h2>
            <Input name="title" label="Title" type="text" />
            <TextArea
              name="description"
              label="Description"
              placeholder="Explain any extra rules or details about this contest, this will be visible on the contest leaderboard."
              hint="Optional"
            />
          </div>
        </div>
        <div className="card">
          <div className="v-stack spaced">
            <h2 className="subtitle">Schedule</h2>
            <div className="v-stack md:h-stack fill spaced">
              <div>
                <Input name="contest_start" label="Start date" type="date" />
              </div>
              <div>
                <Input name="contest_end" label="End date" type="date" />
              </div>
            </div>
            <Input
              name="registration_end"
              label="Sign up deadline"
              type="date"
            />
          </div>
        </div>
        <div className="card v-stack spaced">
          <h2 className="subtitle">Configuration</h2>
          <AutocompleteMultiInput
            name="language_code_allow_list"
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
            name="activity_type_id_allow_list"
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
            disabled={methods.formState.isSubmitting}
          >
            Create
          </button>
        </div>
      </form>
    </FormProvider>
  )
}
