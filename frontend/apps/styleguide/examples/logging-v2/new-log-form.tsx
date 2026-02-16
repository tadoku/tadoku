import { useMemo, useState } from 'react'
import { DateTime } from 'luxon'
import { Input, Modal, Select, TagsInput } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'
import { AmountWithUnit } from 'ui/components/Form'

export default function NewLogForm() {
  const [backdateOpen, setBackdateOpen] = useState(false)
  const { minDate, maxDate } = useMemo(() => {
    const today = DateTime.local()
    return {
      minDate: today.minus({ years: 1 }).toFormat('yyyy-MM-dd'),
      maxDate: today.toFormat('yyyy-MM-dd'),
    }
  }, [])
  const methods = useForm({
    defaultValues: {
      languageCode: 'jpa',
      activity: '1',
      amountValue: 230,
      amountUnit: '1',
      description: '無職転生 5巻',
      tags: ['Book', 'Active pick'],
      date: maxDate,
    },
  })
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  const allTags = [
    'Book',
    'Ebook',
    'Fiction',
    'Non-fiction',
    'Manga',
    'Active pick',
  ]

  const fetchTags = async (input: string): Promise<string[]> => {
    await new Promise(resolve => setTimeout(resolve, 150))
    return allTags.filter(tag =>
      tag.toLowerCase().includes(input.toLowerCase()),
    )
  }

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        <div className="v-stack spaced">
          <Select
            name="languageCode"
            label="Language"
            values={[
              { value: 'jpa', label: 'Japanese' },
              { value: 'zho', label: 'Chinese' },
              { value: 'kor', label: 'Korean' },
            ]}
          />
          <Select
            name="activity"
            label="Activity"
            values={[
              { value: '1', label: 'Reading' },
              { value: '2', label: 'Listening' },
              { value: '3', label: 'Speaking' },
              { value: '4', label: 'Writing' },
            ]}
          />
          <AmountWithUnit
            label="Amount"
            name="amount"
            units={[
              { value: '1', label: 'Pages' },
              { value: '2', label: 'Sentences' },
              { value: '3', label: 'Characters' },
            ]}
            unitsLabel="Unit"
            defaultValue={230}
            min={0}
          />
          <Input
            name="description"
            label="Description"
            type="text"
            placeholder="e.g. One Piece volume 45"
          />
          <TagsInput
            name="tags"
            label="Tags"
            placeholder="Type to search or add tags..."
            maxTags={5}
            getSuggestions={fetchTags}
          />
        </div>
        <div className="flex items-center justify-between border-t border-slate-200 mt-4 pt-4">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setBackdateOpen(true)}
          >
            Backdate?
          </button>
          <button
            type="submit"
            className="btn primary"
            disabled={methods.formState.isSubmitting}
          >
            Submit
          </button>
        </div>
      </form>

      <Modal isOpen={backdateOpen} setIsOpen={setBackdateOpen} title="Backdate log">
        <p className="modal-body mb-3">
          You can backdate up to 1 year ago. Backdated logs won't count towards contests that have already ended.
        </p>
        <Input name="date" label="Date" type="date" min={minDate} max={maxDate} />
        <div className="modal-actions justify-end gap-2">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setBackdateOpen(false)}
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn primary"
            onClick={() => setBackdateOpen(false)}
          >
            Confirm
          </button>
        </div>
      </Modal>
    </FormProvider>
  )
}
