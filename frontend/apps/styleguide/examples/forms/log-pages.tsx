import {
  AutocompleteMultiInput,
  Input,
  Select,
} from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { AmountWithUnit, RadioGroup } from 'ui/components/Form'
import {
  AdjustmentsHorizontalIcon,
  LinkIcon,
  UserIcon,
} from '@heroicons/react/20/solid'

const LogPagesFormSchema = z.object({
  trackingModeSelection: z.enum(['automatic', 'manual', 'personal']),
  languageCode: z.string(),
  activity: z.number(),
  amountValue: z.number().positive(),
  amountUnit: z.number(),
  tags: z
    .array(z.string())
    .min(1, 'Must select at least one tag')
    .max(3, 'Must select three or fewer'),
  description: z.string().optional(),
})

export default function LogPagesForm() {
  const methods = useForm({
    resolver: zodResolver(LogPagesFormSchema),
  })
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  const languages = [
    { value: 'jpa', label: 'Japanese' },
    { value: 'zho', label: 'Chinese' },
    { value: 'kor', label: 'Korean' },
  ]
  const units = [
    { value: '1', label: 'Pages' },
    { value: '2', label: 'Comic pages' },
  ]
  const tags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric']
  const activities = [
    { value: '1', label: 'Reading' },
    { value: '2', label: 'Listening' },
    { value: '3', label: 'Speaking' },
    { value: '4', label: 'Writing' },
  ]

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
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
          label="Where to track?"
          name="trackingModeSelection"
          defaultValue="automatic"
        />
        <Select name="languageCode" label="Language" values={languages} />
        <Select
          name="activity"
          label="Activity"
          values={activities}
          options={{ valueAsNumber: true }}
        />
        <div className="h-stack overflow-visible">
          <AmountWithUnit
            label="Amount"
            name="amount"
            units={units}
            unitsLabel="Unit"
            defaultValue={0}
            min={-1}
          />
        </div>
        <Input
          name="description"
          label="Description"
          type="text"
          placeholder="e.g. One Piece volume 45"
        />
        <AutocompleteMultiInput
          name="tags"
          label="Tags"
          options={tags}
          match={(option, query) =>
            option
              .toLowerCase()
              .replace(/[^a-zA-Z0-9]/g, '')
              .includes(query.toLowerCase())
          }
          getIdForOption={option => option}
          format={option => option}
        />
        <button
          type="submit"
          className="btn primary"
          disabled={methods.formState.isSubmitting}
        >
          Save changes
        </button>
      </form>
    </FormProvider>
  )
}
