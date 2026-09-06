import { Button, Input } from 'paper-ui'
import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function InputRecommendedFixture() {
  const [saved, setSaved] = useState('')
  const methods = useForm<{ title: string }>({
    defaultValues: { title: 'August reading log' },
  })
  return (
    <FormProvider {...methods}>
      <form
        className="paper-stack"
        noValidate
        onSubmit={methods.handleSubmit(({ title }) =>
          setSaved(`Saved: ${title}`),
        )}
      >
        <Input
          name="title"
          label="Log title"
          hint="Give this reading session a short, recognizable name."
          rules={{ required: 'Enter a log title.' }}
          required
        />
        <div className="paper-cluster">
          <Button type="submit">Save log</Button>
        </div>
        <p role="status">{saved}</p>
      </form>
    </FormProvider>
  )
}
