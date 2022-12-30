import { AutocompleteMultiInput, Input, Checkbox } from 'ui'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { date } from '@app/common/regex'
import { ContestConfigurationOptions } from './api'

const ContestFormSchema = z.object({
  contestStart: z.string().regex(date),
  contestEnd: z.string().regex(date),
  registrationStart: z.string().regex(date),
  registrationEnd: z.string().regex(date),
  description: z.string(),
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

interface Props {
  configurationOptions: ContestConfigurationOptions
}

export const ContestForm = ({
  configurationOptions: { languages, activities },
}: Props) => {
  const { register, handleSubmit, formState, control, watch } = useForm({
    resolver: zodResolver(ContestFormSchema),
  })
  const isOfficial = watch('official')
  const isPrivate = watch('private')

  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form
      onSubmit={handleSubmit(onSubmit, errors => console.log(errors))}
      className="v-stack spaced max-w-screen-sm"
    >
      <div className="card">
        <Input
          name="description"
          label="Contest name"
          register={register}
          formState={formState}
          type="text"
        />
      </div>
      <div className="card">
        <div className="v-stack spaced">
          <h2 className="subtitle">Schedule</h2>
          <div className="v-stack md:h-stack fill spaced">
            <div>
              <Input
                name="contestStart"
                label="Start date"
                register={register}
                formState={formState}
                type="date"
              />
            </div>
            <div>
              <Input
                name="contestEnd"
                label="End date"
                register={register}
                formState={formState}
                type="date"
              />
            </div>
          </div>
          <div className="v-stack md:h-stack fill spaced">
            <div>
              <Input
                name="registrationStart"
                label="Sign up start date"
                register={register}
                formState={formState}
                type="date"
              />
            </div>
            <div>
              <Input
                name="registrationEnd"
                label="Sign up deadline"
                register={register}
                formState={formState}
                type="date"
              />
            </div>
          </div>
        </div>
      </div>
      <div className="card v-stack spaced">
        <h2 className="subtitle">Configuration</h2>
        <AutocompleteMultiInput
          name="languageCodeAllowList"
          label="Allowed languages"
          hint="Leave empty to accept all languages"
          options={languages}
          control={control}
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
          control={control}
          match={(option, query) =>
            option.name
              .toLowerCase()
              .replace(/[^a-zA-Z0-9]/g, '')
              .includes(query.toLowerCase())
          }
          getIdForOption={option => option.id}
          format={option => option.name}
        />
        {!isPrivate && (
          <Checkbox
            name="official"
            label="Official round"
            register={register}
            formState={formState}
          />
        )}
        {!isOfficial && (
          <Checkbox
            name="private"
            label="Only those with link can sign up"
            register={register}
            formState={formState}
          />
        )}
      </div>
      <div className="h-stack spaced justify-end">
        <a href="/contests" className="btn ghost">
          Cancel
        </a>
        <button
          type="submit"
          className="btn primary"
          disabled={formState.isSubmitting}
        >
          Create
        </button>
      </div>
    </form>
  )
}
