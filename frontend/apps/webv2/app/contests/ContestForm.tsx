import { AutocompleteMultiInput, Input, Checkbox } from 'ui'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const ContestFormSchema = z.object({
  contestStart: z.date(),
  contestEnd: z.date(),
  registrationStart: z.date(),
  registrationEnd: z.date(),
  description: z.string(),
  private: z.boolean(),
  official: z.boolean().optional().default(false),
  languageCodeAllowList: z.array(z.string()),
  activityTypeIdAllowList: z.array(z.number()),
})

interface Props {}

// TODO: fetch from API
const languages = [
  { code: 'jpa', name: 'Japanese' },
  { code: 'kor', name: 'Korean' },
  { code: 'zho', name: 'Chinese' },
  { code: 'ita', name: 'Italian' },
  { code: 'nld', name: 'Dutch' },
  { code: 'eng', name: 'English' },
]
const activities = [
  { id: 1, name: 'Reading' },
  { id: 2, name: 'Listening' },
  { id: 3, name: 'Speaking' },
  { id: 4, name: 'Writing' },
]

export const ContestForm = ({}: Props) => {
  const { register, handleSubmit, formState, control, watch } = useForm({
    resolver: zodResolver(ContestFormSchema),
  })
  const isOfficial = watch('official')
  const isPrivate = watch('private')

  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form
      onSubmit={handleSubmit(onSubmit, errors => console.log(errors))}
      className="v-stack spaced"
    >
      <div className="h-stack spaced fill">
        <div className="card">
          <Input
            name="description"
            label="Contest name"
            register={register}
            formState={formState}
            type="text"
          />
          <div className="v-stack spaced mt-8">
            <h2 className="subtitle">Schedule</h2>
            <div className="h-stack fill spaced">
              <Input
                name="contestStart"
                label="Start date"
                register={register}
                formState={formState}
                type="date"
              />
              <Input
                name="contestEnd"
                label="End date"
                register={register}
                formState={formState}
                type="date"
                options={{
                  valueAsDate: true,
                }}
              />
            </div>
            <div className="h-stack fill spaced">
              <Input
                name="registrationStart"
                label="Sign up start date"
                register={register}
                formState={formState}
                type="date"
                options={{
                  valueAsDate: true,
                }}
              />
              <Input
                name="registrationEnd"
                label="Sign up deadline"
                register={register}
                formState={formState}
                type="date"
                options={{
                  required: true,
                  valueAsDate: true,
                }}
              />
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
