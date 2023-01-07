import {
  AutocompleteInput,
  AutocompleteMultiInput,
  Input,
  RadioGroup,
  Select,
} from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  ContestRegistrationsView,
  LogConfigurationOptions,
} from '@app/contests/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  AdjustmentsHorizontalIcon,
  LinkIcon,
  UserIcon,
} from '@heroicons/react/20/solid'

export const LogFormSchema = z.object({
  trackingModeSelection: z.enum(['automatic', 'manual', 'personal']),
  contests: z.any(),
  languageCode: z.string(),
  activity: z.number(),
  amount: z.number().positive(),
  unit: z.number(),
  tags: z
    .array(z.string())
    .min(1, 'Must select at least one tag')
    .max(3, 'Must select three or fewer'),
  description: z.string().optional(),
})

export type LogFormSchema = z.infer<typeof LogFormSchema>

interface Props {
  registrations: ContestRegistrationsView
  options: LogConfigurationOptions
}

export const LogForm = ({ registrations, options }: Props) => {
  const defaultValues: Partial<LogFormSchema> = {}

  const methods = useForm({
    resolver: zodResolver(LogFormSchema),
    defaultValues,
  })

  const router = useRouter()
  // const createContestMutation = useCreateContest(id =>
  //   router.replace(routes.contestLeaderboard(id)),
  // )

  // const [createContest] = useDebounce(createContestMutation.mutate, 2500)

  const onSubmit = (data: any) => {
    console.log(data)
    // createContest(data)
  }

  const tags = options.tags
  const activities = options.activities
  const units = options.units
  const languages = options.languages

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
        className="v-stack spaced max-w-screen-md"
      >
        <div className="card v-stack spaced">
          <RadioGroup
            options={[
              {
                value: 'automatic',
                label: 'Automatic',
                description: 'Submit log to all eligible contests',
                IconComponent: LinkIcon,
              },
              {
                value: 'manual',
                label: 'Manual',
                description: 'Choose which contests to submit to',
                IconComponent: AdjustmentsHorizontalIcon,
              },
              {
                value: 'personal',
                label: 'Personal',
                description: 'Do not submit to any contests',
                IconComponent: UserIcon,
              },
            ]}
            label="Contests"
            name="trackingModeSelection"
            defaultValue="automatic"
          />
          <div className="max-w-md v-stack spaced">
            <AutocompleteInput
              name="languageCode"
              label="Language"
              options={languages}
              match={(option, query) =>
                option.name
                  .toLowerCase()
                  .replace(/[^a-zA-Z0-9]/g, '')
                  .includes(query.toLowerCase())
              }
              format={option => option.name}
            />
            <AutocompleteInput
              name="activity"
              label="Activity"
              options={activities}
              match={(option, query) =>
                option.name
                  .toLowerCase()
                  .replace(/[^a-zA-Z0-9]/g, '')
                  .includes(query.toLowerCase())
              }
              format={option => option.name}
            />
            <div className="h-stack spaced">
              <div className="flex-grow">
                <Input
                  name="amount"
                  label="Amount"
                  type="number"
                  defaultValue={0}
                  options={{ valueAsNumber: true }}
                  min={0}
                />
              </div>
              <div className="min-w-[150px]">
                <AutocompleteInput
                  name="unit"
                  label="Unit"
                  options={units}
                  match={(option, query) =>
                    option.name
                      .toLowerCase()
                      .replace(/[^a-zA-Z0-9]/g, '')
                      .includes(query.toLowerCase())
                  }
                  format={option => option.name}
                />
              </div>
            </div>
            <AutocompleteMultiInput
              name="tags"
              label="Tags"
              options={tags}
              match={(option, query) =>
                option.name
                  .toLowerCase()
                  .replace(/[^a-zA-Z0-9]/g, '')
                  .includes(query.toLowerCase())
              }
              getIdForOption={option => option.id}
              format={option => option.name}
            />
            <Input
              name="description"
              label="Description"
              type="text"
              placeholder="e.g. One Piece volume 45"
            />
          </div>
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
