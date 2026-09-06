import { Button, Input } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function InputRecommendedFixture() {
  const [saved, setSaved] = useState('')
  const methods = useForm<{ title: string; date: string }>({
    defaultValues: { title: 'August reading log', date: '2026-08-31' },
  })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        noValidate
        onSubmit={methods.handleSubmit(({ title, date }) =>
          setSaved(`Saved: ${title} · ${date}`),
        )}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Input
            name="title"
            label="Log title"
            rules={{ required: 'Enter a log title.' }}
            required
          />
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
        <p role="status">{saved}</p>
      </form>
    </FormProvider>
  )
}
