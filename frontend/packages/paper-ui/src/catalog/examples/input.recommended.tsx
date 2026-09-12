import { Button, Input } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function InputRecommendedFixture() {
  const [saved, setSaved] = useState('')
  const methods = useForm<{ title: string; date: string; pages: number }>({
    defaultValues: { title: 'August reading log', date: '2026-08-31', pages: 24 },
  })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack w-full max-w-2xl"
        noValidate
        onSubmit={methods.handleSubmit(({ title, date, pages }) =>
          setSaved(`Saved: ${title} · ${pages} pages · ${date}`),
        )}
      >
        <Input
            name="title"
            label="Log title"
            rules={{ required: 'Enter a log title.' }}
            required
        />
        <div className="grid gap-4 sm:grid-cols-2">
          <Input name="pages" label="Pages read" type="number" min={1}
            rules={{ required: 'Enter the number of pages.', valueAsNumber: true, min: { value: 1, message: 'Enter at least one page.' } }} required />
          <Input
            name="date"
            label="Date"
            type="date"
            hint="The date you read, in UTC. Fields stay aligned when only one has a hint."
            rules={{ required: 'Choose a date.' }}
            required
          />
        </div>
        <div className="paper-cluster">
          <Button type="submit">Save log</Button>
        </div>
        {saved && <p role="status">{saved}</p>}
      </form>
    </FormProvider>
  )
}
